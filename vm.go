/*
Copyright (c) 2019 DERO Foundation. All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
may be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package randomx

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/aes"
	"math"
	"runtime"
	"unsafe"
)
import "golang.org/x/crypto/blake2b"

type REG struct {
	Hi uint64
	Lo uint64
}

type VM struct {
	ScratchPad ScratchPad

	Dataset Dataset

	program    ByteCode
	jitProgram VMProgramFunc
}

func NewVM(dataset Dataset) *VM {
	vm := &VM{
		Dataset: dataset,
	}
	if dataset.Cache().HasJIT() {
		vm.jitProgram = mapProgram(nil, int(RandomXCodeSize))
		if dataset.Flags()&RANDOMX_FLAG_SECURE == 0 {
			mapProgramRWX(vm.jitProgram)
		}
	}
	return vm
}

// run calculate hash based on input. Not thread safe.
// Warning: Underlying callers will run float64 SetRoundingMode directly
// It is the caller's responsibility to set and restore the mode to IEEE 754 roundTiesToEven between full executions
// Additionally, runtime.LockOSThread and defer runtime.UnlockOSThread is recommended to prevent other goroutines sharing these changes
func (vm *VM) run(inputHash [64]byte, roundingMode uint8) (reg RegisterFile) {

	reg.FPRC = roundingMode

	// buffer first 128 bytes are entropy below rest are program bytes
	var buffer [16*8 + RANDOMX_PROGRAM_SIZE*8]byte
	aes.FillAes4Rx4(inputHash, buffer[:])

	entropy := (*[16]uint64)(unsafe.Pointer(&buffer))

	// do more initialization before we run

	for i := range entropy[:8] {
		reg.A[i/2][i%2] = SmallPositiveFloatBits(entropy[i])
	}

	// memory registers
	var ma, mx uint32

	ma = uint32(entropy[8] & CacheLineAlignMask)
	mx = uint32(entropy[10])

	addressRegisters := entropy[12]

	var readReg [4]uint64
	for i := range readReg {
		readReg[i] = uint64(i*2) + (addressRegisters & 1)
		addressRegisters >>= 1
	}

	datasetOffset := (entropy[13] % (DatasetExtraItems + 1)) * CacheLineSize

	eMask := [2]uint64{ExponentMask(entropy[14]), ExponentMask(entropy[15])}

	prog := buffer[len(entropy)*8:]
	CompileProgramToByteCode(prog, &vm.program)

	datasetMemory := vm.Dataset.Memory()

	var jitProgram VMProgramFunc

	if vm.jitProgram != nil {
		if datasetMemory == nil {
			if vm.Dataset.Flags()&RANDOMX_FLAG_SECURE > 0 {
				mapProgramRW(vm.jitProgram)
				jitProgram = vm.program.generateCode(vm.jitProgram, nil)
				mapProgramRX(vm.jitProgram)
			} else {
				jitProgram = vm.program.generateCode(vm.jitProgram, nil)
			}
		} else {
			// full mode and we have JIT
			if vm.Dataset.Flags()&RANDOMX_FLAG_SECURE > 0 {
				mapProgramRW(vm.jitProgram)
				jitProgram = vm.program.generateCode(vm.jitProgram, &readReg)
				mapProgramRX(vm.jitProgram)
			} else {
				jitProgram = vm.program.generateCode(vm.jitProgram, &readReg)
			}

			vm.jitProgram.ExecuteFull(&reg, &vm.ScratchPad, &datasetMemory[datasetOffset/CacheLineSize], RANDOMX_PROGRAM_ITERATIONS, ma, mx, eMask)
			return reg
		}
	}

	spAddr0 := uint64(mx)
	spAddr1 := uint64(ma)

	for ic := 0; ic < RANDOMX_PROGRAM_ITERATIONS; ic++ {
		spMix := reg.R[readReg[0]] ^ reg.R[readReg[1]]

		spAddr0 ^= spMix
		spAddr0 &= ScratchpadL3Mask64
		spAddr1 ^= spMix >> 32
		spAddr1 &= ScratchpadL3Mask64

		//TODO: optimize these loads!
		for i := uint64(0); i < RegistersCount; i++ {
			reg.R[i] ^= vm.ScratchPad.Load64(uint32(spAddr0 + 8*i))
		}

		for i := uint64(0); i < RegistersCountFloat; i++ {
			reg.F[i] = vm.ScratchPad.Load32FA(uint32(spAddr1 + 8*i))
		}

		for i := uint64(0); i < RegistersCountFloat; i++ {
			reg.E[i] = vm.ScratchPad.Load32FA(uint32(spAddr1 + 8*(i+RegistersCountFloat)))

			reg.E[i][LOW] = MaskRegisterExponentMantissa(reg.E[i][LOW], eMask[LOW])
			reg.E[i][HIGH] = MaskRegisterExponentMantissa(reg.E[i][HIGH], eMask[HIGH])
		}

		// run the actual bytecode
		if jitProgram != nil {
			// light mode
			jitProgram.Execute(&reg, &vm.ScratchPad, eMask)
		} else {
			vm.program.Execute(&reg, &vm.ScratchPad, eMask)
		}

		mx ^= uint32(reg.R[readReg[2]] ^ reg.R[readReg[3]])
		mx &= uint32(CacheLineAlignMask)

		vm.Dataset.PrefetchDataset(datasetOffset + uint64(mx))
		// execute / load output from diffuser superscalar program to get dataset 64 bytes
		vm.Dataset.ReadDataset(datasetOffset+uint64(ma), &reg.R)

		// swap the elements
		mx, ma = ma, mx

		for i := uint64(0); i < RegistersCount; i++ {
			vm.ScratchPad.Store64(uint32(spAddr1+8*i), reg.R[i])
		}

		for i := uint64(0); i < RegistersCountFloat; i++ {
			reg.F[i][LOW] = Xor(reg.F[i][LOW], reg.E[i][LOW])
			reg.F[i][HIGH] = Xor(reg.F[i][HIGH], reg.E[i][HIGH])

			vm.ScratchPad.Store64(uint32(spAddr0+16*i), math.Float64bits(reg.F[i][LOW]))
			vm.ScratchPad.Store64(uint32(spAddr0+16*i+8), math.Float64bits(reg.F[i][HIGH]))
		}

		spAddr0 = 0
		spAddr1 = 0

	}

	runtime.KeepAlive(buffer)

	return reg

}

func (vm *VM) initScratchpad(seed *[64]byte) {
	vm.ScratchPad.Init(seed)
}

func (vm *VM) runLoops(tempHash [64]byte) RegisterFile {
	if lockThreadDueToRoundingMode {
		// Lock thread due to rounding mode flags
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	roundingMode := uint8(0)

	for chain := 0; chain < RANDOMX_PROGRAM_COUNT-1; chain++ {
		reg := vm.run(tempHash, roundingMode)
		roundingMode = reg.FPRC

		// write R, F, E, A registers
		tempHash = blake2b.Sum512(reg.Memory()[:])
		runtime.KeepAlive(reg)
	}

	// final loop executes here
	reg := vm.run(tempHash, roundingMode)
	// always force a restore
	reg.FPRC = 0xff

	// restore rounding mode to 0
	SetRoundingMode(&reg, 0)

	return reg
}

// CalculateHash Not thread safe.
func (vm *VM) CalculateHash(input []byte, output *[32]byte) {
	tempHash := blake2b.Sum512(input)

	vm.initScratchpad(&tempHash)

	reg := vm.runLoops(tempHash)

	// now hash the scratch pad as it will act as register A
	aes.HashAes1Rx4(vm.ScratchPad[:], &tempHash)

	regMem := reg.Memory()
	// write hash onto register A
	copy(regMem[RegisterFileSize-RegistersCountFloat*2*8:], tempHash[:])

	// write R, F, E, A registers
	*output = blake2b.Sum256(regMem[:])
	runtime.KeepAlive(reg)
}

func (vm *VM) Close() error {
	if vm.jitProgram != nil {
		return vm.jitProgram.Close()
	}
	return nil
}

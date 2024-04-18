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
	"git.gammaspectra.live/P2Pool/go-randomx/v2/aes"
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

	Dataset Randomx_Dataset

	JITProgram VMProgramFunc
}

// Run calculate hash based on input
// Warning: Underlying callers will run float64 SetRoundingMode directly
// It is the caller's responsibility to set and restore the mode to IEEE 754 roundTiesToEven between full executions
// Additionally, runtime.LockOSThread and defer runtime.UnlockOSThread is recommended to prevent other goroutines sharing these changes
func (vm *VM) Run(inputHash [64]byte, roundingMode uint8) (reg RegisterFile) {

	reg.FPRC = roundingMode

	// buffer first 128 bytes are entropy below rest are program bytes
	var buffer [16*8 + RANDOMX_PROGRAM_SIZE*8]byte
	aes.FillAes4Rx4(inputHash, buffer[:])

	entropy := (*[16]uint64)(unsafe.Pointer(&buffer))

	prog := buffer[len(entropy)*8:]

	// do more initialization before we run

	for i := range entropy[:8] {
		reg.A[i/2][i%2] = SmallPositiveFloatBits(entropy[i])
	}

	var mem MemoryRegisters

	mem.ma = entropy[8] & CacheLineAlignMask
	mem.mx = entropy[10]

	addressRegisters := entropy[12]

	var readReg [4]uint64

	for i := range readReg {
		readReg[i] = uint64(i*2) + (addressRegisters & 1)
		addressRegisters >>= 1
	}

	datasetOffset := (entropy[13] % (DATASETEXTRAITEMS + 1)) * CacheLineSize

	eMask := [2]uint64{EMask(entropy[14]), EMask(entropy[15])}

	byteCode := CompileProgramToByteCode(prog)

	spAddr0 := mem.mx
	spAddr1 := mem.ma

	var rlCache RegisterLine

	if vm.JITProgram != nil {
		if vm.Dataset.Flags()&RANDOMX_FLAG_SECURE > 0 {
			mapProgramRW(vm.JITProgram)
			byteCode.generateCode(vm.JITProgram)
			mapProgramRX(vm.JITProgram)
		} else {
			byteCode.generateCode(vm.JITProgram)
		}
	}

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

		// Run the actual bytecode
		if vm.JITProgram != nil {
			vm.JITProgram.Execute(&reg, &vm.ScratchPad, eMask)
		} else {
			byteCode.Execute(&reg, &vm.ScratchPad, eMask)
		}

		mem.mx ^= reg.R[readReg[2]] ^ reg.R[readReg[3]]
		mem.mx &= CacheLineAlignMask

		vm.Dataset.PrefetchDataset(datasetOffset + mem.mx)
		// execute diffuser superscalar program to get dataset 64 bytes
		vm.Dataset.ReadDataset(datasetOffset+mem.ma, &reg.R, &rlCache)

		// swap the elements
		mem.mx, mem.ma = mem.ma, mem.mx

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

	return reg

}

func (vm *VM) InitScratchpad(seed *[64]byte) {
	vm.ScratchPad.Init(seed)
}

func (vm *VM) RunLoops(tempHash [64]byte) RegisterFile {
	hash512, _ := blake2b.New512(nil)

	if lockThreadDueToRoundingMode {
		// Lock thread due to rounding mode flags
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	roundingMode := uint8(0)

	for chain := 0; chain < RANDOMX_PROGRAM_COUNT-1; chain++ {
		reg := vm.Run(tempHash, roundingMode)
		roundingMode = reg.FPRC

		hash512.Reset()

		// write R, F, E, A registers
		hash512.Write(reg.Memory()[:])
		runtime.KeepAlive(reg)

		hash512.Sum(tempHash[:0])
	}

	// final loop executes here
	reg := vm.Run(tempHash, roundingMode)
	// always force a restore
	reg.FPRC = 0xff

	// restore rounding mode to 0
	SetRoundingMode(&reg, 0)

	return reg
}

func (vm *VM) CalculateHash(input []byte, output *[32]byte) {
	tempHash := blake2b.Sum512(input)

	vm.InitScratchpad(&tempHash)

	reg := vm.RunLoops(tempHash)

	// now hash the scratch pad as it will act as register A
	aes.HashAes1Rx4(vm.ScratchPad[:], &tempHash)

	hash256, _ := blake2b.New256(nil)

	hash256.Reset()

	// write R, F, E registers
	hash256.Write(reg.Memory()[:RegisterFileSize-RegistersCountFloat*2*8])
	runtime.KeepAlive(reg)

	// write register A
	hash256.Write(tempHash[:])

	hash256.Sum(output[:0])
}

func (vm *VM) Close() error {
	if vm.JITProgram != nil {
		return vm.JITProgram.Close()
	}
	return nil
}

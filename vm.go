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
	"errors"
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/aes"
	"math"
	"runtime"
	"unsafe"
)
import "golang.org/x/crypto/blake2b"

type VM struct {
	pad ScratchPad

	flags Flags

	// buffer first 128 bytes are entropy below rest are program bytes
	buffer [16*8 + RANDOMX_PROGRAM_SIZE*8]byte

	hashState [blake2b.Size]byte

	registerFile RegisterFile

	AES aes.AES

	Cache   *Cache
	Dataset *Dataset

	program    ByteCode
	jitProgram VMProgramFunc
}

// NewVM  Creates and initializes a RandomX virtual machine.
// *
// * @param flags is any combination of these 5 flags (each flag can be set or not set):
// *        RANDOMX_FLAG_LARGE_PAGES - allocate scratchpad memory in large pages
// *        RANDOMX_FLAG_HARD_AES - virtual machine will use hardware accelerated AES
// *        RANDOMX_FLAG_FULL_MEM - virtual machine will use the full dataset
// *        RANDOMX_FLAG_JIT - virtual machine will use a JIT compiler
// *        RANDOMX_FLAG_SECURE - when combined with RANDOMX_FLAG_JIT, the JIT pages are never
// *                              writable and executable at the same time (W^X policy)
// *        The numeric values of the first 4 flags are ordered so that a higher value will provide
// *        faster hash calculation and a lower numeric value will provide higher portability.
// *        Using RANDOMX_FLAG_DEFAULT (all flags not set) works on all platforms, but is the slowest.
// * @param cache is a pointer to an initialized randomx_cache structure. Can be
// *        NULL if RANDOMX_FLAG_FULL_MEM is set.
// * @param dataset is a pointer to a randomx_dataset structure. Can be NULL
// *        if RANDOMX_FLAG_FULL_MEM is not set.
// *
// * @return Pointer to an initialized randomx_vm structure.
// *         Returns NULL if:
// *         (1) Scratchpad memory allocation fails.
// *         (2) The requested initialization flags are not supported on the current platform.
// *         (3) cache parameter is NULL and RANDOMX_FLAG_FULL_MEM is not set
// *         (4) dataset parameter is NULL and RANDOMX_FLAG_FULL_MEM is set
// */
func NewVM(flags Flags, cache *Cache, dataset *Dataset) (*VM, error) {
	if cache == nil && !flags.Has(RANDOMX_FLAG_FULL_MEM) {
		return nil, errors.New("nil cache in light mode")
	}
	if dataset == nil && flags.Has(RANDOMX_FLAG_FULL_MEM) {
		return nil, errors.New("nil dataset in full mode")
	}

	vm := &VM{
		Cache:   cache,
		Dataset: dataset,
		flags:   flags,
	}

	if flags.Has(RANDOMX_FLAG_HARD_AES) {
		vm.AES = aes.NewHardAES()
	}
	// fallback
	if vm.AES == nil {
		vm.AES = aes.NewSoftAES()
	}

	if flags.HasJIT() {
		vm.jitProgram = mapProgram(nil, int(RandomXCodeSize))
		if !flags.Has(RANDOMX_FLAG_SECURE) {
			mapProgramRWX(vm.jitProgram)
		}
	}

	return vm, nil
}

// run calculate hash based on input. Not thread safe.
// Warning: Underlying callers will run float64 SetRoundingMode directly
// It is the caller's responsibility to set and restore the mode to IEEE 754 roundTiesToEven between full executions
// Additionally, runtime.LockOSThread and defer runtime.UnlockOSThread is recommended to prevent other goroutines sharing these changes
func (vm *VM) run() {

	// buffer first 128 bytes are entropy below rest are program bytes
	vm.AES.FillAes4Rx4(vm.hashState, vm.buffer[:])

	entropy := (*[16]uint64)(unsafe.Pointer(&vm.buffer))

	// do more initialization before we run

	reg := &vm.registerFile
	reg.Clear()

	// initialize constant registers
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

	prog := vm.buffer[len(entropy)*8:]
	CompileProgramToByteCode(prog, &vm.program)

	var jitProgram VMProgramFunc

	if vm.jitProgram != nil {
		if vm.Dataset == nil { //light mode
			if vm.flags.Has(RANDOMX_FLAG_SECURE) {
				mapProgramRW(vm.jitProgram)
				jitProgram = vm.program.generateCode(vm.jitProgram, nil)
				mapProgramRX(vm.jitProgram)
			} else {
				jitProgram = vm.program.generateCode(vm.jitProgram, nil)
			}
		} else {
			// full mode and we have JIT
			if vm.flags.Has(RANDOMX_FLAG_SECURE) {
				mapProgramRW(vm.jitProgram)
				jitProgram = vm.program.generateCode(vm.jitProgram, &readReg)
				mapProgramRX(vm.jitProgram)
			} else {
				jitProgram = vm.program.generateCode(vm.jitProgram, &readReg)
			}

			vm.jitProgram.ExecuteFull(reg, &vm.pad, &vm.Dataset.Memory()[datasetOffset/CacheLineSize], RANDOMX_PROGRAM_ITERATIONS, ma, mx, eMask)
			return
		}
	}

	spAddr0 := uint64(mx)
	spAddr1 := uint64(ma)

	var rlCache RegisterLine

	for ic := 0; ic < RANDOMX_PROGRAM_ITERATIONS; ic++ {
		spMix := reg.R[readReg[0]] ^ reg.R[readReg[1]]

		spAddr0 ^= spMix
		spAddr0 &= ScratchpadL3Mask64
		spAddr1 ^= spMix >> 32
		spAddr1 &= ScratchpadL3Mask64

		//TODO: optimize these loads!
		for i := uint64(0); i < RegistersCount; i++ {
			reg.R[i] ^= vm.pad.Load64(uint32(spAddr0 + 8*i))
		}

		for i := uint64(0); i < RegistersCountFloat; i++ {
			reg.F[i] = vm.pad.Load32FA(uint32(spAddr1 + 8*i))
		}

		for i := uint64(0); i < RegistersCountFloat; i++ {
			reg.E[i] = vm.pad.Load32FA(uint32(spAddr1 + 8*(i+RegistersCountFloat)))

			reg.E[i][LOW] = MaskRegisterExponentMantissa(reg.E[i][LOW], eMask[LOW])
			reg.E[i][HIGH] = MaskRegisterExponentMantissa(reg.E[i][HIGH], eMask[HIGH])
		}

		// run the actual bytecode
		if jitProgram != nil {
			// light mode
			jitProgram.Execute(reg, &vm.pad, eMask)
		} else {
			vm.program.Execute(reg, &vm.pad, eMask)
		}

		mx ^= uint32(reg.R[readReg[2]] ^ reg.R[readReg[3]])
		mx &= uint32(CacheLineAlignMask)

		if vm.Dataset != nil {
			// full mode
			vm.Dataset.prefetchDataset(datasetOffset + uint64(mx))
			// load output from superscalar program to get dataset 64 bytes
			vm.Dataset.readDataset(datasetOffset+uint64(ma), &reg.R)
		} else {
			// light mode
			// execute output from superscalar program to get dataset 64 bytes
			vm.Cache.initDataset(&rlCache, (datasetOffset+uint64(ma))/CacheLineSize)
			for i := range reg.R {
				reg.R[i] ^= rlCache[i]
			}
		}

		// swap the elements
		mx, ma = ma, mx

		for i := uint64(0); i < RegistersCount; i++ {
			vm.pad.Store64(uint32(spAddr1+8*i), reg.R[i])
		}

		for i := uint64(0); i < RegistersCountFloat; i++ {
			reg.F[i][LOW] = Xor(reg.F[i][LOW], reg.E[i][LOW])
			reg.F[i][HIGH] = Xor(reg.F[i][HIGH], reg.E[i][HIGH])

			vm.pad.Store64(uint32(spAddr0+16*i), math.Float64bits(reg.F[i][LOW]))
			vm.pad.Store64(uint32(spAddr0+16*i+8), math.Float64bits(reg.F[i][HIGH]))
		}

		spAddr0 = 0
		spAddr1 = 0

	}
}

func (vm *VM) initScratchpad(seed *[64]byte) {
	clear(vm.pad[:])
	vm.AES.FillAes1Rx4(seed, vm.pad[:])
}

func (vm *VM) runLoops() {
	if lockThreadDueToRoundingMode {
		// Lock thread due to rounding mode flags
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	// always force a restore before startup
	ResetRoundingMode(&vm.registerFile)

	// restore rounding mode at the end
	defer ResetRoundingMode(&vm.registerFile)

	for chain := 0; chain < RANDOMX_PROGRAM_COUNT-1; chain++ {
		vm.run()

		// write R, F, E, A registers
		vm.hashState = blake2b.Sum512(vm.registerFile.Memory()[:])
	}

	// final loop executes here
	vm.run()
}

// SetCache Reinitializes a virtual machine with a new Cache.
// This function should be called anytime the Cache is reinitialized with a new key.
// Does nothing if called with a Cache containing the same key value as already set.
// VM must be initialized without RANDOMX_FLAG_FULL_MEM.
func (vm *VM) SetCache(cache *Cache) {
	if vm.flags.Has(RANDOMX_FLAG_FULL_MEM) {
		panic("unsupported")
	}
	vm.Cache = cache
	//todo
}

// SetDataset Reinitializes a virtual machine with a new Dataset.
// VM must be initialized with RANDOMX_FLAG_FULL_MEM.
func (vm *VM) SetDataset(dataset *Dataset) {
	if !vm.flags.Has(RANDOMX_FLAG_FULL_MEM) {
		panic("unsupported")
	}
	vm.Dataset = dataset
}

// CalculateHash Calculates a RandomX hash value.
func (vm *VM) CalculateHash(input []byte, output *[RANDOMX_HASH_SIZE]byte) {
	vm.hashState = blake2b.Sum512(input)

	vm.initScratchpad(&vm.hashState)

	vm.runLoops()

	// now hash the scratch pad as it will act as register A
	vm.AES.HashAes1Rx4(vm.pad[:], &vm.hashState)

	regMem := vm.registerFile.Memory()
	// write hash onto register A
	copy(regMem[RegisterFileSize-RegistersCountFloat*2*8:], vm.hashState[:])

	// write R, F, E, A registers
	*output = blake2b.Sum256(regMem[:])
}

// Close Releases all memory occupied by the structure.
func (vm *VM) Close() error {
	if vm.jitProgram != nil {
		return vm.jitProgram.Close()
	}
	return nil
}

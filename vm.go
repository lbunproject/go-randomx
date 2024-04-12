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
	"git.gammaspectra.live/P2Pool/go-randomx/v2/asm"
	"math"
	"runtime"
)
import "encoding/binary"
import "golang.org/x/crypto/blake2b"

type REG struct {
	Hi uint64
	Lo uint64
}

type VM struct {
	StateStart [64]byte
	buffer     [RANDOMX_PROGRAM_SIZE*8 + 16*8]byte // first 128 bytes are entropy below rest are program bytes
	Prog       []byte
	ScratchPad [ScratchpadSize]byte

	ByteCode [RANDOMX_PROGRAM_SIZE]InstructionByteCode

	// program configuration  see program.hpp

	entropy [16]uint64

	reg           REGISTER_FILE // the register file
	mem           MemoryRegisters
	config        Config // configuration
	datasetOffset uint64

	Dataset Randomx_Dataset

	Cache *Randomx_Cache // randomx cache

}

func MaskRegisterExponentMantissa(f float64, mode uint64) float64 {
	return math.Float64frombits((math.Float64bits(f) & dynamicMantissaMask) | mode)
}

type Config struct {
	eMask   [2]uint64
	readReg [4]uint64
}

type REGISTER_FILE struct {
	r RegisterLine
	f [4][2]float64
	e [4][2]float64
	a [4][2]float64
}
type MemoryRegisters struct {
	mx, ma uint64
}

const LOW = 0
const HIGH = 1

// calculate hash based on input
func (vm *VM) Run(input_hash [64]byte) {

	//fmt.Printf("%x \n", input_hash)

	aes.FillAes4Rx4(input_hash, vm.buffer[:])

	for i := range vm.entropy {
		vm.entropy[i] = binary.LittleEndian.Uint64(vm.buffer[i*8:])
	}

	vm.Prog = vm.buffer[len(vm.entropy)*8:]

	clear(vm.reg.r[:])

	// do more initialization before we run

	for i := range vm.entropy[:8] {
		vm.reg.a[i/2][i%2] = math.Float64frombits(getSmallPositiveFloatBits(vm.entropy[i]))
	}

	vm.mem.ma = vm.entropy[8] & CacheLineAlignMask
	vm.mem.mx = vm.entropy[10]

	addressRegisters := vm.entropy[12]
	for i := range vm.config.readReg {
		vm.config.readReg[i] = uint64(i*2) + (addressRegisters & 1)
		addressRegisters >>= 1
	}

	vm.datasetOffset = (vm.entropy[13] % (DATASETEXTRAITEMS + 1)) * CacheLineSize
	vm.config.eMask[LOW] = getFloatMask(vm.entropy[14])
	vm.config.eMask[HIGH] = getFloatMask(vm.entropy[15])

	//fmt.Printf("prog %x  entropy 0 %x %f \n", vm.buffer[:32], vm.entropy[0], vm.reg.a[0][HIGH])

	vm.Compile_TO_Bytecode()

	spAddr0 := vm.mem.mx
	spAddr1 := vm.mem.ma

	var rlCache RegisterLine

	for ic := 0; ic < RANDOMX_PROGRAM_ITERATIONS; ic++ {
		spMix := vm.reg.r[vm.config.readReg[0]] ^ vm.reg.r[vm.config.readReg[1]]

		spAddr0 ^= spMix
		spAddr0 &= ScratchpadL3Mask64
		spAddr1 ^= spMix >> 32
		spAddr1 &= ScratchpadL3Mask64

		for i := uint64(0); i < REGISTERSCOUNT; i++ {
			vm.reg.r[i] ^= vm.Load64(spAddr0 + 8*i)
		}

		for i := uint64(0); i < REGISTERCOUNTFLT; i++ {
			vm.reg.f[i] = vm.Load32FA(spAddr1 + 8*i)
		}

		for i := uint64(0); i < REGISTERCOUNTFLT; i++ {
			vm.reg.e[i] = vm.Load32FA(spAddr1 + 8*(i+REGISTERCOUNTFLT))

			vm.reg.e[i][LOW] = MaskRegisterExponentMantissa(vm.reg.e[i][LOW], vm.config.eMask[LOW])
			vm.reg.e[i][HIGH] = MaskRegisterExponentMantissa(vm.reg.e[i][HIGH], vm.config.eMask[HIGH])
		}

		// todo: pass register file directly!
		vm.InterpretByteCode()

		vm.mem.mx ^= vm.reg.r[vm.config.readReg[2]] ^ vm.reg.r[vm.config.readReg[3]]
		vm.mem.mx &= CacheLineAlignMask

		vm.Dataset.PrefetchDataset(vm.datasetOffset + vm.mem.mx)
		// execute diffuser superscalar program to get dataset 64 bytes
		vm.Dataset.ReadDataset(vm.datasetOffset+vm.mem.ma, &vm.reg.r, &rlCache)

		// swap the elements
		vm.mem.mx, vm.mem.ma = vm.mem.ma, vm.mem.mx

		for i := uint64(0); i < REGISTERSCOUNT; i++ {
			binary.LittleEndian.PutUint64(vm.ScratchPad[spAddr1+8*i:], vm.reg.r[i])
		}

		for i := uint64(0); i < REGISTERCOUNTFLT; i++ {
			vm.reg.f[i][LOW] = math.Float64frombits(math.Float64bits(vm.reg.f[i][LOW]) ^ math.Float64bits(vm.reg.e[i][LOW]))
			vm.reg.f[i][HIGH] = math.Float64frombits(math.Float64bits(vm.reg.f[i][HIGH]) ^ math.Float64bits(vm.reg.e[i][HIGH]))

			binary.LittleEndian.PutUint64(vm.ScratchPad[spAddr0+16*i:], math.Float64bits(vm.reg.f[i][LOW]))
			binary.LittleEndian.PutUint64(vm.ScratchPad[spAddr0+16*i+8:], math.Float64bits(vm.reg.f[i][HIGH]))
		}

		spAddr0 = 0
		spAddr1 = 0

	}

}

func (vm *VM) InitScratchpad(seed *[64]byte) {
	// calculate and fill scratchpad
	clear(vm.ScratchPad[:])
	aes.FillAes1Rx4(seed, vm.ScratchPad[:])
}

func (vm *VM) CalculateHash(input []byte, output *[32]byte) {
	var buf [8]byte

	// Lock thread due to rounding mode flags
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	//restore rounding mode to golang expected one
	defer asm.SetRoundingMode(asm.RoundingModeToNearest)

	// reset rounding mode if new hash being calculated
	asm.SetRoundingMode(asm.RoundingModeToNearest)

	tempHash := blake2b.Sum512(input)

	vm.InitScratchpad(&tempHash)

	hash512, _ := blake2b.New512(nil)

	for chain := 0; chain < RANDOMX_PROGRAM_COUNT-1; chain++ {
		vm.Run(tempHash)

		hash512.Reset()
		for i := range vm.reg.r {
			binary.LittleEndian.PutUint64(buf[:], vm.reg.r[i])
			hash512.Write(buf[:])
		}
		for i := range vm.reg.f {
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.f[i][LOW]))
			hash512.Write(buf[:])
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.f[i][HIGH]))
			hash512.Write(buf[:])
		}

		for i := range vm.reg.e {
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.e[i][LOW]))
			hash512.Write(buf[:])
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.e[i][HIGH]))
			hash512.Write(buf[:])
		}

		for i := range vm.reg.a {
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.a[i][LOW]))
			hash512.Write(buf[:])
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.a[i][HIGH]))
			hash512.Write(buf[:])
		}

		hash512.Sum(tempHash[:0])
		//fmt.Printf("%d temphash %x\n", chain, tempHash)
	}

	// final loop executes here
	vm.Run(tempHash)

	// now hash the scratch pad and place into register a
	aes.HashAes1Rx4(vm.ScratchPad[:], &tempHash)

	hash256, _ := blake2b.New256(nil)

	hash256.Reset()

	for i := range vm.reg.r {
		binary.LittleEndian.PutUint64(buf[:], vm.reg.r[i])
		hash256.Write(buf[:])
	}

	for i := range vm.reg.f {
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.f[i][LOW]))
		hash256.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.f[i][HIGH]))
		hash256.Write(buf[:])
	}

	for i := range vm.reg.e {
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.e[i][LOW]))
		hash256.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(vm.reg.e[i][HIGH]))
		hash256.Write(buf[:])
	}

	// copy tempHash as it first copied to register and then hashed
	hash256.Write(tempHash[:])

	hash256.Sum(output[:0])
}

const mask22bit = (uint64(1) << 22) - 1

func getSmallPositiveFloatBits(entropy uint64) uint64 {
	exponent := entropy >> 59 //0..31
	mantissa := entropy & mantissaMask
	exponent += exponentBias
	exponent &= exponentMask
	exponent = exponent << mantissaSize
	return exponent | mantissa
}

func getStaticExponent(entropy uint64) uint64 {
	exponent := constExponentBits
	exponent |= (entropy >> (64 - staticExponentBits)) << dynamicExponentBits
	exponent <<= mantissaSize
	return exponent
}

func getFloatMask(entropy uint64) uint64 {
	return (entropy & mask22bit) | getStaticExponent(entropy)
}

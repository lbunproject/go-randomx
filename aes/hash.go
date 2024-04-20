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

package aes

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/keys"
	"unsafe"
)

// HashAes1Rx4
//
//	Calculate a 512-bit hash of 'input' using 4 lanes of AES.
//	The input is treated as a set of round keys for the encryption
//	of the initial state.
//
//	'inputSize' must be a multiple of 64.
//
//	For a 2 MiB input, this has the same security as 32768-round
//	AES encryption.
//
//	Hashing throughput: >20 GiB/s per CPU core with hardware AES
func HashAes1Rx4(input []byte, output *[64]byte) {
	if len(input)%64 != 0 {
		panic("unsupported")
	}
	hashAes1Rx4(input, output)
}

// FillAes1Rx4
//
//	Fill 'output' with pseudorandom data based on 512-bit 'state'.
//	The state is encrypted using a single AES round per 16 bytes of output
//	in 4 lanes.
//
//	'output' size must be a multiple of 64.
//
//	The modified state is written back to 'state' to allow multiple
//	calls to this function.
func FillAes1Rx4(state *[64]byte, output []byte) {
	if len(output)%len(state) != 0 {
		panic("unsupported")
	}
	fillAes1Rx4(state, output)
}

var fillAes4Rx4Keys0 = [4][4]uint32{
	keys.AesGenerator4R_Keys[0],
	keys.AesGenerator4R_Keys[0],
	keys.AesGenerator4R_Keys[4],
	keys.AesGenerator4R_Keys[4],
}
var fillAes4Rx4Keys1 = [4][4]uint32{
	keys.AesGenerator4R_Keys[1],
	keys.AesGenerator4R_Keys[1],
	keys.AesGenerator4R_Keys[5],
	keys.AesGenerator4R_Keys[5],
}
var fillAes4Rx4Keys2 = [4][4]uint32{
	keys.AesGenerator4R_Keys[2],
	keys.AesGenerator4R_Keys[2],
	keys.AesGenerator4R_Keys[6],
	keys.AesGenerator4R_Keys[6],
}
var fillAes4Rx4Keys3 = [4][4]uint32{
	keys.AesGenerator4R_Keys[3],
	keys.AesGenerator4R_Keys[3],
	keys.AesGenerator4R_Keys[7],
	keys.AesGenerator4R_Keys[7],
}

// FillAes4Rx4 used to generate final program
func FillAes4Rx4(state [64]byte, output []byte) {
	if len(output)%len(state) != 0 {
		panic("unsupported")
	}

	// state is copied on caller

	// Copy state
	states := (*[4][4]uint32)(unsafe.Pointer(&state))

	for outptr := 0; outptr < len(output); outptr += len(state) {
		aesroundtrip_decenc(states, &fillAes4Rx4Keys0)
		aesroundtrip_decenc(states, &fillAes4Rx4Keys1)
		aesroundtrip_decenc(states, &fillAes4Rx4Keys2)
		aesroundtrip_decenc(states, &fillAes4Rx4Keys3)

		copy(output[outptr:], state[:])
	}

}

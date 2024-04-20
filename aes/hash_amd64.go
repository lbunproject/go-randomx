//go:build amd64 && !purego

package aes

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/asm"
	"git.gammaspectra.live/P2Pool/go-randomx/v3/keys"
	"golang.org/x/sys/cpu"
	"unsafe"
)

var supportsAES = cpu.X86.HasAES

func fillAes1Rx4(state *[64]byte, output []byte) {
	// Reference to state without copying
	states := (*[4][4]uint32)(unsafe.Pointer(state))

	if supportsAES {
		asm.FillAes1Rx4(states, &keys.AesGenerator1R_Keys, unsafe.SliceData(output), uint64(len(output)))
		return
	}

	for outptr := 0; outptr < len(output); outptr += len(state) {
		aesroundtrip_decenc(states, &keys.AesGenerator1R_Keys)

		copy(output[outptr:], state[:])
	}
}

func hashAes1Rx4(input []byte, output *[64]byte) {
	if supportsAES {
		asm.HashAes1Rx4(&keys.AesHash1R_State, &keys.AesHash1R_XKeys, output, unsafe.SliceData(input), uint64(len(input)))
		return
	}

	// states are copied
	states := keys.AesHash1R_State

	for input_ptr := 0; input_ptr < len(input); input_ptr += 64 {
		in := (*[4][4]uint32)(unsafe.Pointer(unsafe.SliceData(input[input_ptr:])))

		aesroundtrip_encdec(&states, in)
	}

	aesroundtrip_encdec1(&states, &keys.AesHash1R_XKeys[0])

	aesroundtrip_encdec1(&states, &keys.AesHash1R_XKeys[1])

	copy(output[:], (*[64]byte)(unsafe.Pointer(&states))[:])
}

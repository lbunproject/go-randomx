package aes

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/keys"
	"runtime"
	"unsafe"
)

type softAES struct {
}

func NewSoftAES() AES {
	return softAES{}
}

func (aes softAES) HashAes1Rx4(input []byte, output *[64]byte) {
	if len(input)%len(output) != 0 {
		panic("unsupported")
	}
	// states are copied
	states := (*[4][4]uint32)(unsafe.Pointer(output))
	*states = keys.AesHash1R_State

	for input_ptr := 0; input_ptr < len(input); input_ptr += 64 {
		in := (*[4][4]uint32)(unsafe.Pointer(unsafe.SliceData(input[input_ptr:])))

		soft_aesroundtrip_encdec(states, in)
	}

	soft_aesroundtrip_encdec1(states, &keys.AesHash1R_XKeys[0])

	soft_aesroundtrip_encdec1(states, &keys.AesHash1R_XKeys[1])

	runtime.KeepAlive(output)
}

func (aes softAES) FillAes1Rx4(state *[64]byte, output []byte) {
	if len(output)%len(state) != 0 {
		panic("unsupported")
	}
	// Reference to state without copying
	states := (*[4][4]uint32)(unsafe.Pointer(state))

	for outptr := 0; outptr < len(output); outptr += len(state) {
		soft_aesroundtrip_decenc(states, &keys.AesGenerator1R_Keys)

		copy(output[outptr:], state[:])
	}
}

func (aes softAES) FillAes4Rx4(state [64]byte, output []byte) {
	if len(output)%len(state) != 0 {
		panic("unsupported")
	}

	// state is copied on caller

	// Copy state
	states := (*[4][4]uint32)(unsafe.Pointer(&state))

	for outptr := 0; outptr < len(output); outptr += len(state) {
		soft_aesroundtrip_decenc(states, &fillAes4Rx4Keys0)
		soft_aesroundtrip_decenc(states, &fillAes4Rx4Keys1)
		soft_aesroundtrip_decenc(states, &fillAes4Rx4Keys2)
		soft_aesroundtrip_decenc(states, &fillAes4Rx4Keys3)

		copy(output[outptr:], state[:])
	}
}

func (aes softAES) HashAndFillAes1Rx4(scratchpad []byte, output *[64]byte, fillState *[64]byte) {
	//TODO
	aes.HashAes1Rx4(scratchpad, output)
	aes.FillAes1Rx4(fillState, scratchpad)
}

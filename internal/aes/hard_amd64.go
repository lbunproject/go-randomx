//go:build amd64 && !purego

package aes

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/asm"
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/keys"
	"golang.org/x/sys/cpu"
	"runtime"
	"unsafe"
)

const HasHardAESImplementation = true

type hardAES struct {
}

func NewHardAES() AES {
	if cpu.X86.HasAES {
		return hardAES{}
	}

	return nil
}

func (h hardAES) HashAes1Rx4(input []byte, output *[64]byte) {
	if len(input)%len(output) != 0 {
		panic("unsupported")
	}

	asm.HashAes1Rx4(&keys.AesHash1R_State, &keys.AesHash1R_XKeys, output, unsafe.SliceData(input), uint64(len(input)))
}

func (h hardAES) FillAes1Rx4(state *[64]byte, output []byte) {
	if len(output)%len(state) != 0 {
		panic("unsupported")
	}

	// Reference to state without copying
	states := (*[4][4]uint32)(unsafe.Pointer(state))
	asm.FillAes1Rx4(states, &keys.AesGenerator1R_Keys, unsafe.SliceData(output), uint64(len(output)))
	runtime.KeepAlive(state)
}

func (h hardAES) FillAes4Rx4(state [64]byte, output []byte) {
	if len(output)%len(state) != 0 {
		panic("unsupported")
	}

	// state is copied on caller

	// Copy state
	states := (*[4][4]uint32)(unsafe.Pointer(&state))

	for outptr := 0; outptr < len(output); outptr += len(state) {
		asm.AESRoundTrip_DecEnc(states, &fillAes4Rx4Keys0)
		asm.AESRoundTrip_DecEnc(states, &fillAes4Rx4Keys1)
		asm.AESRoundTrip_DecEnc(states, &fillAes4Rx4Keys2)
		asm.AESRoundTrip_DecEnc(states, &fillAes4Rx4Keys3)

		copy(output[outptr:], state[:])
	}
}

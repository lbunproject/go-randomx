//go:build amd64 && !purego

package aes

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/asm"
)

func aesroundtrip_decenc(states *[4][4]uint32, keys *[4][4]uint32) {
	if supportsAES {
		asm.AESRoundTrip_DecEnc(states, keys)
	} else {
		soft_aesdec(&states[0], &keys[0])
		soft_aesenc(&states[1], &keys[1])
		soft_aesdec(&states[2], &keys[2])
		soft_aesenc(&states[3], &keys[3])
	}
}

func aesroundtrip_encdec(states *[4][4]uint32, keys *[4][4]uint32) {
	if supportsAES {
		asm.AESRoundTrip_EncDec(states, keys)
	} else {
		soft_aesenc(&states[0], &keys[0])
		soft_aesdec(&states[1], &keys[1])
		soft_aesenc(&states[2], &keys[2])
		soft_aesdec(&states[3], &keys[3])
	}
}

func aesroundtrip_encdec1(states *[4][4]uint32, key *[4]uint32) {
	if supportsAES {
		asm.AESRoundTrip_EncDec1(states, key)
	} else {
		soft_aesenc(&states[0], key)
		soft_aesdec(&states[1], key)
		soft_aesenc(&states[2], key)
		soft_aesdec(&states[3], key)
	}
}

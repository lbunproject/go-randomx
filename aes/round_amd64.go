//go:build amd64 && !purego

package aes

import (
	_ "git.gammaspectra.live/P2Pool/go-randomx/v2/asm"
	"golang.org/x/sys/cpu"
	_ "unsafe"
)

//go:linkname hard_aesdec git.gammaspectra.live/P2Pool/go-randomx/v2/asm.aesdec
func hard_aesdec(state *[4]uint32, key *[4]uint32)

//go:linkname hard_aesenc git.gammaspectra.live/P2Pool/go-randomx/v2/asm.aesenc
func hard_aesenc(state *[4]uint32, key *[4]uint32)

//go:linkname hard_aesroundtrip_decenc git.gammaspectra.live/P2Pool/go-randomx/v2/asm.aesroundtrip_decenc
func hard_aesroundtrip_decenc(states *[4][4]uint32, keys *[4][4]uint32)

//go:linkname hard_aesroundtrip_encdec git.gammaspectra.live/P2Pool/go-randomx/v2/asm.aesroundtrip_encdec
func hard_aesroundtrip_encdec(states *[4][4]uint32, keys *[4][4]uint32)

//go:linkname hard_aesroundtrip_encdec1 git.gammaspectra.live/P2Pool/go-randomx/v2/asm.aesroundtrip_encdec1
func hard_aesroundtrip_encdec1(states *[4][4]uint32, key *[4]uint32)

var supportsAES = cpu.X86.HasAES

func aesenc(state *[4]uint32, key *[4]uint32) {
	if supportsAES {
		hard_aesenc(state, key)
	} else {
		soft_aesenc(state, key)
	}
}

func aesdec(state *[4]uint32, key *[4]uint32) {
	if supportsAES {
		hard_aesdec(state, key)
	} else {
		soft_aesdec(state, key)
	}
}

func aesroundtrip_decenc(states *[4][4]uint32, keys *[4][4]uint32) {
	if supportsAES {
		hard_aesroundtrip_decenc(states, keys)
	} else {
		aesdec(&states[0], &keys[0])
		aesenc(&states[1], &keys[1])
		aesdec(&states[2], &keys[2])
		aesenc(&states[3], &keys[3])
	}
}

func aesroundtrip_encdec(states *[4][4]uint32, keys *[4][4]uint32) {
	if supportsAES {
		hard_aesroundtrip_encdec(states, keys)
	} else {
		aesenc(&states[0], &keys[0])
		aesdec(&states[1], &keys[1])
		aesenc(&states[2], &keys[2])
		aesdec(&states[3], &keys[3])
	}
}

func aesroundtrip_encdec1(states *[4][4]uint32, key *[4]uint32) {
	if supportsAES {
		hard_aesroundtrip_encdec1(states, key)
	} else {
		aesenc(&states[0], key)
		aesdec(&states[1], key)
		aesenc(&states[2], key)
		aesdec(&states[3], key)
	}
}

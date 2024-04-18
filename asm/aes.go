//go:build amd64 && !purego

package asm

func AESRoundEncrypt(state *[4]uint32, key *[4]uint32) {
	aesenc(state, key)
}

func AESRoundDecrypt(state *[4]uint32, key *[4]uint32) {
	aesdec(state, key)
}

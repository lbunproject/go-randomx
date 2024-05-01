//go:build !amd64 || purego

package asm

func AESRoundEncrypt(state *[4]uint32, key *[4]uint32) {
	panic("not implemented")
}

func AESRoundDecrypt(state *[4]uint32, key *[4]uint32) {
	panic("not implemented")
}

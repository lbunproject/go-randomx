//go:build (!arm64 && !amd64 && !386) || purego

package asm

func setRoundingMode(mode uint8) {
	panic("not implemented")
}

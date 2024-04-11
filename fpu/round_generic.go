//go:build !arm64 && !amd64 && !386

package fpu

func setRoundingMode(mode uint8) {
	panic("not implemented")
}

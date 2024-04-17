//go:build (!arm64 && !amd64 && !386) || purego

package randomx

import (
	"git.gammaspectra.live/P2Pool/softfloat64"
	"unsafe"
)

func (pad *ScratchPad) Load32F(addr uint32) (lo, hi float64) {
	a := *(*[2]int32)(unsafe.Pointer(&pad[addr]))
	return softfloat64.Int32ToFloat64(a[LOW]), softfloat64.Int32ToFloat64(a[HIGH])
}

func (pad *ScratchPad) Load32FA(addr uint32) [2]float64 {
	a := *(*[2]int32)(unsafe.Pointer(&pad[addr]))
	return [2]float64{softfloat64.Int32ToFloat64(a[LOW]), softfloat64.Int32ToFloat64(a[HIGH])}
}

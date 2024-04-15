package softfloat

import (
	_ "runtime"
	_ "unsafe"
)

//go:linkname funpack64 runtime.funpack64
func funpack64(f uint64) (sign, mant uint64, exp int, inf, nan bool)

//go:linkname fpack64 runtime.fpack64
func fpack64(sign, mant uint64, exp int, trunc uint64) uint64

//go:linkname fadd64 runtime.fadd64
func fadd64(f, g uint64) uint64

//go:linkname fsub64 runtime.fsub64
func fsub64(f, g uint64) uint64

//go:linkname fneg64 runtime.fneg64
func fneg64(f uint64) uint64

//go:linkname fmul64 runtime.fmul64
func fmul64(f uint64) uint64

//go:linkname fdiv64 runtime.fdiv64
func fdiv64(f uint64) uint64

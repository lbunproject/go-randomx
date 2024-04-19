//go:build unix && amd64 && !disable_jit && !purego

#include "textflag.h"

TEXT ·superscalar_run(SB),$0-16

	MOVQ rf+0(FP), SI

    PREFETCHNTA 0(SI)

    // move register line to registers
    MOVQ 0(SI), R8
    MOVQ 8(SI), R9
    MOVQ 16(SI), R10
    MOVQ 24(SI), R11
    MOVQ 32(SI), R12
    MOVQ 40(SI), R13
    MOVQ 48(SI), R14
    MOVQ 56(SI), R15

    MOVQ jmp+8(FP), AX
    // jump to JIT code
    CALL AX


    // todo: not supported by golang
    // PREFETCHW 0(SI)

    // move registers back to register line
    MOVQ R8, 0(SI)
    MOVQ R9, 8(SI)
    MOVQ R10, 16(SI)
    MOVQ R11, 24(SI)
    MOVQ R12, 32(SI)
    MOVQ R13, 40(SI)
    MOVQ R14, 48(SI)
    MOVQ R15, 56(SI)

    RET

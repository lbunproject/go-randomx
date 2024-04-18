//go:build amd64 && !purego

#include "textflag.h"

// func Cpuid(op uint32) (eax, ebx, ecx, edx uint32)
TEXT ·Cpuid(SB), 7, $0
        XORQ CX, CX
        MOVL op+0(FP), AX
        CPUID
        MOVL AX, eax+8(FP)
        MOVL BX, ebx+12(FP)
        MOVL CX, ecx+16(FP)
        MOVL DX, edx+20(FP)
        RET


// func Cpuidex(op, op2 uint32) (eax, ebx, ecx, edx uint32)
TEXT ·Cpuidex(SB), 7, $0
        MOVL op+0(FP), AX
        MOVL op2+4(FP), CX
        CPUID
        MOVL AX, eax+8(FP)
        MOVL BX, ebx+12(FP)
        MOVL CX, ecx+16(FP)
        MOVL DX, edx+20(FP)
        RET

// func xgetbv(index uint32) (eax, edx uint32)
TEXT ·Xgetbv(SB), 7, $0
        MOVL index+0(FP), CX
        BYTE $0x0f; BYTE $0x01; BYTE $0xd0 // XGETBV
        MOVL AX, eax+8(FP)
        MOVL DX, edx+12(FP)
        RET

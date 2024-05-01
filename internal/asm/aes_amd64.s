//go:build amd64 && !purego

#include "textflag.h"

TEXT ·FillAes1Rx4(SB),NOSPLIT|NOFRAME,$0-32
	MOVQ states+0(FP), AX
	MOVQ keys+8(FP), BX
	MOVQ output+16(FP), CX
	MOVQ outputLen+24(FP), DX

    // initial state
	VMOVDQU 0(AX), X0
	VMOVDQU 16(AX), X1
	VMOVDQU 32(AX), X2
	VMOVDQU 48(AX), X3

    // keys: X4-X7
	VMOVDQU 0(BX), X4
	VMOVDQU 16(BX), X5
	VMOVDQU 32(BX), X6
	VMOVDQU 48(BX), X7

loop:

	AESDEC X4, X0
	AESENC X5, X1
	AESDEC X6, X2
    AESENC X7, X3

    // store state onto output
    VMOVDQU X0, 0(CX)
    VMOVDQU X1, 16(CX)
    VMOVDQU X2, 32(CX)
    VMOVDQU X3, 48(CX)
    ADDQ $64, CX

    // outputLen -= 64, continue if not 0
    SUBQ $64, DX
    JNE loop

    // offload initial state
	VMOVDQU X0, 0(AX)
	VMOVDQU X1, 16(AX)
	VMOVDQU X2, 32(AX)
	VMOVDQU X3, 48(AX)
	RET


TEXT ·HashAes1Rx4(SB),NOSPLIT|NOFRAME,$0-40
	MOVQ initialState+0(FP), AX

    // initial state
	VMOVDQU 0(AX), X0
	VMOVDQU 16(AX), X1
	VMOVDQU 32(AX), X2
	VMOVDQU 48(AX), X3


	MOVQ xKeys+8(FP), AX
	MOVQ output+16(FP), BX
	MOVQ input+24(FP), CX
	MOVQ inputLen+32(FP), DX

loop:
    // input as keys: X4-X7
	VMOVDQU 0(CX), X4
	VMOVDQU 16(CX), X5
	VMOVDQU 32(CX), X6
	VMOVDQU 48(CX), X7

	AESENC X4, X0
	AESDEC X5, X1
	AESENC X6, X2
    AESDEC X7, X3

    ADDQ $64, CX
    // inputLen -= 64, continue if not 0
    SUBQ $64, DX
    JNE loop

    // do encdec1 with both keys!
	VMOVDQU 0(AX), X4
	VMOVDQU 16(AX), X5

    AESENC X4, X0
    AESDEC X4, X1
    AESENC X4, X2
    AESDEC X4, X3

    AESENC X5, X0
    AESDEC X5, X1
    AESENC X5, X2
    AESDEC X5, X3

    // offload into output
	VMOVDQU X0, 0(BX)
	VMOVDQU X1, 16(BX)
	VMOVDQU X2, 32(BX)
	VMOVDQU X3, 48(BX)
	RET

TEXT ·AESRoundTrip_DecEnc(SB),NOSPLIT|NOFRAME,$0-16
	MOVQ states+0(FP), AX
	MOVQ keys+8(FP), BX

	VMOVDQU 0(AX), X0
	VMOVDQU 0(BX), X1
	VMOVDQU 16(AX), X2
	VMOVDQU 16(BX), X3
	VMOVDQU 32(AX), X4
	VMOVDQU 32(BX), X5
	VMOVDQU 48(AX), X6
	VMOVDQU 48(BX), X7

	AESDEC X1, X0
	AESENC X3, X2
	AESDEC X5, X4
    AESENC X7, X6

	VMOVDQU X0, 0(AX)
	VMOVDQU X2, 16(AX)
	VMOVDQU X4, 32(AX)
	VMOVDQU X6, 48(AX)
	RET


TEXT ·AESRoundTrip_EncDec(SB),NOSPLIT|NOFRAME,$0-16
	MOVQ states+0(FP), AX
	MOVQ keys+8(FP), BX

	VMOVDQU 0(AX), X0
	VMOVDQU 0(BX), X1
	VMOVDQU 16(AX), X2
	VMOVDQU 16(BX), X3
	VMOVDQU 32(AX), X4
	VMOVDQU 32(BX), X5
	VMOVDQU 48(AX), X6
	VMOVDQU 48(BX), X7

	AESENC X1, X0
	AESDEC X3, X2
	AESENC X5, X4
    AESDEC X7, X6

	VMOVDQU X0, 0(AX)
	VMOVDQU X2, 16(AX)
	VMOVDQU X4, 32(AX)
	VMOVDQU X6, 48(AX)
	RET


TEXT ·AESRoundTrip_EncDec1(SB),NOSPLIT|NOFRAME,$0-16
	MOVQ states+0(FP), AX
	MOVQ key+8(FP), BX

	VMOVDQU 0(BX), X0
	VMOVDQU 0(AX), X1
	VMOVDQU 16(AX), X2
	VMOVDQU 32(AX), X3
	VMOVDQU 48(AX), X4

	AESENC X0, X1
	AESDEC X0, X2
	AESENC X0, X3
    AESDEC X0, X4

	VMOVDQU X1, 0(AX)
	VMOVDQU X2, 16(AX)
	VMOVDQU X3, 32(AX)
	VMOVDQU X4, 48(AX)
	RET

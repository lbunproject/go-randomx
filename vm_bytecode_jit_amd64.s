//go:build unix && amd64 && !disable_jit && !purego

#include "textflag.h"

TEXT ·vm_run(SB),$8-40

    // move register file to registers
	MOVQ rf+0(FP), AX

    PREFETCHNTA 0(AX)
    // r0-r7
    MOVQ (0*8)(AX), R8
    MOVQ (1*8)(AX), R9
    MOVQ (2*8)(AX), R10
    MOVQ (3*8)(AX), R11
    MOVQ (4*8)(AX), R12
    MOVQ (5*8)(AX), R13
    MOVQ (6*8)(AX), R14
    MOVQ (7*8)(AX), R15

    // f0-f3
    VMOVUPD (8*8)(AX), X0
    VMOVUPD (10*8)(AX), X1
    VMOVUPD (12*8)(AX), X2
    VMOVUPD (14*8)(AX), X3
    // e0-e3
    VMOVUPD (16*8)(AX), X4
    VMOVUPD (18*8)(AX), X5
    VMOVUPD (20*8)(AX), X6
    VMOVUPD (22*8)(AX), X7
    // a0-a3
    VMOVUPD (24*8)(AX), X8
    VMOVUPD (26*8)(AX), X9
    VMOVUPD (28*8)(AX), X10
    VMOVUPD (30*8)(AX), X11

    //TODO: rest of init

    // mantissa mask
	//VMOVQ $0x00ffffffffffffff, $0x00ffffffffffffff, X13
    MOVQ $0x00ffffffffffffff, AX
	VMOVQ AX, X13
	VPBROADCASTQ X13, X13

    // eMask
	VMOVDQU64 eMask+16(FP), X14

    // scale mask
	//VMOVQ $0x80F0000000000000, $0x80F0000000000000, X15
    MOVQ $0x80F0000000000000, AX
	VMOVQ AX, X15
	VPBROADCASTQ X15, X15

    // scratchpad pointer
    MOVQ pad+8(FP), SI

    // JIT location
    MOVQ jmp+32(FP), AX

    // jump to JIT code
    CALL AX


    // move register file back to registers
	MOVQ rf+0(FP), AX

    PREFETCHT0 0(AX)
    // r0-r7
    MOVQ R8, (0*8)(AX)
    MOVQ R9, (1*8)(AX)
    MOVQ R10, (2*8)(AX)
    MOVQ R11, (3*8)(AX)
    MOVQ R12, (4*8)(AX)
    MOVQ R13, (5*8)(AX)
    MOVQ R14, (6*8)(AX)
    MOVQ R15, (7*8)(AX)

    // f0-f3
    VMOVUPD X0, (8*8)(AX)
    VMOVUPD X1, (10*8)(AX)
    VMOVUPD X2, (12*8)(AX)
    VMOVUPD X3, (14*8)(AX)
    // e0-e3
    VMOVUPD X4, (16*8)(AX)
    VMOVUPD X5, (18*8)(AX)
    VMOVUPD X6, (20*8)(AX)
    VMOVUPD X7, (22*8)(AX)

    // a0-a3 are constant, no need to move

    RET

package randomx

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v2/softfloat"
	"math"
	"math/bits"
)

type ByteCodeInstruction struct {
	Dst, Src byte
	ImmB     uint8
	Opcode   ByteCodeInstructionOp
	MemMask  uint32
	Imm      uint64
	/*
		union {
			int_reg_t* idst;
			rx_vec_f128* fdst;
		};
		union {
			int_reg_t* isrc;
			rx_vec_f128* fsrc;
		};
		union {
			uint64_t imm;
			int64_t simm;
		};
		InstructionType type;
		union {
			int16_t target;
			uint16_t shift;
		};
		uint32_t memMask;
	*/

}

func (i ByteCodeInstruction) jumpTarget() int {
	return int(int16((uint16(i.ImmB) << 8) | uint16(i.Dst)))
}

func (i ByteCodeInstruction) getScratchpadAddress(ptr uint64) uint32 {
	return uint32(ptr+i.Imm) & i.MemMask
}

func (i ByteCodeInstruction) getScratchpadZeroAddress() uint32 {
	return uint32(i.Imm) & i.MemMask
}

type ByteCode [RANDOMX_PROGRAM_SIZE]ByteCodeInstruction

// Execute Runs a RandomX program with the given register file and scratchpad
// Warning: This will call asm.SetRoundingMode directly
// It is the caller's responsibility to set and restore the mode to softfloat.RoundingModeToNearest between full executions
// Additionally, runtime.LockOSThread and defer runtime.UnlockOSThread is recommended to prevent other goroutines sharing these changes
func (c *ByteCode) Execute(f *RegisterFile, pad *ScratchPad, eMask [2]uint64) {
	for pc := 0; pc < RANDOMX_PROGRAM_SIZE; pc++ {
		i := &c[pc]
		switch i.Opcode {
		case VM_NOP: // we do nothing
		case VM_IADD_RS:
			f.R[i.Dst] += (f.R[i.Src] << i.ImmB) + i.Imm
		case VM_IADD_M:
			f.R[i.Dst] += pad.Load64(i.getScratchpadAddress(f.R[i.Src]))
		case VM_IADD_MZ:
			f.R[i.Dst] += pad.Load64(uint32(i.Imm))
		case VM_ISUB_R:
			f.R[i.Dst] -= f.R[i.Src]
		case VM_ISUB_I:
			f.R[i.Dst] -= i.Imm
		case VM_ISUB_M:
			f.R[i.Dst] -= pad.Load64(i.getScratchpadAddress(f.R[i.Src]))
		case VM_ISUB_MZ:
			f.R[i.Dst] -= pad.Load64(uint32(i.Imm))
		case VM_IMUL_R:
			f.R[i.Dst] *= f.R[i.Src]
		case VM_IMUL_I:
			// also handles imul_rcp
			f.R[i.Dst] *= i.Imm
		case VM_IMUL_M:
			f.R[i.Dst] *= pad.Load64(i.getScratchpadAddress(f.R[i.Src]))
		case VM_IMUL_MZ:
			f.R[i.Dst] *= pad.Load64(uint32(i.Imm))
		case VM_IMULH_R:
			f.R[i.Dst], _ = bits.Mul64(f.R[i.Dst], f.R[i.Src])
		case VM_IMULH_M:
			f.R[i.Dst], _ = bits.Mul64(f.R[i.Dst], pad.Load64(i.getScratchpadAddress(f.R[i.Src])))
		case VM_IMULH_MZ:
			f.R[i.Dst], _ = bits.Mul64(f.R[i.Dst], pad.Load64(uint32(i.Imm)))
		case VM_ISMULH_R:
			f.R[i.Dst] = smulh(int64(f.R[i.Dst]), int64(f.R[i.Src]))
		case VM_ISMULH_M:
			f.R[i.Dst] = smulh(int64(f.R[i.Dst]), int64(pad.Load64(i.getScratchpadAddress(f.R[i.Src]))))
		case VM_ISMULH_MZ:
			f.R[i.Dst] = smulh(int64(f.R[i.Dst]), int64(pad.Load64(uint32(i.Imm))))
		case VM_INEG_R:
			f.R[i.Dst] = -f.R[i.Dst]
		case VM_IXOR_R:
			f.R[i.Dst] ^= f.R[i.Src]
		case VM_IXOR_I:
			f.R[i.Dst] ^= i.Imm
		case VM_IXOR_M:
			f.R[i.Dst] ^= pad.Load64(i.getScratchpadAddress(f.R[i.Src]))
		case VM_IXOR_MZ:
			f.R[i.Dst] ^= pad.Load64(uint32(i.Imm))
		case VM_IROR_R:
			f.R[i.Dst] = bits.RotateLeft64(f.R[i.Dst], 0-int(f.R[i.Src]&63))
		case VM_IROR_I:
			//todo: can merge into VM_IROL_I
			f.R[i.Dst] = bits.RotateLeft64(f.R[i.Dst], 0-int(i.Imm&63))
		case VM_IROL_R:
			f.R[i.Dst] = bits.RotateLeft64(f.R[i.Dst], int(f.R[i.Src]&63))
		case VM_IROL_I:
			f.R[i.Dst] = bits.RotateLeft64(f.R[i.Dst], int(i.Imm&63))
		case VM_ISWAP_R:
			f.R[i.Dst], f.R[i.Src] = f.R[i.Src], f.R[i.Dst]
		case VM_FSWAP_RF:
			f.F[i.Dst][HIGH], f.F[i.Dst][LOW] = f.F[i.Dst][LOW], f.F[i.Dst][HIGH]
		case VM_FSWAP_RE:
			f.E[i.Dst][HIGH], f.E[i.Dst][LOW] = f.E[i.Dst][LOW], f.E[i.Dst][HIGH]
		case VM_FADD_R:
			f.F[i.Dst][LOW] += f.A[i.Src][LOW]
			f.F[i.Dst][HIGH] += f.A[i.Src][HIGH]
		case VM_FADD_M:
			lo, hi := pad.Load32F(i.getScratchpadAddress(f.R[i.Src]))
			f.F[i.Dst][LOW] += lo
			f.F[i.Dst][HIGH] += hi
		case VM_FSUB_R:
			f.F[i.Dst][LOW] -= f.A[i.Src][LOW]
			f.F[i.Dst][HIGH] -= f.A[i.Src][HIGH]
		case VM_FSUB_M:
			lo, hi := pad.Load32F(i.getScratchpadAddress(f.R[i.Src]))
			f.F[i.Dst][LOW] -= lo
			f.F[i.Dst][HIGH] -= hi
		case VM_FSCAL_R:
			// no dependent on rounding modes
			f.F[i.Dst][LOW] = softfloat.ScaleNegate(f.F[i.Dst][LOW])
			f.F[i.Dst][HIGH] = softfloat.ScaleNegate(f.F[i.Dst][HIGH])
		case VM_FMUL_R:
			f.E[i.Dst][LOW] *= f.A[i.Src][LOW]
			f.E[i.Dst][HIGH] *= f.A[i.Src][HIGH]
		case VM_FDIV_M:
			lo, hi := pad.Load32F(i.getScratchpadAddress(f.R[i.Src]))
			f.E[i.Dst][LOW] /= softfloat.MaskRegisterExponentMantissa(lo, eMask[LOW])
			f.E[i.Dst][HIGH] /= softfloat.MaskRegisterExponentMantissa(hi, eMask[HIGH])
		case VM_FSQRT_R:
			f.E[i.Dst][LOW] = math.Sqrt(f.E[i.Dst][LOW])
			f.E[i.Dst][HIGH] = math.Sqrt(f.E[i.Dst][HIGH])
		case VM_CBRANCH:
			f.R[i.Src] += i.Imm
			if (f.R[i.Src] & uint64(i.MemMask)) == 0 {
				pc = i.jumpTarget()
			}
		case VM_CFROUND:
			tmp := (bits.RotateLeft64(f.R[i.Src], 0-int(i.Imm))) % 4 // rotate right
			f.SetRoundingMode(softfloat.RoundingMode(tmp))
		case VM_ISTORE:
			pad.Store64(i.getScratchpadAddress(f.R[i.Dst]), f.R[i.Src])
		}
	}
}

type ByteCodeInstructionOp int

const (
	VM_NOP = ByteCodeInstructionOp(iota)
	VM_IADD_RS
	VM_IADD_M
	VM_IADD_MZ
	VM_ISUB_R
	VM_ISUB_I
	VM_ISUB_M
	VM_ISUB_MZ
	VM_IMUL_R
	VM_IMUL_I
	VM_IMUL_M
	VM_IMUL_MZ
	VM_IMULH_R
	VM_IMULH_M
	VM_IMULH_MZ
	VM_ISMULH_R
	VM_ISMULH_M
	VM_ISMULH_MZ
	VM_INEG_R
	VM_IXOR_R
	VM_IXOR_I
	VM_IXOR_M
	VM_IXOR_MZ
	VM_IROR_R
	VM_IROR_I
	VM_IROL_R
	VM_IROL_I
	VM_ISWAP_R
	VM_FSWAP_RF
	VM_FSWAP_RE
	VM_FADD_R
	VM_FADD_M
	VM_FSUB_R
	VM_FSUB_M
	VM_FSCAL_R
	VM_FMUL_R
	VM_FDIV_M
	VM_FSQRT_R
	VM_CBRANCH
	VM_CFROUND
	VM_ISTORE
)

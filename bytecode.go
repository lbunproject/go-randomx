package randomx

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v2/asm"
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

func (c *ByteCode) Execute(f RegisterFile, pad *ScratchPad, eMask [2]uint64) RegisterFile {
	for pc := 0; pc < RANDOMX_PROGRAM_SIZE; pc++ {
		i := &c[pc]
		switch i.Opcode {
		case VM_IADD_RS:
			f.r[i.Dst] += (f.r[i.Src] << i.ImmB) + i.Imm
		case VM_IADD_M:
			f.r[i.Dst] += pad.Load64(i.getScratchpadAddress(f.r[i.Src]))
		case VM_IADD_MZ:
			f.r[i.Dst] += pad.Load64(uint32(i.Imm))
		case VM_ISUB_R:
			f.r[i.Dst] -= f.r[i.Src]
		case VM_ISUB_I:
			f.r[i.Dst] -= i.Imm
		case VM_ISUB_M:
			f.r[i.Dst] -= pad.Load64(i.getScratchpadAddress(f.r[i.Src]))
		case VM_ISUB_MZ:
			f.r[i.Dst] -= pad.Load64(uint32(i.Imm))
		case VM_IMUL_R:
			f.r[i.Dst] *= f.r[i.Src]
		case VM_IMUL_I:
			// also handles imul_rcp
			f.r[i.Dst] *= i.Imm
		case VM_IMUL_M:
			f.r[i.Dst] *= pad.Load64(i.getScratchpadAddress(f.r[i.Src]))
		case VM_IMUL_MZ:
			f.r[i.Dst] *= pad.Load64(uint32(i.Imm))
		case VM_IMULH_R:
			f.r[i.Dst], _ = bits.Mul64(f.r[i.Dst], f.r[i.Src])
		case VM_IMULH_M:
			f.r[i.Dst], _ = bits.Mul64(f.r[i.Dst], pad.Load64(i.getScratchpadAddress(f.r[i.Src])))
		case VM_IMULH_MZ:
			f.r[i.Dst], _ = bits.Mul64(f.r[i.Dst], pad.Load64(uint32(i.Imm)))
		case VM_ISMULH_R:
			f.r[i.Dst] = smulh(int64(f.r[i.Dst]), int64(f.r[i.Src]))
		case VM_ISMULH_M:
			f.r[i.Dst] = smulh(int64(f.r[i.Dst]), int64(pad.Load64(i.getScratchpadAddress(f.r[i.Src]))))
		case VM_ISMULH_MZ:
			f.r[i.Dst] = smulh(int64(f.r[i.Dst]), int64(pad.Load64(uint32(i.Imm))))
		case VM_INEG_R:
			//f.r[i.Dst] = (^(f.r[i.Dst])) + 1 // 2's complement negative
			f.r[i.Dst] = -f.r[i.Dst]
		case VM_IXOR_R:
			f.r[i.Dst] ^= f.r[i.Src]
		case VM_IXOR_I:
			f.r[i.Dst] ^= i.Imm
		case VM_IXOR_M:
			f.r[i.Dst] ^= pad.Load64(i.getScratchpadAddress(f.r[i.Src]))
		case VM_IXOR_MZ:
			f.r[i.Dst] ^= pad.Load64(uint32(i.Imm))
		case VM_IROR_R:
			f.r[i.Dst] = bits.RotateLeft64(f.r[i.Dst], 0-int(f.r[i.Src]&63))
		case VM_IROR_I:
			//todo: can merge into VM_IROL_I
			f.r[i.Dst] = bits.RotateLeft64(f.r[i.Dst], 0-int(i.Imm&63))
		case VM_IROL_R:
			f.r[i.Dst] = bits.RotateLeft64(f.r[i.Dst], int(f.r[i.Src]&63))
		case VM_IROL_I:
			f.r[i.Dst] = bits.RotateLeft64(f.r[i.Dst], int(i.Imm&63))
		case VM_ISWAP_R:
			f.r[i.Dst], f.r[i.Src] = f.r[i.Src], f.r[i.Dst]
		case VM_FSWAP_RF:
			f.f[i.Dst][HIGH], f.f[i.Dst][LOW] = f.f[i.Dst][LOW], f.f[i.Dst][HIGH]
		case VM_FSWAP_RE:
			f.e[i.Dst][HIGH], f.e[i.Dst][LOW] = f.e[i.Dst][LOW], f.e[i.Dst][HIGH]
		case VM_FADD_R:
			f.f[i.Dst][LOW] += f.a[i.Src][LOW]
			f.f[i.Dst][HIGH] += f.a[i.Src][HIGH]
		case VM_FADD_M:
			lo, hi := pad.Load32F(i.getScratchpadAddress(f.r[i.Src]))
			f.f[i.Dst][LOW] += lo
			f.f[i.Dst][HIGH] += hi
		case VM_FSUB_R:
			f.f[i.Dst][LOW] -= f.a[i.Src][LOW]
			f.f[i.Dst][HIGH] -= f.a[i.Src][HIGH]
		case VM_FSUB_M:
			lo, hi := pad.Load32F(i.getScratchpadAddress(f.r[i.Src]))
			f.f[i.Dst][LOW] -= lo
			f.f[i.Dst][HIGH] -= hi
		case VM_FSCAL_R:
			// no dependent on rounding modes
			f.f[i.Dst][LOW] = math.Float64frombits(math.Float64bits(f.f[i.Dst][LOW]) ^ 0x80F0000000000000)
			f.f[i.Dst][HIGH] = math.Float64frombits(math.Float64bits(f.f[i.Dst][HIGH]) ^ 0x80F0000000000000)
		case VM_FMUL_R:
			f.e[i.Dst][LOW] *= f.a[i.Src][LOW]
			f.e[i.Dst][HIGH] *= f.a[i.Src][HIGH]
		case VM_FDIV_M:
			lo, hi := pad.Load32F(i.getScratchpadAddress(f.r[i.Src]))
			f.e[i.Dst][LOW] /= MaskRegisterExponentMantissa(lo, eMask[LOW])
			f.e[i.Dst][HIGH] /= MaskRegisterExponentMantissa(hi, eMask[HIGH])
		case VM_FSQRT_R:
			f.e[i.Dst][LOW] = math.Sqrt(f.e[i.Dst][LOW])
			f.e[i.Dst][HIGH] = math.Sqrt(f.e[i.Dst][HIGH])
		case VM_CBRANCH:
			f.r[i.Src] += i.Imm
			if (f.r[i.Src] & uint64(i.MemMask)) == 0 {
				pc = i.jumpTarget()
			}
		case VM_CFROUND:
			tmp := (bits.RotateLeft64(f.r[i.Src], 0-int(i.Imm))) % 4 // rotate right
			asm.SetRoundingMode(asm.RoundingMode(tmp))
		case VM_ISTORE:
			pad.Store64(i.getScratchpadAddress(f.r[i.Dst]), f.r[i.Src])
		case VM_NOP: // we do nothing
		}
	}
	return f
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

package randomx

import (
	"encoding/binary"
	"git.gammaspectra.live/P2Pool/go-randomx/v2/asm"
	"math"
	"math/bits"
)

type ByteCodeInstruction struct {
	dst, src   byte
	idst, isrc *uint64
	fdst, fsrc *[2]float64
	imm        uint64
	simm       int64
	Opcode     ByteCodeInstructionOp
	target     int16
	shift      uint8
	memMask    uint32
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

func (i ByteCodeInstruction) getScratchpadSrcAddress() uint64 {
	return (*i.isrc + i.imm) & uint64(i.memMask)
}

func (i ByteCodeInstruction) getScratchpadZeroAddress() uint64 {
	return i.imm & uint64(i.memMask)
}

func (i ByteCodeInstruction) getScratchpadDestAddress() uint64 {
	return (*i.idst + i.imm) & uint64(i.memMask)
}

type ByteCode [RANDOMX_PROGRAM_SIZE]ByteCodeInstruction

func (c *ByteCode) Interpret(vm *VM) {
	for pc := 0; pc < RANDOMX_PROGRAM_SIZE; pc++ {
		ibc := c[pc]
		switch ibc.Opcode {
		case VM_IADD_RS:
			*ibc.idst += (*ibc.isrc << ibc.shift) + ibc.imm
		case VM_IADD_M:
			*ibc.idst += vm.Load64(ibc.getScratchpadSrcAddress())
		case VM_IADD_MZ:
			*ibc.idst += vm.Load64(ibc.getScratchpadZeroAddress())
		case VM_ISUB_R:
			*ibc.idst -= *ibc.isrc
		case VM_ISUB_M:
			*ibc.idst -= vm.Load64(ibc.getScratchpadSrcAddress())
		case VM_ISUB_MZ:
			*ibc.idst -= vm.Load64(ibc.getScratchpadZeroAddress())
		case VM_IMUL_R:
			// also handles imul_rcp
			*ibc.idst *= *ibc.isrc
		case VM_IMUL_M:
			*ibc.idst *= vm.Load64(ibc.getScratchpadSrcAddress())
		case VM_IMUL_MZ:
			*ibc.idst *= vm.Load64(ibc.getScratchpadZeroAddress())
		case VM_IMULH_R:
			*ibc.idst, _ = bits.Mul64(*ibc.idst, *ibc.isrc)
		case VM_IMULH_M:
			*ibc.idst, _ = bits.Mul64(*ibc.idst, vm.Load64(ibc.getScratchpadSrcAddress()))
		case VM_IMULH_MZ:
			*ibc.idst, _ = bits.Mul64(*ibc.idst, vm.Load64(ibc.getScratchpadZeroAddress()))
		case VM_ISMULH_R:
			*ibc.idst = smulh(int64(*ibc.idst), int64(*ibc.isrc))
		case VM_ISMULH_M:
			*ibc.idst = smulh(int64(*ibc.idst), int64(vm.Load64(ibc.getScratchpadSrcAddress())))
		case VM_ISMULH_MZ:
			*ibc.idst = smulh(int64(*ibc.idst), int64(vm.Load64(ibc.getScratchpadZeroAddress())))
		case VM_INEG_R:
			*ibc.idst = (^(*ibc.idst)) + 1 // 2's complement negative
		case VM_IXOR_R:
			*ibc.idst ^= *ibc.isrc
		case VM_IXOR_M:
			*ibc.idst ^= vm.Load64(ibc.getScratchpadSrcAddress())
		case VM_IXOR_MZ:
			*ibc.idst ^= vm.Load64(ibc.getScratchpadZeroAddress())
		case VM_IROR_R:
			*ibc.idst = bits.RotateLeft64(*ibc.idst, 0-int(*ibc.isrc&63))
		case VM_IROL_R:
			*ibc.idst = bits.RotateLeft64(*ibc.idst, int(*ibc.isrc&63))
		case VM_ISWAP_R:
			*ibc.idst, *ibc.isrc = *ibc.isrc, *ibc.idst
		case VM_FSWAP_R:
			ibc.fdst[HIGH], ibc.fdst[LOW] = ibc.fdst[LOW], ibc.fdst[HIGH]
		case VM_FADD_R:
			ibc.fdst[LOW] += ibc.fsrc[LOW]
			ibc.fdst[HIGH] += ibc.fsrc[HIGH]
		case VM_FADD_M:
			lo, hi := vm.Load32F(ibc.getScratchpadSrcAddress())
			ibc.fdst[LOW] += lo
			ibc.fdst[HIGH] += hi
		case VM_FSUB_R:
			ibc.fdst[LOW] -= ibc.fsrc[LOW]
			ibc.fdst[HIGH] -= ibc.fsrc[HIGH]
		case VM_FSUB_M:
			lo, hi := vm.Load32F(ibc.getScratchpadSrcAddress())
			ibc.fdst[LOW] -= lo
			ibc.fdst[HIGH] -= hi
		case VM_FSCAL_R:
			// no dependent on rounding modes
			ibc.fdst[LOW] = math.Float64frombits(math.Float64bits(ibc.fdst[LOW]) ^ 0x80F0000000000000)
			ibc.fdst[HIGH] = math.Float64frombits(math.Float64bits(ibc.fdst[HIGH]) ^ 0x80F0000000000000)
		case VM_FMUL_R:
			ibc.fdst[LOW] *= ibc.fsrc[LOW]
			ibc.fdst[HIGH] *= ibc.fsrc[HIGH]
		case VM_FDIV_M:
			lo, hi := vm.Load32F(ibc.getScratchpadSrcAddress())
			ibc.fdst[LOW] /= MaskRegisterExponentMantissa(lo, vm.config.eMask[LOW])
			ibc.fdst[HIGH] /= MaskRegisterExponentMantissa(hi, vm.config.eMask[HIGH])
		case VM_FSQRT_R:
			ibc.fdst[LOW] = math.Sqrt(ibc.fdst[LOW])
			ibc.fdst[HIGH] = math.Sqrt(ibc.fdst[HIGH])
		case VM_CBRANCH:
			*ibc.isrc += ibc.imm
			if (*ibc.isrc & uint64(ibc.memMask)) == 0 {
				pc = int(ibc.target)
			}
		case VM_CFROUND:
			tmp := (bits.RotateLeft64(*ibc.isrc, 0-int(ibc.imm))) % 4 // rotate right
			asm.SetRoundingMode(asm.RoundingMode(tmp))
		case VM_ISTORE:
			binary.LittleEndian.PutUint64(vm.ScratchPad[(*ibc.idst+ibc.imm)&uint64(ibc.memMask):], *ibc.isrc)
		case VM_NOP: // we do nothing
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
	VM_ISUB_M
	VM_ISUB_MZ
	VM_IMUL_R
	VM_IMUL_M
	VM_IMUL_MZ
	VM_IMULH_R
	VM_IMULH_M
	VM_IMULH_MZ
	VM_ISMULH_R
	VM_ISMULH_M
	VM_ISMULH_MZ
	VM_IMUL_RCP
	VM_INEG_R
	VM_IXOR_R
	VM_IXOR_M
	VM_IXOR_MZ
	VM_IROR_R
	VM_IROL_R
	VM_ISWAP_R
	VM_FSWAP_R
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

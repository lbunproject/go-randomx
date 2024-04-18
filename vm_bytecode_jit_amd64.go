//go:build unix && amd64 && !disable_jit && !purego

package randomx

import (
	"encoding/binary"
	"math/bits"
	"unsafe"
)

//go:noescape
func vm_run(rf *RegisterFile, pad *ScratchPad, eMask [2]uint64, jmp uintptr)

func (f VMProgramFunc) Execute(rf *RegisterFile, pad *ScratchPad, eMask [2]uint64) {
	if f == nil {
		panic("program is nil")
	}

	jmpPtr := uintptr(unsafe.Pointer(unsafe.SliceData(f)))
	vm_run(rf, pad, eMask, jmpPtr)
}

func (c *ByteCode) generateCode(program []byte) {
	program = program[:0]

	var instructionOffsets [RANDOMX_PROGRAM_SIZE]int32
	var codePos int32

	for ix := range c {
		instructionOffsets[ix] = codePos
		curLen := len(program)

		instr := &c[ix]
		switch instr.Opcode {

		case VM_IADD_RS:
			program = append(program, REX_LEA...)
			if instr.Dst == RegisterNeedsDisplacement {
				program = append(program, 0xac)
			} else {
				program = append(program, 0x04+8*instr.Dst)
			}
			program = append(program, genSIB(int(instr.ImmB), int(instr.Src), int(instr.Dst)))
			if instr.Dst == RegisterNeedsDisplacement {
				program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))
			}

		case VM_IADD_M:
			program = genAddressReg(program, instr, true)
			program = append(program, REX_ADD_RM...)
			program = append(program, 0x04+8*instr.Dst)
			program = append(program, 0x06)
		case VM_IADD_MZ:
			program = append(program, REX_ADD_RM...)
			program = append(program, 0x86+8*instr.Dst)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))

		case VM_ISUB_R:
			program = append(program, REX_SUB_RR...)
			program = append(program, 0xc0+8*instr.Dst+instr.Src)
		case VM_ISUB_I:
			program = append(program, REX_81...)
			program = append(program, 0xe8+instr.Dst)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))

		case VM_ISUB_M:
			program = genAddressReg(program, instr, true)
			program = append(program, REX_SUB_RM...)
			program = append(program, 0x04+8*instr.Dst)
			program = append(program, 0x06)
		case VM_ISUB_MZ:
			program = append(program, REX_SUB_RM...)
			program = append(program, 0x86+8*instr.Dst)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))

		case VM_IMUL_R:
			program = append(program, REX_IMUL_RR...)
			program = append(program, 0xc0+8*instr.Dst+instr.Src)
		case VM_IMUL_I:
			// also handles imul_rcp, with 64-bit special
			if bits.Len64(instr.Imm) > 32 {
				program = append(program, MOV_RAX_I...)
				program = binary.LittleEndian.AppendUint64(program, instr.Imm)
				program = append(program, REX_IMUL_RM...)
				program = append(program, 0xc0+8*instr.Dst)
			} else {
				program = append(program, REX_IMUL_RRI...)
				program = append(program, 0xc0+9*instr.Dst)
				program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))
			}

		case VM_IMUL_M:
			program = genAddressReg(program, instr, true)
			program = append(program, REX_IMUL_RM...)
			program = append(program, 0x04+8*instr.Dst)
			program = append(program, 0x06)
		case VM_IMUL_MZ:
			program = append(program, REX_IMUL_RM...)
			program = append(program, 0x86+8*instr.Dst)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))

		case VM_IMULH_R:
			program = append(program, REX_MOV_RR64...)
			program = append(program, 0xc0+instr.Dst)
			program = append(program, REX_MUL_R...)
			program = append(program, 0xe0+instr.Src)
			program = append(program, REX_MOV_R64R...)
			program = append(program, 0xc2+8*instr.Dst)

		case VM_IMULH_M:
			program = genAddressReg(program, instr, false)
			program = append(program, REX_MOV_RR64...)
			program = append(program, 0xc0+instr.Dst)
			program = append(program, REX_MUL_MEM...)
			program = append(program, REX_MOV_R64R...)
			program = append(program, 0xc2+8*instr.Dst)
		case VM_IMULH_MZ:
			program = append(program, REX_MOV_RR64...)
			program = append(program, 0xc0+instr.Dst)
			program = append(program, REX_MUL_M...)
			program = append(program, 0xa6)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))
			program = append(program, REX_MOV_R64R...)
			program = append(program, 0xc2+8*instr.Dst)

		case VM_ISMULH_R:
			program = append(program, REX_MOV_RR64...)
			program = append(program, 0xc0+instr.Dst)
			program = append(program, REX_MUL_R...)
			program = append(program, 0xe8+instr.Src)
			program = append(program, REX_MOV_R64R...)
			program = append(program, 0xc2+8*instr.Dst)

		case VM_ISMULH_M:
			program = genAddressReg(program, instr, false)
			program = append(program, REX_MOV_RR64...)
			program = append(program, 0xc0+instr.Dst)
			program = append(program, REX_IMUL_MEM...)
			program = append(program, REX_MOV_R64R...)
			program = append(program, 0xc2+8*instr.Dst)
		case VM_ISMULH_MZ:
			program = append(program, REX_MOV_RR64...)
			program = append(program, 0xc0+instr.Dst)
			program = append(program, REX_MUL_M...)
			program = append(program, 0xae)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))
			program = append(program, REX_MOV_R64R...)
			program = append(program, 0xc2+8*instr.Dst)

		case VM_INEG_R:
			program = append(program, REX_NEG...)
			program = append(program, 0xd8+instr.Dst)

		case VM_IXOR_R:
			program = append(program, REX_XOR_RR...)
			program = append(program, 0xc0+8*instr.Dst+instr.Src)
		case VM_IXOR_I:
			program = append(program, REX_XOR_RI...)
			program = append(program, 0xf0+instr.Dst)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))

		case VM_IXOR_M:
			program = genAddressReg(program, instr, true)
			program = append(program, REX_XOR_RM...)
			program = append(program, 0x04+8*instr.Dst)
			program = append(program, 0x06)
		case VM_IXOR_MZ:
			program = append(program, REX_XOR_RM...)
			program = append(program, 0x86+8*instr.Dst)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))

		case VM_IROR_R:
			program = append(program, REX_MOV_RR...)
			program = append(program, 0xc8+instr.Src)
			program = append(program, REX_ROT_CL...)
			program = append(program, 0xc8+instr.Dst)
		case VM_IROR_I:
			program = append(program, REX_ROT_I8...)
			program = append(program, 0xc8+instr.Dst)
			program = append(program, byte(instr.Imm&63))

		case VM_IROL_R:
			program = append(program, REX_MOV_RR...)
			program = append(program, 0xc8+instr.Src)
			program = append(program, REX_ROT_CL...)
			program = append(program, 0xc0+instr.Dst)
		case VM_IROL_I:
			program = append(program, REX_ROT_I8...)
			program = append(program, 0xc0+instr.Dst)
			program = append(program, byte(instr.Imm&63))

		case VM_ISWAP_R:
			program = append(program, REX_XCHG...)
			program = append(program, 0xc0+instr.Src+8*instr.Dst)

		case VM_FSWAP_RF:
			program = append(program, SHUFPD...)
			program = append(program, 0xc0+9*instr.Dst)
			program = append(program, 1)
		case VM_FSWAP_RE:
			program = append(program, SHUFPD...)
			program = append(program, 0xc0+9*(instr.Dst+RegistersCountFloat))
			program = append(program, 1)

		case VM_FADD_R:
			program = append(program, REX_ADDPD...)
			program = append(program, 0xc0+instr.Src+8*instr.Dst)

		case VM_FADD_M:
			program = genAddressReg(program, instr, true)
			program = append(program, REX_CVTDQ2PD_XMM12...)
			program = append(program, REX_ADDPD...)
			program = append(program, 0xc4+8*instr.Dst)

		case VM_FSUB_R:
			program = append(program, REX_SUBPD...)
			program = append(program, 0xc0+instr.Src+8*instr.Dst)

		case VM_FSUB_M:
			program = genAddressReg(program, instr, true)
			program = append(program, REX_CVTDQ2PD_XMM12...)
			program = append(program, REX_SUBPD...)
			program = append(program, 0xc4+8*instr.Dst)

		case VM_FSCAL_R:
			program = append(program, REX_XORPS...)
			program = append(program, 0xc7+8*instr.Dst)

		case VM_FMUL_R:
			program = append(program, REX_MULPD...)
			program = append(program, 0xe0+instr.Src+8*instr.Dst)

		case VM_FDIV_M:
			program = genAddressReg(program, instr, true)
			program = append(program, REX_CVTDQ2PD_XMM12...)
			program = append(program, REX_ANDPS_XMM12...)
			program = append(program, REX_DIVPD...)
			program = append(program, 0xe4+8*instr.Dst)

		case VM_FSQRT_R:
			program = append(program, SQRTPD...)
			program = append(program, 0xe4+9*instr.Dst)

		case VM_CFROUND:
			program = append(program, REX_MOV_RR64...)
			program = append(program, 0xc0+instr.Src)
			rotate := byte((13 - instr.Imm) & 63)
			if rotate != 0 {
				program = append(program, ROL_RAX...)
				program = append(program, rotate)
			}
			program = append(program, AND_OR_MOV_LDMXCSR...)
		case VM_CBRANCH:
			reg := instr.Dst
			target := instr.jumpTarget() + 1

			jmpOffset := instructionOffsets[target] - (codePos + 16)

			if BranchesWithin32B {
				branchBegin := uint32(codePos + 7)
				branchEnd := branchBegin
				if jmpOffset >= -128 {
					branchEnd += 9
				} else {
					branchEnd += 13
				}
				// If the jump crosses or touches 32-byte boundary, align it
				if (branchBegin ^ branchEnd) >= 32 {
					alignmentSize := 32 - (branchBegin & 31)
					alignmentSize -= alignmentSize

					program = append(program, JMP_ALIGN_PREFIX[alignmentSize]...)
				}
			}
			program = append(program, REX_ADD_I...)
			program = append(program, 0xc0+reg)
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))

			program = append(program, REX_TEST...)
			program = append(program, 0xc0+reg)
			program = binary.LittleEndian.AppendUint32(program, instr.MemMask)

			if jmpOffset >= -128 {
				program = append(program, JZ_SHORT)
				program = append(program, byte(jmpOffset))
			} else {
				program = append(program, JZ...)
				program = binary.LittleEndian.AppendUint32(program, uint32(jmpOffset-4))
			}

		case VM_ISTORE:
			//genAddressRegDst
			program = append(program, LEA_32...)
			program = append(program, 0x80+instr.Dst)
			if instr.Dst == RegisterNeedsSib {
				program = append(program, 0x24)
			}
			program = binary.LittleEndian.AppendUint32(program, uint32(instr.Imm))
			program = append(program, AND_EAX_I)
			program = binary.LittleEndian.AppendUint32(program, instr.MemMask)

			program = append(program, REX_MOV_MR...)
			program = append(program, 0x04+8*instr.Src)
			program = append(program, 0x06)
		case VM_NOP:
			program = append(program, NOP1...)
		}

		codePos += int32(len(program) - curLen)
	}
	program = append(program, RET)
}

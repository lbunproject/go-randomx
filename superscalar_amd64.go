//go:build unix && amd64 && !disable_jit && !purego

package randomx

import (
	"encoding/binary"
)

var REX_SUB_RR = []byte{0x4d, 0x2b}
var REX_MOV_RR64 = []byte{0x49, 0x8b}
var REX_MOV_R64R = []byte{0x4c, 0x8b}
var REX_IMUL_RR = []byte{0x4d, 0x0f, 0xaf}
var REX_IMUL_RM = []byte{0x4c, 0x0f, 0xaf}
var REX_MUL_R = []byte{0x49, 0xf7}
var REX_81 = []byte{0x49, 0x81}

var MOV_RAX_I = []byte{0x48, 0xb8}
var REX_LEA = []byte{0x4f, 0x8d}
var REX_XOR_RR = []byte{0x4D, 0x33}
var REX_XOR_RI = []byte{0x49, 0x81}
var REX_ROT_I8 = []byte{0x49, 0xc1}

func genSIB(scale, index, base int) byte {
	return byte((scale << 6) | (index << 3) | base)
}

/*
push rbp
push rbx
push rsi
push r12
push r13
push r14
push r15
mov    rbp,rsp
sub    rsp,(0x8*7)

mov    rsi, rax; # register dataset

prefetchnta byte ptr [rsi]

mov r8, qword ptr [rsi+0]
mov r9, qword ptr [rsi+8]
mov r10, qword ptr [rsi+16]
mov r11, qword ptr [rsi+24]
mov r12, qword ptr [rsi+32]
mov r13, qword ptr [rsi+40]
mov r14, qword ptr [rsi+48]
mov r15, qword ptr [rsi+56]
*/
var codeInitBlock = []byte{0x55, 0x53, 0x56, 0x41, 0x54, 0x41, 0x55, 0x41, 0x56, 0x41, 0x57, 0x48, 0x89, 0xE5, 0x48, 0x83, 0xEC, 0x38, 0x48, 0x89, 0xC6, 0x0F, 0x18, 0x06, 0x4C, 0x8B, 0x06, 0x4C, 0x8B, 0x4E, 0x08, 0x4C, 0x8B, 0x56, 0x10, 0x4C, 0x8B, 0x5E, 0x18, 0x4C, 0x8B, 0x66, 0x20, 0x4C, 0x8B, 0x6E, 0x28, 0x4C, 0x8B, 0x76, 0x30, 0x4C, 0x8B, 0x7E, 0x38}

/*
prefetchw byte ptr [rsi]

mov qword ptr [rsi+0], r8
mov qword ptr [rsi+8], r9
mov qword ptr [rsi+16], r10
mov qword ptr [rsi+24], r11
mov qword ptr [rsi+32], r12
mov qword ptr [rsi+40], r13
mov qword ptr [rsi+48], r14
mov qword ptr [rsi+56], r15

add    rsp,(0x8*7)
pop r15
pop r14
pop r13
pop r12
pop rsi
pop rbx
pop rbp
ret
*/
var codeRetBlock = []byte{0x0F, 0x0D, 0x0E, 0x4C, 0x89, 0x06, 0x4C, 0x89, 0x4E, 0x08, 0x4C, 0x89, 0x56, 0x10, 0x4C, 0x89, 0x5E, 0x18, 0x4C, 0x89, 0x66, 0x20, 0x4C, 0x89, 0x6E, 0x28, 0x4C, 0x89, 0x76, 0x30, 0x4C, 0x89, 0x7E, 0x38, 0x48, 0x83, 0xC4, 0x38, 0x41, 0x5F, 0x41, 0x5E, 0x41, 0x5D, 0x41, 0x5C, 0x5E, 0x5B, 0x5D, 0xC3}

// generateSuperscalarCode
func generateSuperscalarCode(scalarProgram SuperScalarProgram) ProgramFunc {

	var program []byte

	program = append(program, codeInitBlock...)

	p := scalarProgram.Program()
	for i := range p {
		instr := &p[i]

		dst := instr.Dst_Reg % RegistersCount
		src := instr.Src_Reg % RegistersCount

		switch instr.Opcode {
		case S_ISUB_R:
			program = append(program, REX_SUB_RR...)
			program = append(program, byte(0xc0+8*dst+src))
		case S_IXOR_R:
			program = append(program, REX_XOR_RR...)
			program = append(program, byte(0xc0+8*dst+src))
		case S_IADD_RS:
			program = append(program, REX_LEA...)
			program = append(program,
				byte(0x04+8*dst),
				genSIB(int(instr.Imm32), src, dst),
			)
		case S_IMUL_R:
			program = append(program, REX_IMUL_RR...)
			program = append(program, byte(0xc0+8*dst+src))
		case S_IROR_C:
			program = append(program, REX_ROT_I8...)
			program = append(program,
				byte(0xc8+dst),
				byte(instr.Imm32&63),
			)

		case S_IADD_C7, S_IADD_C8, S_IADD_C9:
			program = append(program, REX_81...)
			program = append(program, byte(0xc0+dst))
			program = binary.LittleEndian.AppendUint32(program, instr.Imm32)
			//TODO: align NOP on C8/C9
		case S_IXOR_C7, S_IXOR_C8, S_IXOR_C9:
			program = append(program, REX_XOR_RI...)
			program = append(program, byte(0xf0+dst))
			program = binary.LittleEndian.AppendUint32(program, instr.Imm32)
			//TODO: align NOP on C8/C9

		case S_IMULH_R:
			program = append(program, REX_MOV_RR64...)
			program = append(program, byte(0xc0+dst))
			program = append(program, REX_MUL_R...)
			program = append(program, byte(0xe0+src))
			program = append(program, REX_MOV_R64R...)
			program = append(program, byte(0xc2+8*dst))
		case S_ISMULH_R:
			program = append(program, REX_MOV_RR64...)
			program = append(program, byte(0xc0+dst))
			program = append(program, REX_MUL_R...)
			program = append(program, byte(0xe8+src))
			program = append(program, REX_MOV_R64R...)
			program = append(program, byte(0xc2+8*dst))
		case S_IMUL_RCP:
			program = append(program, MOV_RAX_I...)
			program = binary.LittleEndian.AppendUint64(program, randomx_reciprocal(instr.Imm32))
			program = append(program, REX_IMUL_RM...)
			program = append(program, byte(0xc0+8*instr.Dst_Reg))
		default:
			panic("unreachable")
		}
	}

	program = append(program, codeRetBlock...)

	return mapProgram(program)
}

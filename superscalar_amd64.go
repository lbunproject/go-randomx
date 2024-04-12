//go:build unix && amd64 && !disable_jit

package randomx

import (
	"encoding/binary"
)

const MaxRandomXInstrCodeSize = 32   //FDIV_M requires up to 32 bytes of x86 code
const MaxSuperscalarInstrSize = 14   //IMUL_RCP requires 14 bytes of x86 code
const SuperscalarProgramHeader = 128 //overhead per superscalar program
const CodeAlign = 4096               //align code size to a multiple of 4 KiB
const ReserveCodeSize = CodeAlign    //function prologue/epilogue + reserve

func alignSize[T ~uintptr | ~uint32 | ~uint64 | ~int64 | ~int32 | ~int](pos, align T) T {
	return ((pos-1)/align + 1) * align
}

var RandomXCodeSize = alignSize[uint64](ReserveCodeSize+MaxRandomXInstrCodeSize*RANDOMX_PROGRAM_SIZE, CodeAlign)
var SuperscalarSize = alignSize[uint64](ReserveCodeSize+(SuperscalarProgramHeader+MaxSuperscalarInstrSize*SuperscalarMaxSize)*RANDOMX_CACHE_ACCESSES, CodeAlign)

var CodeSize = uint32(RandomXCodeSize + SuperscalarSize)

var superScalarHashOffset = int32(RandomXCodeSize)

var REX_ADD_RR = []byte{0x4d, 0x03}
var REX_ADD_RM = []byte{0x4c, 0x03}
var REX_SUB_RR = []byte{0x4d, 0x2b}
var REX_SUB_RM = []byte{0x4c, 0x2b}
var REX_MOV_RR = []byte{0x41, 0x8b}
var REX_MOV_RR64 = []byte{0x49, 0x8b}
var REX_MOV_R64R = []byte{0x4c, 0x8b}
var REX_IMUL_RR = []byte{0x4d, 0x0f, 0xaf}
var REX_IMUL_RRI = []byte{0x4d, 0x69}
var REX_IMUL_RM = []byte{0x4c, 0x0f, 0xaf}
var REX_MUL_R = []byte{0x49, 0xf7}
var REX_MUL_M = []byte{0x48, 0xf7}
var REX_81 = []byte{0x49, 0x81}
var AND_EAX_I byte = 0x25

var MOV_EAX_I byte = 0xb8

var MOV_RAX_I = []byte{0x48, 0xb8}
var MOV_RCX_I = []byte{0x48, 0xb9}
var REX_LEA = []byte{0x4f, 0x8d}
var REX_MUL_MEM = []byte{0x48, 0xf7, 0x24, 0x0e}
var REX_IMUL_MEM = []byte{0x48, 0xf7, 0x2c, 0x0e}
var REX_SHR_RAX = []byte{0x48, 0xc1, 0xe8}
var RAX_ADD_SBB_1 = []byte{0x48, 0x83, 0xC0, 0x01, 0x48, 0x83, 0xD8, 0x00}
var MUL_RCX = []byte{0x48, 0xf7, 0xe1}
var REX_SHR_RDX = []byte{0x48, 0xc1, 0xea}
var REX_SH = []byte{0x49, 0xc1}
var MOV_RCX_RAX_SAR_RCX_63 = []byte{0x48, 0x89, 0xc1, 0x48, 0xc1, 0xf9, 0x3f}
var AND_ECX_I = []byte{0x81, 0xe1}
var ADD_RAX_RCX = []byte{0x48, 0x01, 0xC8}
var SAR_RAX_I8 = []byte{0x48, 0xC1, 0xF8}
var NEG_RAX = []byte{0x48, 0xF7, 0xD8}
var ADD_R_RAX = []byte{0x4C, 0x03}
var XOR_EAX_EAX = []byte{0x33, 0xC0}
var ADD_RDX_R = []byte{0x4c, 0x01}
var SUB_RDX_R = []byte{0x4c, 0x29}
var SAR_RDX_I8 = []byte{0x48, 0xC1, 0xFA}
var TEST_RDX_RDX = []byte{0x48, 0x85, 0xD2}
var SETS_AL_ADD_RDX_RAX = []byte{0x0F, 0x98, 0xC0, 0x48, 0x03, 0xD0}
var REX_NEG = []byte{0x49, 0xF7}
var REX_XOR_RR = []byte{0x4D, 0x33}
var REX_XOR_RI = []byte{0x49, 0x81}
var REX_XOR_RM = []byte{0x4c, 0x33}
var REX_ROT_CL = []byte{0x49, 0xd3}
var REX_ROT_I8 = []byte{0x49, 0xc1}
var SHUFPD = []byte{0x66, 0x0f, 0xc6}
var REX_ADDPD = []byte{0x66, 0x41, 0x0f, 0x58}
var REX_CVTDQ2PD_XMM12 = []byte{0xf3, 0x44, 0x0f, 0xe6, 0x24, 0x06}
var REX_SUBPD = []byte{0x66, 0x41, 0x0f, 0x5c}
var REX_XORPS = []byte{0x41, 0x0f, 0x57}
var REX_MULPD = []byte{0x66, 0x41, 0x0f, 0x59}
var REX_MAXPD = []byte{0x66, 0x41, 0x0f, 0x5f}
var REX_DIVPD = []byte{0x66, 0x41, 0x0f, 0x5e}
var SQRTPD = []byte{0x66, 0x0f, 0x51}
var AND_OR_MOV_LDMXCSR = []byte{0x25, 0x00, 0x60, 0x00, 0x00, 0x0D, 0xC0, 0x9F, 0x00, 0x00, 0x50, 0x0F, 0xAE, 0x14, 0x24, 0x58}
var ROL_RAX = []byte{0x48, 0xc1, 0xc0}
var XOR_ECX_ECX = []byte{0x33, 0xC9}
var REX_CMP_R32I = []byte{0x41, 0x81}
var REX_CMP_M32I = []byte{0x81, 0x3c, 0x06}
var MOVAPD = []byte{0x66, 0x0f, 0x29}
var REX_MOV_MR = []byte{0x4c, 0x89}
var REX_XOR_EAX = []byte{0x41, 0x33}
var SUB_EBX = []byte{0x83, 0xEB, 0x01}
var JNZ = []byte{0x0f, 0x85}
var JMP = 0xe9

var REX_XOR_RAX_R64 = []byte{0x49, 0x33}
var REX_XCHG = []byte{0x4d, 0x87}
var REX_ANDPS_XMM12 = []byte{0x45, 0x0F, 0x54, 0xE5, 0x45, 0x0F, 0x56, 0xE6}
var REX_PADD = []byte{0x66, 0x44, 0x0f}
var PADD_OPCODES = []byte{0xfc, 0xfd, 0xfe, 0xd4}
var CALL = 0xe8

var REX_ADD_I = []byte{0x49, 0x81}
var REX_TEST = []byte{0x49, 0xF7}
var JZ = []byte{0x0f, 0x84}
var JZ_SHORT = 0x74

var RET byte = 0xc3

var LEA_32 = []byte{0x41, 0x8d}
var MOVNTI = []byte{0x4c, 0x0f, 0xc3}
var ADD_EBX_I = []byte{0x81, 0xc3}

var NOP1 = []byte{0x90}
var NOP2 = []byte{0x66, 0x90}
var NOP3 = []byte{0x66, 0x66, 0x90}
var NOP4 = []byte{0x0F, 0x1F, 0x40, 0x00}
var NOP5 = []byte{0x0F, 0x1F, 0x44, 0x00, 0x00}
var NOP6 = []byte{0x66, 0x0F, 0x1F, 0x44, 0x00, 0x00}
var NOP7 = []byte{0x0F, 0x1F, 0x80, 0x00, 0x00, 0x00, 0x00}
var NOP8 = []byte{0x0F, 0x1F, 0x84, 0x00, 0x00, 0x00, 0x00, 0x00}

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

		dst := instr.Dst_Reg % REGISTERSCOUNT
		src := instr.Src_Reg % REGISTERSCOUNT

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

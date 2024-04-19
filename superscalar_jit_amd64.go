//go:build unix && amd64 && !disable_jit && !purego

package randomx

import (
	"encoding/binary"
	"runtime"
	"unsafe"
)

//go:noescape
func superscalar_run(rf, jmp uintptr)

func (f SuperScalarProgramFunc) Execute(rf uintptr) {
	if f == nil {
		panic("program is nil")
	}

	superscalar_run(rf, uintptr(unsafe.Pointer(unsafe.SliceData(f))))
	return

	var reservedStackHack [8 * 8]byte
	for i := range reservedStackHack {
		reservedStackHack[i] = uint8(i)
	}

	memoryPtr := &f
	fun := *(*func(v uintptr))(unsafe.Pointer(&memoryPtr))
	fun(rf)

	for i := range reservedStackHack {
		reservedStackHack[i] = uint8(-i)
	}
	runtime.KeepAlive(reservedStackHack)
}

// generateSuperscalarCode
func generateSuperscalarCode(scalarProgram SuperScalarProgram) SuperScalarProgramFunc {

	var program []byte

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

	program = append(program, RET)

	return mapProgram(program)
}

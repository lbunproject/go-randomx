package randomx

import "unsafe"

const RegistersCount = 8
const RegistersCountFloat = 4

const LOW = 0
const HIGH = 1

type RegisterLine [RegistersCount]uint64

type RegisterFile struct {
	R RegisterLine
	F [RegistersCountFloat][2]float64
	E [RegistersCountFloat][2]float64
	A [RegistersCountFloat][2]float64

	FPRC uint8
}

const RegisterFileSize = RegistersCount*8 + RegistersCountFloat*2*8*3

func (rf *RegisterFile) Memory() *[RegisterFileSize]byte {
	return (*[RegisterFileSize]byte)(unsafe.Pointer(rf))
}

type MemoryRegisters struct {
	mx, ma uint64
}

package randomx

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v2/asm"
	"git.gammaspectra.live/P2Pool/go-randomx/v2/softfloat"
)

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

	FPRC softfloat.RoundingMode
}

func (f *RegisterFile) SetRoundingMode(mode softfloat.RoundingMode) {
	if f.FPRC == mode {
		return
	}
	f.FPRC = mode
	asm.SetRoundingMode(mode)
}

type MemoryRegisters struct {
	mx, ma uint64
}

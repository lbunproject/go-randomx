package randomx

const RegistersCount = 8
const RegistersCountFloat = 4

const LOW = 0
const HIGH = 1

type RegisterLine [RegistersCount]uint64

type RegisterFile struct {
	r RegisterLine
	f [RegistersCountFloat][2]float64
	e [RegistersCountFloat][2]float64
	a [RegistersCountFloat][2]float64
}

type MemoryRegisters struct {
	mx, ma uint64
}

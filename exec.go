package randomx

import "git.gammaspectra.live/P2Pool/go-randomx/v3/internal/memory"

type SuperScalarProgramFunc []byte

type VMProgramFunc []byte

func (f SuperScalarProgramFunc) Close() error {
	return memory.FreeSlice(pageAllocator, f)
}

func (f VMProgramFunc) Close() error {
	return memory.FreeSlice(pageAllocator, f)
}

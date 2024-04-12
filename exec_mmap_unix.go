//go:build unix && !disable_jit

package randomx

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

func (f ProgramFunc) Execute(rl *RegisterLine) {
	if f == nil {
		panic("program is nil")
	}
	memoryPtr := &f
	fun := *(*func(rl *RegisterLine))(unsafe.Pointer(&memoryPtr))

	fun(rl)
}

func (f ProgramFunc) Close() error {
	return unix.Munmap(f)
}

func mapProgram(program []byte) ProgramFunc {
	// Write only
	execFunc, err := unix.Mmap(-1, 0, len(program), unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		panic(err)
	}

	// Introduce machine code into the memory region
	copy(execFunc, program)

	// uphold W^X

	// Read and Exec only
	err = unix.Mprotect(execFunc, unix.PROT_READ|unix.PROT_EXEC)
	if err != nil {
		defer func() {
			// unmap if we err
			err := unix.Munmap(execFunc)
			if err != nil {
				panic(err)
			}
		}()
		panic(err)
	}

	return execFunc
}

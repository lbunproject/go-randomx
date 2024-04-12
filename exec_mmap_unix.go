//go:build unix && !disable_jit

package randomx

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

func (f ProgramFunc) Execute(rl *RegisterLine) {
	memoryPtr := &f
	fun := *(*func(rl *RegisterLine))(unsafe.Pointer(&memoryPtr))

	fun(rl)
}

func (f ProgramFunc) Close() error {
	return unix.Munmap(f)
}

func mapProgram(program []byte) ProgramFunc {
	execFunc, err := unix.Mmap(
		-1,
		0,
		len(program),
		unix.PROT_READ|unix.PROT_WRITE|unix.PROT_EXEC,
		unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		panic(err)
	}

	copy(execFunc, program)

	// Remove PROT_WRITE
	err = unix.Mprotect(execFunc, unix.PROT_READ|unix.PROT_EXEC)
	if err != nil {
		panic(err)
	}

	return execFunc
}

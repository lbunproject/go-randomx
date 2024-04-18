//go:build unix && !disable_jit && !purego

package randomx

import (
	"golang.org/x/sys/unix"
)

func (f SuperScalarProgramFunc) Close() error {
	return unix.Munmap(f)
}
func (f VMProgramFunc) Close() error {
	return unix.Munmap(f)
}

func mapProgramRW(execFunc []byte) {
	err := unix.Mprotect(execFunc, unix.PROT_READ|unix.PROT_WRITE)
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
}

func mapProgramRX(execFunc []byte) {
	err := unix.Mprotect(execFunc, unix.PROT_READ|unix.PROT_EXEC)
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
}

// mapProgramRWX insecure!
func mapProgramRWX(execFunc []byte) {
	err := unix.Mprotect(execFunc, unix.PROT_READ|unix.PROT_WRITE|unix.PROT_EXEC)
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
}

func mapProgram(program []byte, size int) []byte {
	// Read and Write only
	execFunc, err := unix.Mmap(-1, 0, max(size, len(program)), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
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

package randomx

import (
	"golang.org/x/sys/cpu"
	"runtime"
)

const RANDOMX_FLAG_DEFAULT = 0

const (
	// RANDOMX_FLAG_LARGE_PAGES not implemented
	RANDOMX_FLAG_LARGE_PAGES = 1 << iota
	// RANDOMX_FLAG_HARD_AES not implemented
	RANDOMX_FLAG_HARD_AES
	// RANDOMX_FLAG_FULL_MEM Selects between full or light mode dataset
	RANDOMX_FLAG_FULL_MEM
	// RANDOMX_FLAG_JIT Enables JIT features
	RANDOMX_FLAG_JIT
	// RANDOMX_FLAG_SECURE Enables W^X for JIT code
	RANDOMX_FLAG_SECURE
	RANDOMX_FLAG_ARGON2_SSSE3
	RANDOMX_FLAG_ARGON2_AVX2
	RANDOMX_FLAG_ARGON2
)

func GetFlags() (flags uint64) {
	flags = RANDOMX_FLAG_DEFAULT
	if runtime.GOARCH == "amd64" {
		flags |= RANDOMX_FLAG_JIT

		if cpu.X86.HasAES {
			flags |= RANDOMX_FLAG_HARD_AES
		}
	}
	return flags
}

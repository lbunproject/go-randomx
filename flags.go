package randomx

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/aes"
	"golang.org/x/sys/cpu"
	"runtime"
)

type Flags uint64

func (f Flags) Has(flags Flags) bool {
	return f&flags == flags
}

func (f Flags) HasJIT() bool {
	return f.Has(RANDOMX_FLAG_JIT) && supportsJIT
}

const RANDOMX_FLAG_DEFAULT Flags = 0

const (
	// RANDOMX_FLAG_LARGE_PAGES Select large page allocation for dataset
	RANDOMX_FLAG_LARGE_PAGES = Flags(1 << iota)
	// RANDOMX_FLAG_HARD_AES Selects between hardware or software AES
	RANDOMX_FLAG_HARD_AES
	// RANDOMX_FLAG_FULL_MEM Selects between full or light mode dataset
	RANDOMX_FLAG_FULL_MEM
	// RANDOMX_FLAG_JIT Enables JIT features
	RANDOMX_FLAG_JIT
	// RANDOMX_FLAG_SECURE Enables W^X for JIT code
	RANDOMX_FLAG_SECURE
	RANDOMX_FLAG_ARGON2_SSSE3
	RANDOMX_FLAG_ARGON2_AVX2
	RANDOMX_FLAG_ARGON2 = RANDOMX_FLAG_ARGON2_AVX2 | RANDOMX_FLAG_ARGON2_SSSE3
)

// GetFlags The recommended flags to be used on the current machine.
// Does not include:
// * RANDOMX_FLAG_LARGE_PAGES
// * RANDOMX_FLAG_FULL_MEM
// * RANDOMX_FLAG_SECURE
// These flags must be added manually if desired.
//
// On OpenBSD RANDOMX_FLAG_SECURE is enabled by default in JIT mode as W^X is enforced by the OS.
func GetFlags() (flags Flags) {
	flags = RANDOMX_FLAG_DEFAULT
	if runtime.GOARCH == "amd64" {
		flags |= RANDOMX_FLAG_JIT

		if aes.HasHardAESImplementation && cpu.X86.HasAES {
			flags |= RANDOMX_FLAG_HARD_AES
		}

		if cpu.X86.HasSSSE3 {
			flags |= RANDOMX_FLAG_ARGON2_SSSE3
		}

		if cpu.X86.HasAVX2 {
			flags |= RANDOMX_FLAG_ARGON2_AVX2
		}
	}

	if runtime.GOOS == "openbsd" || runtime.GOOS == "netbsd" || ((runtime.GOOS == "darwin" || runtime.GOOS == "ios") && runtime.GOARCH == "arm64") {
		flags |= RANDOMX_FLAG_SECURE
	}

	return flags
}

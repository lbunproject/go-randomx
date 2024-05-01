package randomx

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/aes"
	"golang.org/x/sys/cpu"
	"runtime"
)

type Flag uint64

const RANDOMX_FLAG_DEFAULT Flag = 0

const (
	// RANDOMX_FLAG_LARGE_PAGES not implemented
	RANDOMX_FLAG_LARGE_PAGES = Flag(1 << iota)
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
	RANDOMX_FLAG_ARGON2 = RANDOMX_FLAG_ARGON2_AVX2 | RANDOMX_FLAG_ARGON2_SSSE3
)

func GetFlags() (flags Flag) {
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

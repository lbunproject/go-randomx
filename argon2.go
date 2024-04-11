package randomx

import "golang.org/x/crypto/blake2b"

import (
	_ "golang.org/x/crypto/argon2"
	_ "unsafe"
)

// see reference configuration.h
// Cache size in KiB. Must be a power of 2.
const RANDOMX_ARGON_MEMORY = 262144

// Number of Argon2d iterations for Cache initialization.
const RANDOMX_ARGON_ITERATIONS = 3

// Number of parallel lanes for Cache initialization.
const RANDOMX_ARGON_LANES = 1

// Argon2d salt
const RANDOMX_ARGON_SALT = "RandomX\x03"
const ArgonSaltSize uint32 = 8 //sizeof("" RANDOMX_ARGON_SALT) - 1

const ArgonBlockSize uint32 = 1024

type argonBlock [128]uint64

const syncPoints = 4

//go:linkname argon2_initHash golang.org/x/crypto/argon2.initHash
func argon2_initHash(password, salt, key, data []byte, time, memory, threads, keyLen uint32, mode int) [blake2b.Size + 8]byte

//go:linkname argon2_initBlocks golang.org/x/crypto/argon2.initBlocks
func argon2_initBlocks(h0 *[blake2b.Size + 8]byte, memory, threads uint32) []argonBlock

//go:linkname argon2_processBlocks golang.org/x/crypto/argon2.processBlocks
func argon2_processBlocks(B []argonBlock, time, memory, threads uint32, mode int)

// argon2_buildBlocks From golang.org/x/crypto/argon2.deriveKey without last deriveKey call
func argon2_buildBlocks(password, salt, secret, data []byte, time, memory uint32, threads uint8, keyLen uint32) []argonBlock {
	if time < 1 {
		panic("argon2: number of rounds too small")
	}
	if threads < 1 {
		panic("argon2: parallelism degree too low")
	}
	const mode = 0 /* argon2d */
	h0 := argon2_initHash(password, salt, secret, data, time, memory, uint32(threads), keyLen, mode)

	memory = memory / (syncPoints * uint32(threads)) * (syncPoints * uint32(threads))
	if memory < 2*syncPoints*uint32(threads) {
		memory = 2 * syncPoints * uint32(threads)
	}
	B := argon2_initBlocks(&h0, memory, uint32(threads))
	argon2_processBlocks(B, time, memory, uint32(threads), mode)

	return B
}

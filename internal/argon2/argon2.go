package argon2

import "golang.org/x/crypto/blake2b"

import (
	_ "golang.org/x/crypto/argon2"
	_ "unsafe"
)

const BlockSize uint32 = 1024

type Block [BlockSize / 8]uint64

const syncPoints = 4

//go:linkname initHash golang.org/x/crypto/argon2.initHash
func initHash(password, salt, key, data []byte, time, memory, threads, keyLen uint32, mode int) [blake2b.Size + 8]byte

//go:linkname initBlocks golang.org/x/crypto/argon2.initBlocks
func initBlocks(h0 *[blake2b.Size + 8]byte, memory, threads uint32) []Block

//go:linkname processBlocks golang.org/x/crypto/argon2.processBlocks
func processBlocks(B []Block, time, memory, threads uint32, mode int)

// BuildBlocks From golang.org/x/crypto/argon2.deriveKey without last deriveKey call
func BuildBlocks(password, salt []byte, time, memory uint32, threads uint8) []Block {
	if time < 1 {
		panic("argon2: number of rounds too small")
	}
	if threads < 1 {
		panic("argon2: parallelism degree too low")
	}
	const mode = 0 /* argon2d */
	const keyLen = 0
	h0 := initHash(password, salt, nil, nil, time, memory, uint32(threads), keyLen, mode)

	memory = memory / (syncPoints * uint32(threads)) * (syncPoints * uint32(threads))
	if memory < 2*syncPoints*uint32(threads) {
		memory = 2 * syncPoints * uint32(threads)
	}
	B := initBlocks(&h0, memory, uint32(threads))
	processBlocks(B, time, memory, uint32(threads), mode)

	return B
}

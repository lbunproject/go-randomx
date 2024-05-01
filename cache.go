package randomx

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/argon2"
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/blake2"
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/keys"
	"runtime"
	"slices"
	"unsafe"
)

type MemoryBlock [128]uint64

func (m *MemoryBlock) GetLine(addr uint64) *RegisterLine {
	addr >>= 3
	return (*RegisterLine)(unsafe.Pointer(unsafe.SliceData(m[addr : addr+8 : addr+8])))
}

type Cache struct {
	blocks []MemoryBlock

	programs [RANDOMX_PROGRAM_COUNT]SuperScalarProgram

	jitPrograms [RANDOMX_PROGRAM_COUNT]SuperScalarProgramFunc

	flags Flags
}

// NewCache Creates a randomx_cache structure and allocates memory for RandomX Cache.
// *
// * @param flags is any combination of these 2 flags (each flag can be set or not set):
// *        RANDOMX_FLAG_LARGE_PAGES - allocate memory in large pages
// *        RANDOMX_FLAG_JIT - create cache structure with JIT compilation support; this makes
// *                           subsequent Dataset initialization faster
// *        Optionally, one of these two flags may be selected:
// *        RANDOMX_FLAG_ARGON2_SSSE3 - optimized Argon2 for CPUs with the SSSE3 instruction set
// *                                   makes subsequent cache initialization faster
// *        RANDOMX_FLAG_ARGON2_AVX2 - optimized Argon2 for CPUs with the AVX2 instruction set
// *                                   makes subsequent cache initialization faster
// *
// * @return Pointer to an allocated randomx_cache structure.
// *         Returns NULL if:
// *         (1) memory allocation fails
// *         (2) the RANDOMX_FLAG_JIT is set and JIT compilation is not supported on the current platform
// *         (3) an invalid or unsupported RANDOMX_FLAG_ARGON2 value is set
// */
func NewCache(flags Flags) *Cache {
	return &Cache{
		flags: flags,
	}
}

func (c *Cache) hasInitializedJIT() bool {
	return c.flags.HasJIT() && c.jitPrograms[0] != nil
}

// Close Releases all memory occupied by the Cache structure.
func (c *Cache) Close() error {
	for _, p := range c.jitPrograms {
		if p != nil {
			err := p.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Init Initializes the cache memory and SuperscalarHash using the provided key value.
// Does nothing if called again with the same key value.
func (c *Cache) Init(key []byte) {

	kkey := slices.Clone(key)

	argonBlocks := argon2.BuildBlocks(kkey, []byte(RANDOMX_ARGON_SALT), RANDOMX_ARGON_ITERATIONS, RANDOMX_ARGON_MEMORY, RANDOMX_ARGON_LANES)

	memoryBlocks := unsafe.Slice((*MemoryBlock)(unsafe.Pointer(unsafe.SliceData(argonBlocks))), int(unsafe.Sizeof(argon2.Block{}))/int(unsafe.Sizeof(MemoryBlock{}))*len(argonBlocks))

	c.blocks = memoryBlocks

	const nonce uint32 = 0

	gen := blake2.New(kkey, nonce)
	for i := range c.programs {
		// build a superscalar program
		prog := BuildSuperScalarProgram(gen)

		if c.flags.HasJIT() {
			c.jitPrograms[i] = generateSuperscalarCode(prog)
			// fallback if can't compile program
			if c.jitPrograms[i] == nil {
				c.programs[i] = prog
			} else {
				c.programs[i] = SuperScalarProgram{prog[0]}
			}
		} else {
			c.programs[i] = prog
		}
	}

}

const Mask = CacheSize/CacheLineSize - 1

// getMixBlock fetch a 64 byte block in uint64 form
func (c *Cache) getMixBlock(addr uint64) *RegisterLine {

	addr = (addr & Mask) * CacheLineSize

	block := addr / 1024
	return c.blocks[block].GetLine(addr % 1024)
}

func (c *Cache) GetMemory() []MemoryBlock {
	return c.blocks
}

func (c *Cache) initDataset(rl *RegisterLine, itemNumber uint64) {
	registerValue := itemNumber

	rl[0] = (itemNumber + 1) * keys.SuperScalar_Constants[0]
	rl[1] = rl[0] ^ keys.SuperScalar_Constants[1]
	rl[2] = rl[0] ^ keys.SuperScalar_Constants[2]
	rl[3] = rl[0] ^ keys.SuperScalar_Constants[3]
	rl[4] = rl[0] ^ keys.SuperScalar_Constants[4]
	rl[5] = rl[0] ^ keys.SuperScalar_Constants[5]
	rl[6] = rl[0] ^ keys.SuperScalar_Constants[6]
	rl[7] = rl[0] ^ keys.SuperScalar_Constants[7]

	if c.hasInitializedJIT() {
		if c.flags.HasJIT() {
			// Lock due to external JIT madness
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
		}

		for i := 0; i < RANDOMX_CACHE_ACCESSES; i++ {
			mix := c.getMixBlock(registerValue)

			c.jitPrograms[i].Execute(uintptr(unsafe.Pointer(rl)))

			for q := range rl {
				rl[q] ^= mix[q]
			}

			registerValue = rl[c.programs[i].AddressRegister()]

		}
	} else {
		for i := 0; i < RANDOMX_CACHE_ACCESSES; i++ {
			mix := c.getMixBlock(registerValue)

			program := c.programs[i]

			executeSuperscalar(program.Program(), rl)

			for q := range rl {
				rl[q] ^= mix[q]
			}

			registerValue = rl[program.AddressRegister()]

		}
	}
}

func (c *Cache) datasetInit(dataset []RegisterLine, startItem, endItem uint64) {
	for itemNumber := startItem; itemNumber < endItem; itemNumber, dataset = itemNumber+1, dataset[1:] {
		c.initDataset(&dataset[0], itemNumber)
	}
}

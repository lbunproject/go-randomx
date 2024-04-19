package randomx

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v2/argon2"
	"git.gammaspectra.live/P2Pool/go-randomx/v2/keys"
	"runtime"
	"slices"
	"unsafe"
)

type MemoryBlock [128]uint64

func (m *MemoryBlock) GetLine(addr uint64) *RegisterLine {
	addr >>= 3
	return (*RegisterLine)(unsafe.Pointer(unsafe.SliceData(m[addr : addr+8 : addr+8])))
}

type Randomx_Cache struct {
	Blocks []MemoryBlock

	Programs [RANDOMX_PROGRAM_COUNT]SuperScalarProgram

	JitPrograms [RANDOMX_PROGRAM_COUNT]SuperScalarProgramFunc

	Flags uint64
}

func Randomx_alloc_cache(flags uint64) *Randomx_Cache {
	if flags == RANDOMX_FLAG_DEFAULT {
		flags = RANDOMX_FLAG_JIT
	}
	return &Randomx_Cache{
		Flags: flags,
	}
}

func (cache *Randomx_Cache) HasJIT() bool {
	return cache.Flags&RANDOMX_FLAG_JIT > 0 && cache.JitPrograms[0] != nil
}

func (cache *Randomx_Cache) VM_Initialize() *VM {

	return &VM{
		Dataset: &Randomx_DatasetLight{
			Cache: cache,
		},
	}
}

func (cache *Randomx_Cache) Close() error {
	for _, p := range cache.JitPrograms {
		if p != nil {
			err := p.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cache *Randomx_Cache) Init(key []byte) {
	// Lock due to external JIT madness
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	kkey := slices.Clone(key)

	argonBlocks := argon2.BuildBlocks(kkey, []byte(RANDOMX_ARGON_SALT), []byte{}, []byte{}, RANDOMX_ARGON_ITERATIONS, RANDOMX_ARGON_MEMORY, RANDOMX_ARGON_LANES, 0)

	memoryBlocks := unsafe.Slice((*MemoryBlock)(unsafe.Pointer(unsafe.SliceData(argonBlocks))), int(unsafe.Sizeof(argon2.Block{}))/int(unsafe.Sizeof(MemoryBlock{}))*len(argonBlocks))

	cache.Blocks = memoryBlocks

	nonce := uint32(0) //uint32(len(key))
	gen := Init_Blake2Generator(key, nonce)
	for i := 0; i < 8; i++ {
		cache.Programs[i] = Build_SuperScalar_Program(gen) // build a superscalar program
		if cache.Flags&RANDOMX_FLAG_JIT > 0 {
			cache.JitPrograms[i] = generateSuperscalarCode(cache.Programs[i])
		}
	}

}

const Mask = CacheSize/CacheLineSize - 1

// GetMixBlock fetch a 64 byte block in uint64 form
func (cache *Randomx_Cache) GetMixBlock(addr uint64) *RegisterLine {

	addr = (addr & Mask) * CacheLineSize

	block := addr / 1024
	return cache.Blocks[block].GetLine(addr % 1024)
}

func (cache *Randomx_Cache) InitDatasetItem(rl *RegisterLine, itemNumber uint64) {
	registerValue := itemNumber

	rl[0] = (itemNumber + 1) * keys.SuperScalar_Constants[0]
	rl[1] = rl[0] ^ keys.SuperScalar_Constants[1]
	rl[2] = rl[0] ^ keys.SuperScalar_Constants[2]
	rl[3] = rl[0] ^ keys.SuperScalar_Constants[3]
	rl[4] = rl[0] ^ keys.SuperScalar_Constants[4]
	rl[5] = rl[0] ^ keys.SuperScalar_Constants[5]
	rl[6] = rl[0] ^ keys.SuperScalar_Constants[6]
	rl[7] = rl[0] ^ keys.SuperScalar_Constants[7]

	for i := 0; i < RANDOMX_CACHE_ACCESSES; i++ {
		mix := cache.GetMixBlock(registerValue)

		program := cache.Programs[i]

		executeSuperscalar(program.Program(), rl)

		for q := range rl {
			rl[q] ^= mix[q]
		}

		registerValue = rl[program.AddressRegister()]

	}
}

func (cache *Randomx_Cache) InitDatasetItemJIT(rl *RegisterLine, itemNumber uint64) {
	registerValue := itemNumber

	rl[0] = (itemNumber + 1) * keys.SuperScalar_Constants[0]
	rl[1] = rl[0] ^ keys.SuperScalar_Constants[1]
	rl[2] = rl[0] ^ keys.SuperScalar_Constants[2]
	rl[3] = rl[0] ^ keys.SuperScalar_Constants[3]
	rl[4] = rl[0] ^ keys.SuperScalar_Constants[4]
	rl[5] = rl[0] ^ keys.SuperScalar_Constants[5]
	rl[6] = rl[0] ^ keys.SuperScalar_Constants[6]
	rl[7] = rl[0] ^ keys.SuperScalar_Constants[7]

	for i := 0; i < RANDOMX_CACHE_ACCESSES; i++ {
		mix := cache.GetMixBlock(registerValue)

		cache.JitPrograms[i].Execute(uintptr(unsafe.Pointer(rl)))

		for q := range rl {
			rl[q] ^= mix[q]
		}

		registerValue = rl[cache.Programs[i].AddressRegister()]

	}
}

func (cache *Randomx_Cache) initDataset(dataset []RegisterLine, startItem, endItem uint64) {
	panic("todo")
	for itemNumber := startItem; itemNumber < endItem; itemNumber, dataset = itemNumber+1, dataset[1:] {
		cache.InitDatasetItem(&dataset[0], itemNumber)
	}
}

package randomx

type Randomx_Cache struct {
	Blocks []block

	Programs [RANDOMX_PROGRAM_COUNT]*SuperScalarProgram
}

func Randomx_alloc_cache(flags uint64) *Randomx_Cache {
	return &Randomx_Cache{}
}

func (cache *Randomx_Cache) VM_Initialize() *VM {

	return &VM{
		Dataset: &Randomx_DatasetLight{
			Cache: cache,
		},
	}
}

func (cache *Randomx_Cache) Randomx_init_cache(key []byte) {
	//fmt.Printf("appending null byte is not necessary but only done for testing")
	kkey := append([]byte{}, key...)
	//kkey = append(kkey,0)
	//cache->initialize(cache, key, keySize);
	cache.Blocks = buildBlocks(argon2d, kkey, []byte(RANDOMX_ARGON_SALT), []byte{}, []byte{}, RANDOMX_ARGON_ITERATIONS, RANDOMX_ARGON_MEMORY, RANDOMX_ARGON_LANES, 0)

}

// fetch a 64 byte block in uint64 form
func (cache *Randomx_Cache) GetMixBlock(addr uint64) []uint64 {

	mask := CacheSize/CacheLineSize - 1

	addr = (addr & mask) * CacheLineSize

	block := addr / 1024
	index_within_block := (addr % 1024) / 8

	return cache.Blocks[block][index_within_block : index_within_block+8]
}

func (cache *Randomx_Cache) InitDatasetItem(out *registerLine, itemnumber uint64) {
	const superscalarMul0 uint64 = 6364136223846793005
	const superscalarAdd1 uint64 = 9298411001130361340
	const superscalarAdd2 uint64 = 12065312585734608966
	const superscalarAdd3 uint64 = 9306329213124626780
	const superscalarAdd4 uint64 = 5281919268842080866
	const superscalarAdd5 uint64 = 10536153434571861004
	const superscalarAdd6 uint64 = 3398623926847679864
	const superscalarAdd7 uint64 = 9549104520008361294

	var rl registerLine

	register_value := itemnumber
	_ = register_value

	rl[0] = (itemnumber + 1) * superscalarMul0
	rl[1] = rl[0] ^ superscalarAdd1
	rl[2] = rl[0] ^ superscalarAdd2
	rl[3] = rl[0] ^ superscalarAdd3
	rl[4] = rl[0] ^ superscalarAdd4
	rl[5] = rl[0] ^ superscalarAdd5
	rl[6] = rl[0] ^ superscalarAdd6
	rl[7] = rl[0] ^ superscalarAdd7

	for i := 0; i < RANDOMX_CACHE_ACCESSES; i++ {
		mix := cache.GetMixBlock(register_value)

		program := cache.Programs[i]

		executeSuperscalar(program, &rl)

		for q := range rl {
			rl[q] ^= mix[q]
		}

		register_value = rl[program.AddressRegister]

	}

	for q := range rl {
		out[q] = rl[q]
	}
}

func (cache *Randomx_Cache) initDataset(start_item, end_item uint64) {
	for itemnumber := start_item; itemnumber < end_item; itemnumber++ {

		cache.InitDatasetItem(nil, itemnumber)

		// dataset_index += CacheLineSize
		//fmt.Printf("exiting dataset item\n")
		break

	}
}

//go:build amd64 || arm64 || arm64be || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x || sparc64

package randomx

const DatasetSize = RANDOMX_DATASET_BASE_SIZE + RANDOMX_DATASET_EXTRA_SIZE

const DatasetItemCount = DatasetSize / CacheLineSize

type DatasetFull struct {
	cache  *Cache
	memory [DatasetItemCount]RegisterLine
}

func NewFullDataset(cache *Cache) *DatasetFull {
	return &DatasetFull{
		cache: cache,
	}
}

func (d *DatasetFull) PrefetchDataset(address uint64) {

}

func (d *DatasetFull) ReadDataset(address uint64, r *RegisterLine) {
	cache := &d.memory[address/CacheLineSize]

	for i := range r {
		r[i] ^= cache[i]
	}
}

func (d *DatasetFull) Cache() *Cache {
	return d.cache
}

func (d *DatasetFull) Flags() Flag {
	return d.cache.Flags
}

func (d *DatasetFull) Memory() []RegisterLine {
	return d.memory[:]
}

func (d *DatasetFull) InitDataset(startItem, itemCount uint64) {
	if startItem >= DatasetItemCount || itemCount > DatasetItemCount {
		panic("out of range")
	}
	if startItem+itemCount > DatasetItemCount {
		panic("out of range")
	}
	d.cache.InitDataset(d.memory[startItem:startItem+itemCount], startItem, startItem+itemCount)
}

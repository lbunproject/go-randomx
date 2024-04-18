package randomx

type Randomx_DatasetLight struct {
	Cache  *Randomx_Cache
	Memory []uint64
}

func (d *Randomx_DatasetLight) PrefetchDataset(address uint64) {

}

func (d *Randomx_DatasetLight) ReadDataset(address uint64, r, cache *RegisterLine) {
	if d.Cache.HasJIT() {
		d.Cache.InitDatasetItemJIT(cache, address/CacheLineSize)
	} else {
		d.Cache.InitDatasetItem(cache, address/CacheLineSize)
	}

	for i := range r {
		r[i] ^= cache[i]
	}
}

func (d *Randomx_DatasetLight) Flags() uint64 {
	return d.Cache.Flags
}

func (d *Randomx_DatasetLight) InitDataset(startItem, endItem uint64) {
	//d.Cache.initDataset(d.Cache.Programs)
}

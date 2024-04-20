package randomx

type DatasetLight struct {
	cache *Cache
}

func NewLightDataset(cache *Cache) *DatasetLight {
	return &DatasetLight{
		cache: cache,
	}
}

func (d *DatasetLight) PrefetchDataset(address uint64) {

}

func (d *DatasetLight) ReadDataset(address uint64, r *RegisterLine) {
	var cache RegisterLine
	if d.cache.HasJIT() {
		d.cache.InitDatasetItemJIT(&cache, address/CacheLineSize)
	} else {
		d.cache.InitDatasetItem(&cache, address/CacheLineSize)
	}

	for i := range r {
		r[i] ^= cache[i]
	}
}

func (d *DatasetLight) Flags() uint64 {
	return d.cache.Flags
}

func (d *DatasetLight) Cache() *Cache {
	return d.cache
}

func (d *DatasetLight) Memory() []RegisterLine {
	return nil
}

func (d *DatasetLight) InitDataset(startItem, itemCount uint64) {

}

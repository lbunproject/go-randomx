package randomx

type Randomx_DatasetLight struct {
	Cache *Randomx_Cache
}

func (d *Randomx_DatasetLight) PrefetchDataset(address uint64) {

}

func (d *Randomx_DatasetLight) ReadDataset(address uint64, r *registerLine) {
	var out registerLine

	d.Cache.InitDatasetItem(&out, address/CacheLineSize)

	for i := range r {
		r[i] ^= out[i]
	}
}

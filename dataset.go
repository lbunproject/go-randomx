package randomx

import "sync"

type Dataset interface {
	InitDataset(startItem, itemCount uint64)
	ReadDataset(address uint64, r *RegisterLine)
	PrefetchDataset(address uint64)
	Flags() Flag
	Cache() *Cache
	Memory() []RegisterLine
}

func NewDataset(cache *Cache) Dataset {
	if cache.Flags&RANDOMX_FLAG_FULL_MEM > 0 {
		if ds := NewFullDataset(cache); ds != nil {
			return ds
		}
		return nil
	}
	return NewLightDataset(cache)
}

func InitDatasetParallel(dataset Dataset, n int) {
	n = max(1, n)

	var wg sync.WaitGroup
	for i := uint64(1); i < uint64(n); i++ {
		a := (DatasetItemCount * i) / uint64(n)
		b := (DatasetItemCount * (i + 1)) / uint64(n)

		wg.Add(1)
		go func(a, b uint64) {
			defer wg.Done()
			dataset.InitDataset(a, b-a)
		}(a, b)
	}

	dataset.InitDataset(0, DatasetItemCount/uint64(n))
	wg.Wait()
}

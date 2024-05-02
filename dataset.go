package randomx

import (
	"errors"
	"sync"
	"unsafe"
)

const DatasetSize = RANDOMX_DATASET_BASE_SIZE + RANDOMX_DATASET_EXTRA_SIZE

const DatasetItemCount = DatasetSize / CacheLineSize

type Dataset struct {
	memory []RegisterLine
}

// NewDataset Creates a randomx_dataset structure and allocates memory for RandomX Dataset.
// Only one flag is supported (can be set or not set): RANDOMX_FLAG_LARGE_PAGES - allocate memory in large pages
// Returns nil if allocation fails
func NewDataset(flags Flags) (result *Dataset, err error) {
	defer func() {
		//catch too large memory allocation or unable to allocate, for example on 32-bit targets or out of memory
		if r := recover(); r != nil {
			result = nil
			if e, ok := r.(error); ok && e != nil {
				err = e
			} else {
				err = errors.New("out of memory")
			}
		}
	}()

	//todo: implement large pages, align allocation
	alignedMemory := make([]RegisterLine, DatasetItemCount)
	assertAlignedTo16(uintptr(unsafe.Pointer(unsafe.SliceData(alignedMemory))))

	//todo: err on not large pages

	return &Dataset{
		memory: alignedMemory,
	}, nil
}

func (d *Dataset) prefetchDataset(address uint64) {

}

func (d *Dataset) readDataset(address uint64, r *RegisterLine) {
	cache := &d.memory[address/CacheLineSize]

	for i := range r {
		r[i] ^= cache[i]
	}
}

// Memory Returns a pointer to the internal memory buffer of the dataset structure.
// The size of the internal memory buffer is DatasetItemCount * RANDOMX_DATASET_ITEM_SIZE.
func (d *Dataset) Memory() []RegisterLine {
	return d.memory
}

func (d *Dataset) InitDataset(cache *Cache, startItem, itemCount uint64) {
	if startItem >= DatasetItemCount || itemCount > DatasetItemCount {
		panic("out of range")
	}
	if startItem+itemCount > DatasetItemCount {
		panic("out of range")
	}
	cache.datasetInit(d.memory[startItem:startItem+itemCount], startItem, startItem+itemCount)
}

func (d *Dataset) Close() error {
	return nil
}

func (d *Dataset) InitDatasetParallel(cache *Cache, n int) {
	n = max(1, n)

	var wg sync.WaitGroup
	for i := uint64(1); i < uint64(n); i++ {
		a := (DatasetItemCount * i) / uint64(n)
		b := (DatasetItemCount * (i + 1)) / uint64(n)

		wg.Add(1)
		go func(a, b uint64) {
			defer wg.Done()
			d.InitDataset(cache, a, b-a)
		}(a, b)
	}

	d.InitDataset(cache, 0, DatasetItemCount/uint64(n))
	wg.Wait()
}

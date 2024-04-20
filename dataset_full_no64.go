//go:build !(amd64 || arm64 || arm64be || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x || sparc64)

package randomx

const DatasetSize = RANDOMX_DATASET_BASE_SIZE + RANDOMX_DATASET_EXTRA_SIZE

const DatasetItemCount = DatasetSize / CacheLineSize

type DatasetFull struct {
}

func NewFullDataset(cache *Cache) *DatasetFull {
	return nil
}

func (d *DatasetFull) PrefetchDataset(address uint64) {

}

func (d *DatasetFull) ReadDataset(address uint64, r *RegisterLine) {

}

func (d *DatasetFull) Cache() *Cache {
	return nil
}

func (d *DatasetFull) Flags() uint64 {
	return 0
}

func (d *DatasetFull) Memory() []RegisterLine {
	return nil
}

func (d *DatasetFull) InitDataset(startItem, itemCount uint64) {

}

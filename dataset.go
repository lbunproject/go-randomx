package randomx

type Randomx_Dataset interface {
	InitDataset(startItem, endItem uint64)
	ReadDataset(address uint64, r *RegisterLine)
	PrefetchDataset(address uint64)
}

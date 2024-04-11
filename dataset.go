package randomx

type Randomx_Dataset interface {
	ReadDataset(address uint64, r *registerLine)
	PrefetchDataset(address uint64)
}

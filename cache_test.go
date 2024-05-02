package randomx

import "testing"

func Test_Cache_Init(t *testing.T) {
	t.Parallel()

	cache, err := NewCache(GetFlags())
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	cache.Init(Tests[1].key)

	memory := cache.GetMemory()

	var tests = []struct {
		index int
		value uint64
	}{
		{0, 0x191e0e1d23c02186},
		{1568413, 0xf1b62fe6210bf8b1},
		{33554431, 0x1f47f056d05cd99b},
	}

	for i, tt := range tests {
		if memory[tt.index/128][tt.index%128] != tt.value {
			t.Errorf("i=%d, index=%d", i, tt.index)
			t.Errorf("expected=%016x, actual=%016x", tt.value, memory[tt.index/128][tt.index%128])
		}
	}

}

func Test_Cache_InitDataset(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		index int
		value uint64
	}{
		{0, 0x680588a85ae222db},
		{10000000, 0x7943a1f6186ffb72},
		{20000000, 0x9035244d718095e1},
		{30000000, 0x145a5091f7853099},
	}

	t.Run("interpreter", func(t *testing.T) {
		t.Parallel()

		flags := GetFlags()
		flags &^= RANDOMX_FLAG_JIT

		cache, err := NewCache(flags)
		if err != nil {
			t.Fatal(err)
		}
		defer cache.Close()
		cache.Init(Tests[1].key)

		var datasetItem RegisterLine

		for i, tt := range tests {
			cache.initDataset(&datasetItem, uint64(tt.index))
			if datasetItem[0] != tt.value {
				t.Errorf("i=%d, index=%d", i, tt.index)
				t.Errorf("expected=%016x, actual=%016x", tt.value, datasetItem[0])
			}
		}
	})

	t.Run("compiler", func(t *testing.T) {
		t.Parallel()

		flags := GetFlags()
		flags |= RANDOMX_FLAG_JIT
		if !flags.HasJIT() {
			t.Skip("not supported on this platform")
		}

		cache, err := NewCache(flags)
		if err != nil {
			t.Fatal(err)
		}
		defer cache.Close()
		cache.Init(Tests[1].key)
		if !cache.hasInitializedJIT() {
			t.Skip("not supported on this platform")
		}

		var datasetItem RegisterLine

		for i, tt := range tests {
			cache.initDataset(&datasetItem, uint64(tt.index))
			if datasetItem[0] != tt.value {
				t.Errorf("i=%d, index=%d", i, tt.index)
				t.Errorf("expected=%016x, actual=%016x", tt.value, datasetItem[0])
			}
		}
	})
}

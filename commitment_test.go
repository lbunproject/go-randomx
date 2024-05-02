package randomx

import (
	"encoding/hex"
	"testing"
)

func Test_CalculateCommitment(t *testing.T) {
	t.Parallel()

	cache := NewCache(GetFlags())
	defer cache.Close()

	test := Tests[1]

	cache.Init(test.key)

	vm, err := NewVM(GetFlags(), cache, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer vm.Close()

	var outputHash [RANDOMX_HASH_SIZE]byte

	vm.CalculateHash(test.input, &outputHash)
	CalculateCommitment(test.input, &outputHash, &outputHash)

	outputHex := hex.EncodeToString(outputHash[:])

	expected := "d53ccf348b75291b7be76f0a7ac8208bbced734b912f6fca60539ab6f86be919"

	if expected != outputHex {
		t.Errorf("key=%v, input=%v", test.key, test.input)
		t.Errorf("expected=%s, actual=%s", expected, outputHex)
		t.FailNow()
	}
}

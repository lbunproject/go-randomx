/*
Copyright (c) 2019 DERO Foundation. All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
may be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package randomx

import (
	"encoding/hex"
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/aes"
	"os"
	"runtime"
	"slices"
	"strings"
)
import "testing"

type testdata struct {
	name  string
	key   []byte
	input []byte
	// expected result, in hex
	expected string
}

func mustHex(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}

var Tests = []testdata{
	{"example", []byte("RandomX example key\x00"), []byte("RandomX example input\x00"), "8a48e5f9db45ab79d9080574c4d81954fe6ac63842214aff73c244b26330b7c9"},
	{"test_a", []byte("test key 000"), []byte("This is a test"), "639183aae1bf4c9a35884cb46b09cad9175f04efd7684e7262a0ac1c2f0b4e3f"},
	{"test_b", []byte("test key 000"), []byte("Lorem ipsum dolor sit amet"), "300a0adb47603dedb42228ccb2b211104f4da45af709cd7547cd049e9489c969"},
	{"test_c", []byte("test key 000"), []byte("sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"), "c36d4ed4191e617309867ed66a443be4075014e2b061bcdaf9ce7b721d2b77a8"},
	{"test_d", []byte("test key 001"), []byte("sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"), "e9ff4503201c0c2cca26d285c93ae883f9b1d30c9eb240b820756f2d5a7905fc"},
	{"test_e", []byte("test key 001"), mustHex("0b0b98bea7e805e0010a2126d287a2a0cc833d312cb786385a7c2f9de69d25537f584a9bc9977b00000000666fd8753bf61a8631f12984e3fd44f4014eca629276817b56f32e9b68bd82f416"), "c56414121acda1713c2f2a819d8ae38aed7c80c35c2a769298d34f03833cd5f1"},
}

func testFlags(name string, flags Flags) (f Flags, skip bool) {
	flags |= GetFlags()

	nn := strings.Split(name, "/")
	switch nn[len(nn)-1] {
	case "interpreter":
		flags &^= RANDOMX_FLAG_JIT
	case "compiler":
		flags |= RANDOMX_FLAG_JIT
		if !flags.HasJIT() {
			return flags, true
		}

	case "softaes":
		flags &^= RANDOMX_FLAG_HARD_AES
	case "hardaes":
		flags |= RANDOMX_FLAG_HARD_AES
		if aes.NewHardAES() == nil {
			return flags, true
		}
	}

	return flags, false
}

func Test_RandomXLight(t *testing.T) {
	t.Parallel()
	for _, n := range []string{"interpreter", "compiler", "softaes", "hardaes"} {
		t.Run(n, func(t *testing.T) {
			t.Parallel()
			tFlags, skip := testFlags(t.Name(), 0)
			if skip {
				t.Skip("not supported on this platform")
			}

			c := NewCache(tFlags)
			if c == nil {
				t.Fatal("nil cache")
			}
			defer func() {
				err := c.Close()
				if err != nil {
					t.Error(err)
				}
			}()

			for _, test := range Tests {
				t.Run(test.name, func(t *testing.T) {
					c.Init(test.key)

					vm, err := NewVM(tFlags, c, nil)
					if err != nil {
						t.Fatal(err)
					}
					defer func() {
						err := vm.Close()
						if err != nil {
							t.Error(err)
						}
					}()

					var outputHash [RANDOMX_HASH_SIZE]byte

					vm.CalculateHash(test.input, &outputHash)

					outputHex := hex.EncodeToString(outputHash[:])

					if outputHex != test.expected {
						t.Errorf("key=%v, input=%v", test.key, test.input)
						t.Errorf("expected=%s, actual=%s", test.expected, outputHex)
						t.FailNow()
					}
				})
			}

		})
	}
}

func Test_RandomXFull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full mode with -short")
	}

	if os.Getenv("CI") != "" {
		t.Skip("Skipping full mode in CI environment")
	}

	for _, n := range []string{"interpreter", "compiler", "softaes", "hardaes"} {
		t.Run(n, func(t *testing.T) {

			tFlags, skip := testFlags(t.Name(), RANDOMX_FLAG_FULL_MEM)
			if skip {
				t.Skip("not supported on this platform")
			}

			c := NewCache(tFlags)
			if c == nil {
				t.Fatal("nil cache")
			}
			defer func() {
				err := c.Close()
				if err != nil {
					t.Error(err)
				}
			}()

			dataset, err := NewDataset(tFlags)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				err := dataset.Close()
				if err != nil {
					t.Error(err)
				}
			}()

			for _, test := range Tests {
				t.Run(test.name, func(t *testing.T) {
					c.Init(test.key)
					dataset.InitDatasetParallel(c, runtime.NumCPU())

					vm, err := NewVM(tFlags, nil, dataset)
					if err != nil {
						t.Fatal(err)
					}
					defer func() {
						err := vm.Close()
						if err != nil {
							t.Error(err)
						}
					}()

					var outputHash [RANDOMX_HASH_SIZE]byte

					vm.CalculateHash(test.input, &outputHash)

					outputHex := hex.EncodeToString(outputHash[:])

					if outputHex != test.expected {
						t.Errorf("key=%v, input=%v", test.key, test.input)
						t.Errorf("expected=%s, actual=%s", test.expected, outputHex)
						t.FailNow()
					}
				})

				// cleanup between runs
				runtime.GC()
			}

		})

		// cleanup 2 GiB between runs
		runtime.GC()
	}
}

var BenchmarkTest = Tests[0]
var BenchmarkCache *Cache
var BenchmarkDataset *Dataset

var BenchmarkFlags = GetFlags()

func TestMain(m *testing.M) {
	if slices.Contains(os.Args, "-test.bench") {
		flags := GetFlags()
		flags |= RANDOMX_FLAG_FULL_MEM
		var err error
		//init light and full dataset
		BenchmarkCache = NewCache(flags)
		defer BenchmarkCache.Close()
		BenchmarkCache.Init(BenchmarkTest.key)

		BenchmarkDataset, err = NewDataset(flags | RANDOMX_FLAG_FULL_MEM)
		if err != nil {
			panic(err)
		}
		defer BenchmarkDataset.Close()
		BenchmarkDataset.InitDatasetParallel(BenchmarkCache, runtime.NumCPU())
	}
	os.Exit(m.Run())
}

func Benchmark_RandomXLight(b *testing.B) {
	b.ReportAllocs()

	vm, err := NewVM(BenchmarkFlags, BenchmarkCache, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer vm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var output_hash [32]byte
		vm.CalculateHash(BenchmarkTest.input, &output_hash)
		runtime.KeepAlive(output_hash)
	}
}

func Benchmark_RandomXFull(b *testing.B) {
	b.ReportAllocs()

	vm, err := NewVM(BenchmarkFlags|RANDOMX_FLAG_FULL_MEM, nil, BenchmarkDataset)
	if err != nil {
		b.Fatal(err)
	}
	defer vm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var output_hash [32]byte
		vm.CalculateHash(BenchmarkTest.input, &output_hash)
		runtime.KeepAlive(output_hash)
	}
}

func Benchmark_RandomXLight_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var output_hash [32]byte

		vm, err := NewVM(BenchmarkFlags, BenchmarkCache, nil)
		if err != nil {
			b.Fatal(err)
		}
		defer vm.Close()

		for pb.Next() {
			vm.CalculateHash(BenchmarkTest.input, &output_hash)
			runtime.KeepAlive(output_hash)
		}
	})
}

func Benchmark_RandomXFull_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var output_hash [32]byte

		vm, err := NewVM(BenchmarkFlags|RANDOMX_FLAG_FULL_MEM, nil, BenchmarkDataset)
		if err != nil {
			b.Fatal(err)
		}
		defer vm.Close()

		for pb.Next() {
			vm.CalculateHash(BenchmarkTest.input, &output_hash)
			runtime.KeepAlive(output_hash)
		}
	})
}

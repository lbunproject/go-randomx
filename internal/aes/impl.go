package aes

type AES interface {

	// HashAes1Rx4
	//
	// Calculate a 512-bit hash of 'input' using 4 lanes of AES.
	// The input is treated as a set of round keys for the encryption
	// of the initial state.
	//
	// 'input' size must be a multiple of 64.
	//
	// For a 2 MiB input, this has the same security as 32768-round
	// AES encryption.
	//
	// Hashing throughput: >20 GiB/s per CPU core with hardware AES
	HashAes1Rx4(input []byte, output *[64]byte)

	// FillAes1Rx4
	//
	// Fill 'output' with pseudorandom data based on 512-bit 'state'.
	// The state is encrypted using a single AES round per 16 bytes of output
	// in 4 lanes.
	//
	// 'output' size must be a multiple of 64.
	//
	// The modified state is written back to 'state' to allow multiple
	// calls to this function.
	FillAes1Rx4(state *[64]byte, output []byte)

	// FillAes4Rx4 used to generate final program
	//
	// 'state' is copied when calling
	FillAes4Rx4(state [64]byte, output []byte)
}

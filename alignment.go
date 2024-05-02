package randomx

func assertAlignedTo16(ptr uintptr) {
	if ptr&0b1111 != 0 {
		panic("not aligned to 16")
	}
}

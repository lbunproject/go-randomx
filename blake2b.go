package randomx

import (
	"encoding/binary"
	"golang.org/x/crypto/blake2b"
)

type Blake2Generator struct {
	data           [64]byte
	dataindex      int
	allocRegIndex  [8]int
	allocRegisters [8]Register
}

func Init_Blake2Generator(key []byte, nonce uint32) *Blake2Generator {
	var b Blake2Generator
	b.dataindex = len(b.data)
	if len(key) > 60 {
		copy(b.data[:], key[0:60])
	} else {
		copy(b.data[:], key)
	}
	binary.LittleEndian.PutUint32(b.data[60:], nonce)

	return &b
}

func (b *Blake2Generator) checkdata(bytesNeeded int) {
	if b.dataindex+bytesNeeded > cap(b.data) {
		//blake2b(data, sizeof(data), data, sizeof(data), nullptr, 0);
		h := blake2b.Sum512(b.data[:])
		copy(b.data[:], h[:])
		b.dataindex = 0
	}

}

func (b *Blake2Generator) GetByte() byte {
	b.checkdata(1)
	ret := b.data[b.dataindex]
	//fmt.Printf("returning byte %02x\n", ret)
	b.dataindex++
	return ret
}
func (b *Blake2Generator) GetUint32() uint32 {
	b.checkdata(4)
	ret := binary.LittleEndian.Uint32(b.data[b.dataindex:])
	b.dataindex += 4

	return ret
}

//go:build !amd64 || purego

package aes

func aesenc(state *[4]uint32, key *[4]uint32) {
	soft_aesenc(state, key)
}

func aesdec(state *[4]uint32, key *[4]uint32) {
	soft_aesdec(state, key)
}

func aesroundtrip_decenc(states *[4][4]uint32, keys *[4][4]uint32) {
	aesdec(&states[0], &keys[0])
	aesenc(&states[1], &keys[1])
	aesdec(&states[2], &keys[2])
	aesenc(&states[3], &keys[3])
}

func aesroundtrip_encdec(states *[4][4]uint32, keys *[4][4]uint32) {
	aesenc(&states[0], &keys[0])
	aesdec(&states[1], &keys[1])
	aesenc(&states[2], &keys[2])
	aesdec(&states[3], &keys[3])
}

func aesroundtrip_encdec1(states *[4][4]uint32, key *[4]uint32) {
	aesenc(&states[0], key)
	aesdec(&states[1], key)
	aesenc(&states[2], key)
	aesdec(&states[3], key)
}

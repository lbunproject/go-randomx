package aes

var te0, te1, te2, te3 = encLut[0], encLut[1], encLut[2], encLut[3]

func soft_aesenc(state *[4]uint32, key *[4]uint32) {

	s0 := state[0]
	s1 := state[1]
	s2 := state[2]
	s3 := state[3]

	state[0] = key[0] ^ te0[uint8(s0)] ^ te1[uint8(s1>>8)] ^ te2[uint8(s2>>16)] ^ te3[uint8(s3>>24)]
	state[1] = key[1] ^ te0[uint8(s1)] ^ te1[uint8(s2>>8)] ^ te2[uint8(s3>>16)] ^ te3[uint8(s0>>24)]
	state[2] = key[2] ^ te0[uint8(s2)] ^ te1[uint8(s3>>8)] ^ te2[uint8(s0>>16)] ^ te3[uint8(s1>>24)]
	state[3] = key[3] ^ te0[uint8(s3)] ^ te1[uint8(s0>>8)] ^ te2[uint8(s1>>16)] ^ te3[uint8(s2>>24)]
}

var td0, td1, td2, td3 = decLut[0], decLut[1], decLut[2], decLut[3]

func soft_aesdec(state *[4]uint32, key *[4]uint32) {

	s0 := state[0]
	s1 := state[1]
	s2 := state[2]
	s3 := state[3]

	state[0] = key[0] ^ td0[uint8(s0)] ^ td1[uint8(s3>>8)] ^ td2[uint8(s2>>16)] ^ td3[uint8(s1>>24)]
	state[1] = key[1] ^ td0[uint8(s1)] ^ td1[uint8(s0>>8)] ^ td2[uint8(s3>>16)] ^ td3[uint8(s2>>24)]
	state[2] = key[2] ^ td0[uint8(s2)] ^ td1[uint8(s1>>8)] ^ td2[uint8(s0>>16)] ^ td3[uint8(s3>>24)]
	state[3] = key[3] ^ td0[uint8(s3)] ^ td1[uint8(s2>>8)] ^ td2[uint8(s1>>16)] ^ td3[uint8(s0>>24)]
}

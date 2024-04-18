//go:build amd64 && !purego

package asm

//go:noescape
func aesenc(state *[4]uint32, key *[4]uint32)

//go:noescape
func aesdec(state *[4]uint32, key *[4]uint32)

//go:noescape
func aesroundtrip_decenc(states *[4][4]uint32, keys *[4][4]uint32)

//go:noescape
func aesroundtrip_encdec(states *[4][4]uint32, keys *[4][4]uint32)

//go:noescape
func aesroundtrip_encdec1(states *[4][4]uint32, key *[4]uint32)

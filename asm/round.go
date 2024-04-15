package asm

import "git.gammaspectra.live/P2Pool/go-randomx/v2/softfloat"

func SetRoundingMode(mode softfloat.RoundingMode) {
	setRoundingMode(uint8(mode))
}

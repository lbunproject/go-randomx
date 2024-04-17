package randomx

import "math"

const (
	mantbits64 uint = 52
	expbits64  uint = 11
)

const mantissaMask = (uint64(1) << mantbits64) - 1
const exponentMask = (uint64(1) << expbits64) - 1
const exponentBias = 1023
const dynamicExponentBits = 4
const staticExponentBits = 4
const constExponentBits uint64 = 0x300
const dynamicMantissaMask = (uint64(1) << (mantbits64 + dynamicExponentBits)) - 1

const mask22bit = (uint64(1) << 22) - 1

func MaskRegisterExponentMantissa(f float64, mode uint64) float64 {
	return math.Float64frombits((math.Float64bits(f) & dynamicMantissaMask) | mode)
}

func ScaleNegate(f float64) float64 {
	return math.Float64frombits(math.Float64bits(f) ^ 0x80F0000000000000)
}

func SmallPositiveFloatBits(entropy uint64) float64 {
	exponent := entropy >> 59 //0..31
	mantissa := entropy & mantissaMask
	exponent += exponentBias
	exponent &= exponentMask
	exponent = exponent << mantbits64
	return math.Float64frombits(exponent | mantissa)
}

func StaticExponent(entropy uint64) uint64 {
	exponent := constExponentBits
	exponent |= (entropy >> (64 - staticExponentBits)) << dynamicExponentBits
	exponent <<= mantbits64
	return exponent
}

func EMask(entropy uint64) uint64 {
	return (entropy & mask22bit) | StaticExponent(entropy)
}

func Xor(a, b float64) float64 {
	return math.Float64frombits(math.Float64bits(a) ^ math.Float64bits(b))
}

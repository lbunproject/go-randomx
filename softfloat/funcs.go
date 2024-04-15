package softfloat

import "math"

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

package softfloat

const (
	mantbits64 uint = 52
	expbits64  uint = 11
	bias64          = -1<<(expbits64-1) + 1

	nan64 uint64 = (1<<expbits64-1)<<mantbits64 + 1<<(mantbits64-1) // quiet NaN, 0 payload
	inf64 uint64 = (1<<expbits64 - 1) << mantbits64
	neg64 uint64 = 1 << (expbits64 + mantbits64)
)

const mantissaMask = (uint64(1) << mantbits64) - 1
const exponentMask = (uint64(1) << expbits64) - 1
const exponentBias = 1023
const dynamicExponentBits = 4
const staticExponentBits = 4
const constExponentBits uint64 = 0x300
const dynamicMantissaMask = (uint64(1) << (mantbits64 + dynamicExponentBits)) - 1

const mask22bit = (uint64(1) << 22) - 1

type RoundingMode uint8

const (
	// RoundingModeToNearest IEEE 754 roundTiesToEven
	RoundingModeToNearest = RoundingMode(iota)

	// RoundingModeToNegative IEEE 754 roundTowardNegative
	RoundingModeToNegative

	// RoundingModeToPositive IEEE 754 roundTowardPositive
	RoundingModeToPositive

	// RoundingModeToZero IEEE 754 roundTowardZero
	RoundingModeToZero
)

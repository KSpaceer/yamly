package token

import (
	"github.com/KSpaceer/yamly/chars"
	"math/bits"
)

type conformationBitmap uint16

func (b conformationBitmap) Set(cst chars.CharSetType, val bool) conformationBitmap {
	if val {
		return b.setTrue(cst)
	}
	return b.setFalse(cst)
}

func (b conformationBitmap) Get(cst chars.CharSetType) (bool, bool) {
	shift := 2 * bits.TrailingZeros16(uint16(cst))
	val := (b & (3 << shift)) >> shift
	switch val {
	case 0, 3:
		return false, false
	case 1:
		return true, true
	case 2:
		return false, true
	}
	return false, false
}

func (b conformationBitmap) setTrue(cst chars.CharSetType) conformationBitmap {
	shift := 2 * bits.TrailingZeros16(uint16(cst))
	// remove false
	b = b & ^(1 << (1 + shift))
	return b | (1 << shift)
}

func (b conformationBitmap) setFalse(cst chars.CharSetType) conformationBitmap {
	shift := 1 + 2*bits.TrailingZeros16(uint16(cst))
	// remove true
	b = b & ^(1 << (shift - 1))
	return b | (1 << shift)
}

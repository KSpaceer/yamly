package token

import "math/bits"

type conformationBitmap uint16

func (b conformationBitmap) Set(cst CharSetType, val bool) conformationBitmap {
	if val {
		return b.setTrue(cst)
	}
	return b.setFalse(cst)
}

func (b conformationBitmap) Get(cst CharSetType) (bool, bool) {
	shift := 2 * bits.TrailingZeros16(uint16(cst))
	val := (b & 3 << shift) >> shift
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

func (b conformationBitmap) setTrue(cst CharSetType) conformationBitmap {
	return b | (1 << (2 * bits.TrailingZeros16(uint16(cst))))
}

func (b conformationBitmap) setFalse(cst CharSetType) conformationBitmap {
	return b | (1 << (1 + 2*bits.TrailingZeros16(uint16(cst))))
}
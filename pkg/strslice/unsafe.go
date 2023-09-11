//go:build !nounsafe

package strslice

import (
	"unsafe"
)

func StringToBytesSlice(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BytesSliceToString(s []byte) string {
	return unsafe.String(unsafe.SliceData(s), len(s))
}

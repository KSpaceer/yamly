//go:build nounsafe || appengine

package strslice

func StringToBytes(s string) []byte {
	return []byte(s)
}

func BytesToString(s []byte) string {
	return string(s)
}

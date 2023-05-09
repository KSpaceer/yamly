package lexer

import "sync"

type runeStream struct {
	runes []rune
	pos   int
}

var (
	runeStreamPool = sync.Pool{
		New: func() any {
			return newRuneStream()
		},
	}
)

func newRuneStream() *runeStream {
	return &runeStream{
		runes: nil,
		pos:   0,
	}
}

func NewRuneStream(src string) RuneStream {
	rs := runeStreamPool.Get().(*runeStream)
	rs.runes = []rune(src)
	rs.pos = 0
	return rs
}

func ReleaseRuneStream(rs RuneStream) {
	switch rs.(type) {
	case *runeStream:
		runeStreamPool.Put(rs)
	}
}

func (r runeStream) EOF() bool {
	return r.pos == len(r.runes)
}

func (r runeStream) Next() rune {
	rn := r.runes[r.pos]
	r.pos++
	return rn
}

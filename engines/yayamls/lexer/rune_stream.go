package lexer

import (
	"unicode/utf8"
)

// EOF indicates the end of file.
const EOF rune = -1

type runeStream struct {
	src []byte
	pos int
}

func newRuneStream(src []byte) *runeStream {
	return &runeStream{
		src: src,
		pos: 0,
	}
}

func (r *runeStream) Next() rune {
	if r.pos >= len(r.src) {
		return EOF
	}
	next, diff := utf8.DecodeRune(r.src[r.pos:])
	r.pos += diff
	return next
}

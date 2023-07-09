package lexer

const EOF rune = -1

type runeStream struct {
	runes []rune
	pos   int
}

func newRuneStream(src string) *runeStream {
	return &runeStream{
		runes: []rune(src),
		pos:   0,
	}
}

func (r *runeStream) Next() rune {
	if r.pos >= len(r.runes) {
		return EOF
	}
	next := r.runes[r.pos]
	r.pos++
	return next
}

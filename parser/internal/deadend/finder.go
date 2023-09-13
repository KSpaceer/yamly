package deadend

import "github.com/KSpaceer/yamly/token"

const deadEndTriggerThreshold = 64

type Finder struct {
	m map[token.Position]int8
}

func NewFinder() Finder {
	return Finder{
		m: make(map[token.Position]int8),
	}
}

func (f *Finder) Mark(tok token.Token) bool {
	f.m[tok.Start]++
	return f.m[tok.Start] > deadEndTriggerThreshold
}

func (f *Finder) Reset() {
	clear(f.m)
}

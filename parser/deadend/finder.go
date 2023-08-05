package deadend

import "github.com/KSpaceer/yayamls/token"

const deadEndTriggerThreshold = 10

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
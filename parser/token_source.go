package parser

import (
	"github.com/KSpaceer/yayamls/pkg/cpaccessor"
	"github.com/KSpaceer/yayamls/token"
)

type RawTokenModer interface {
	SetRawMode()
	UnsetRawMode()
}

type ConfigurableTokenStream interface {
	cpaccessor.ResourceStream[token.Token]
	RawTokenModer
}

type tokenSource struct {
	cpaccessor.CheckpointingAccessor[token.Token]
	RawTokenModer
}

func newTokenSource(cts ConfigurableTokenStream) *tokenSource {
	return &tokenSource{
		CheckpointingAccessor: cpaccessor.NewCheckpointingAccessor[token.Token](cts),
		RawTokenModer:         cts,
	}
}

type simpleTokenStream struct {
	tokens []token.Token
	index  int
}

func newSimpleTokenStream(tokens []token.Token) ConfigurableTokenStream {
	return &simpleTokenStream{
		tokens: tokens,
		index:  0,
	}
}

func (t *simpleTokenStream) Next() token.Token {
	if t.index >= len(t.tokens) {
		return token.Token{Type: token.EOFType}
	}
	tok := t.tokens[t.index]
	t.index++
	return tok
}

func (t *simpleTokenStream) SetRawMode() {}

func (*simpleTokenStream) UnsetRawMode() {}

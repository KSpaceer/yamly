package parser

import (
	"sync"

	"github.com/KSpaceer/yamly/engines/yayamls/pkg/cpaccessor"
	"github.com/KSpaceer/yamly/engines/yayamls/token"
)

// RawTokenModer is used to set "raw mode" in lexer,
// meaning it almost will not change it's state from incoming tokens.
type RawTokenModer interface {
	SetRawMode()
	UnsetRawMode()
}

// ConfigurableTokenStream represents a token stream with
// abilities to set checkpoints and change mode to raw tokens.
type ConfigurableTokenStream interface {
	cpaccessor.ResourceStream[token.Token]
	RawTokenModer
}

type tokenSource struct {
	*cpaccessor.CheckpointingAccessor[token.Token]
	RawTokenModer
}

var accessorsPool = sync.Pool{}

func getCheckpointingAccessor(cts ConfigurableTokenStream) *cpaccessor.CheckpointingAccessor[token.Token] {
	if ca, ok := accessorsPool.Get().(*cpaccessor.CheckpointingAccessor[token.Token]); ok {
		ca.SetStream(cts)
		return ca
	}
	ca := cpaccessor.NewCheckpointingAccessor[token.Token]()
	ca.SetStream(cts)
	return &ca
}

func newTokenSource(cts ConfigurableTokenStream) *tokenSource {
	return &tokenSource{
		CheckpointingAccessor: getCheckpointingAccessor(cts),
		RawTokenModer:         cts,
	}
}

func (ts *tokenSource) release() {
	ts.CheckpointingAccessor.Reset()
	accessorsPool.Put(&ts.CheckpointingAccessor)
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

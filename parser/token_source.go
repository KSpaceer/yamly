package parser

import (
	"github.com/KSpaceer/fastyaml/cpaccessor"
	"github.com/KSpaceer/fastyaml/token"
)

type RawTokenModer interface {
	SetRawMode()
	UnsetRawMode()
}

type ConfigurableTokenStream interface {
	cpaccessor.ResourceStream[token.Token]
	RawTokenModer
}

func newTokenSource(cts ConfigurableTokenStream) *tokenSource {
	return &tokenSource{
		CheckpointingAccessor: cpaccessor.NewCheckpointingAccessor[token.Token](cts),
		RawTokenModer:         cts,
	}
}

type tokenSource struct {
	cpaccessor.CheckpointingAccessor[token.Token]
	RawTokenModer
}

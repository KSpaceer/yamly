package lexer

import "github.com/KSpaceer/fastyaml/token"

type TokenStream interface {
	Next() token.Token
	Save()
	Rollback()
}

type RuneStream interface {
	EOF() bool
	Next() rune
}

type lexer struct{}

func Tokenize(src string) TokenStream {
	rs := NewRuneStream(src)
}

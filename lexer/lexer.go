package lexer

type TokenStream interface {
	EOF() bool
	Next() Token
}

type RuneStream interface {
	EOF() bool
	Next() rune
}

type Token struct{}

type lexer struct{}

func Tokenize(src string) TokenStream {
	rs := NewRuneStream(src)
}

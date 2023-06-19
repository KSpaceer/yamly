package lexer

import "github.com/KSpaceer/fastyaml/token"

type tokenizer struct {
	rs          RuneStream
	buf         []rune
	quoted      bool
	singleLine  bool
	startOfLine bool
}

func (t *tokenizer) EOF() bool {
	return t.rs.EOF()
}

func (t *tokenizer) Next() token.Token {
	// token, ok := t.tryGetToken()
	return token.Token{}
}

func (t *tokenizer) tryGetToken() (token.Token, bool) {
	if t.rs.EOF() {
		return token.Token{}, false
	}
	return token.Token{}, false

	var haveFormedToken bool
	for !haveFormedToken {
		r := t.rs.Next()
		switch r {
		}
	}
	return token.Token{}, false
}

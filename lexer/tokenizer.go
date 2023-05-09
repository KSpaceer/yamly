package lexer

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

func (t *tokenizer) Next() Token {
	token, ok := t.tryGetToken()
}

func (t *tokenizer) tryGetToken() (Token, bool) {
	if t.rs.EOF() {
		return Token{}, false
	}

	var haveFormedToken bool
	for !haveFormedToken {
		r := t.rs.Next()
		switch r {
		case ' ':
			if t.startOfLine || t.quoted {
				t.buf = append(t.buf, r)
			}

		}
	}
}

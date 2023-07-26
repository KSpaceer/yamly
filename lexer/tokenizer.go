package lexer

import (
	"github.com/KSpaceer/yayamls/cpaccessor"
	"github.com/KSpaceer/yayamls/token"
	"strings"
)

const lookaheadBufferPreallocationSize = 8

type tokenizer struct {
	ra  cpaccessor.CheckpointingAccessor[rune]
	ctx context
	pos token.Position

	lookaheadBuf  []rune
	lookbehindTok token.Token

	preparedToken    token.Token
	hasPreparedToken bool
}

func NewTokenStream(src string) TokenStream {
	t := &tokenizer{
		ra:           cpaccessor.NewCheckpointingAccessor[rune](newRuneStream(src)),
		ctx:          newContext(),
		lookaheadBuf: make([]rune, 0, lookaheadBufferPreallocationSize),
		pos: token.Position{
			Row: 1,
		},
	}
	return t
}

func (t *tokenizer) SetRawMode() {
	t.ctx.switchContext(rawContextType)
}

func (t *tokenizer) UnsetRawMode() {
	if t.ctx.currentType() == rawContextType {
		t.ctx.revertContext()
	}
}

func (t *tokenizer) Next() token.Token {
	if t.hasPreparedToken {
		t.lookbehindTok = t.preparedToken
		tok := t.preparedToken
		t.preparedToken = token.Token{}
		t.hasPreparedToken = false
		return tok
	}
	return t.emitToken()
}

func (t *tokenizer) emitToken() token.Token {
	tok := token.Token{}
	var originBuilder strings.Builder
	for {

		r := t.ra.Next()
		t.pos.Column++

		curPos := t.pos

		specialTok, ok := t.ctx.matchSpecialToken(t, r)
		if ok {
			if tok.Type != token.UnknownType {
				t.preparedToken = specialTok
				t.hasPreparedToken = true

				tok.Origin = originBuilder.String()
				tok.End = curPos
				tok.End.Column--
			} else {
				tok = specialTok
				t.lookbehindTok = tok
			}
			break
		}

		if tok.Type == token.UnknownType {
			tok.Type = token.StringType
			tok.Start = curPos
			t.lookbehindTok = tok
		}
		originBuilder.WriteRune(r)
	}
	return tok
}

func (t *tokenizer) lookahead(count int, predicate func([]rune) bool) bool {
	t.ra.SetCheckpoint()
	for i := 0; i < count; i++ {
		t.lookaheadBuf = append(t.lookaheadBuf, t.ra.Next())
	}
	result := predicate(t.lookaheadBuf)
	t.lookaheadBuf = t.lookaheadBuf[:0]
	t.ra.Rollback()
	return result
}

func (t *tokenizer) lookbehind(predicate func(token.Token) bool) bool {
	return predicate(t.lookbehindTok)
}

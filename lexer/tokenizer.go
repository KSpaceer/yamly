package lexer

import (
	"github.com/KSpaceer/yayamls/pkg/cpaccessor"
	"github.com/KSpaceer/yayamls/pkg/strslice"
	"github.com/KSpaceer/yayamls/token"
	"strings"
)

const lookaheadBufferPreallocationSize = 8

type Tokenizer struct {
	ra  cpaccessor.CheckpointingAccessor[rune]
	ctx context
	pos token.Position

	lookaheadBuf  []rune
	lookbehindTok token.Token

	preparedToken    token.Token
	hasPreparedToken bool
}

type tokenizerOpts struct {
	unsafe bool
}

type TokenizerOption func(*tokenizerOpts)

func WithUnsafe() TokenizerOption {
	return func(opts *tokenizerOpts) {
		opts.unsafe = true
	}
}

func NewTokenizer(src string, opts ...TokenizerOption) *Tokenizer {
	var o tokenizerOpts
	for _, opt := range opts {
		opt(&o)
	}
	t := &Tokenizer{
		ra:           cpaccessor.NewCheckpointingAccessor[rune](),
		ctx:          newContext(),
		lookaheadBuf: make([]rune, 0, lookaheadBufferPreallocationSize),
		pos: token.Position{
			Row: 1,
		},
	}
	var runeSrc []byte
	if o.unsafe {
		runeSrc = strslice.StringToBytesSlice(src)
	} else {
		runeSrc = []byte(src)
	}

	t.ra.SetStream(newRuneStream(runeSrc))
	return t
}

func (t *Tokenizer) SetRawMode() {
	t.ctx.setRawModeValue(true)
}

func (t *Tokenizer) UnsetRawMode() {
	t.ctx.setRawModeValue(false)
}

func (t *Tokenizer) Next() token.Token {
	if t.hasPreparedToken {
		t.lookbehindTok = t.preparedToken
		tok := t.preparedToken
		t.preparedToken = token.Token{}
		t.hasPreparedToken = false
		return tok
	}
	return t.emitToken()
}

func (t *Tokenizer) emitToken() token.Token {
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
				// decreasing column, because we are currently at rune right after
				// string token
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

func (t *Tokenizer) lookahead(count int, predicate func([]rune) bool) bool {
	t.ra.SetCheckpoint()
	for i := 0; i < count; i++ {
		t.lookaheadBuf = append(t.lookaheadBuf, t.ra.Next())
	}
	result := predicate(t.lookaheadBuf)
	t.lookaheadBuf = t.lookaheadBuf[:0]
	t.ra.Rollback()
	return result
}

func (t *Tokenizer) lookbehind(predicate func(token.Token) bool) bool {
	return predicate(t.lookbehindTok)
}

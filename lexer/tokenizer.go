package lexer

import (
	"github.com/KSpaceer/fastyaml/cpaccessor"
	"github.com/KSpaceer/fastyaml/token"
	"strings"
)

type scanningContext int8

const (
	baseContext scanningContext = iota
	commentContext
	blockObjectContext
	flowObjectContext
	multilineBlockStartContext
	singleQuoteContext
	doubleQuoteContext
)

type tokenizer struct {
	ra       cpaccessor.CheckpointingAccessor[rune]
	ctxStack []scanningContext
	pos      token.Position

	lookaheadBuf  []rune
	lookbehindTok token.Token

	preparedToken    token.Token
	hasPreparedToken bool
}

const lookaheadBufferPreallocationSize = 8

func NewTokenStream(src string) TokenStream {
	t := &tokenizer{
		ra:           cpaccessor.NewCheckpointingAccessor[rune](newRuneStream(src)),
		lookaheadBuf: make([]rune, 0, lookaheadBufferPreallocationSize),
		pos: token.Position{
			Row:    1,
			Column: 0,
		},
	}
	return t
}

func (t *tokenizer) Next() token.Token {
	if t.hasPreparedToken {
		tok := t.preparedToken
		t.preparedToken = token.Token{}
		t.hasPreparedToken = false
		return tok
	}
	ctx := baseContext
	if len(t.ctxStack) > 0 {
		ctx = t.ctxStack[len(t.ctxStack)-1]
	}

	var specialTokenMatcher func(*tokenizer, rune) (token.Token, bool)
	switch ctx {
	case baseContext:
		specialTokenMatcher = tryGetBaseSpecialToken
	default:
		return token.Token{}
	}
	return t.emitToken(specialTokenMatcher)
}

func (t *tokenizer) emitToken(specialTokenMatcher func(*tokenizer, rune) (token.Token, bool)) token.Token {
	tok := token.Token{
		Type:   0,
		Start:  t.pos,
		End:    token.Position{},
		Origin: "",
	}
	var originBuilder strings.Builder
	for {
		r := t.ra.Next()
		t.pos.Column++

		curPos := t.pos
		specialTok, ok := specialTokenMatcher(t, r)
		if ok {
			if tok.Type != token.UnknownType {
				t.preparedToken = specialTok
				t.hasPreparedToken = true

				tok.Origin = originBuilder.String()
				tok.End = curPos
			} else {
				tok = specialTok
			}
			break
		}

		if tok.Type == token.UnknownType {
			tok.Type = token.StringType
		}
		originBuilder.WriteRune(r)
	}
	return tok
}

func tryGetComment

func tryGetBaseSpecialToken(t *tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{
		Type:   0,
		Start:  t.pos,
		End:    token.Position{},
		Origin: "",
	}

	switch r {
	case EOF:
		tok.End = t.pos
		tok.Type = token.EOFType
		return tok, true
	case token.ByteOrderMarkCharacter:
		tok.End = t.pos
		tok.Type = token.BOMType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.SequenceEntryCharacter:
		lookaheadPred := func(runes []rune) bool {
			return token.IsWhitespaceChar(runes[0]) || token.IsLineBreakChar(runes[0])
		}

		if t.lookahead(1, lookaheadPred) && t.lookbehind(isNonWordTypedToken) {
			t.ctxStack = append(t.ctxStack, blockObjectContext)
			tok.End = t.pos
			tok.Type = token.SequenceEntryType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.MappingKeyCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return token.IsWhitespaceChar(runes[0])
		}) && t.lookbehind(isNonWordTypedToken) {
			tok.End = t.pos
			tok.Type = token.MappingKeyType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.MappingValueCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return token.IsWhitespaceChar(runes[0])
		}) {
			tok.End = t.pos
			tok.Type = token.MappingValueType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.SequenceStartCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			t.ctxStack = append(t.ctxStack, flowObjectContext)
			tok.End = t.pos
			tok.Type = token.SequenceStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}

		if t.lookahead(3, func(runes []rune) bool {
			return runes[0] == runes[1] && runes[1] == token.DirectiveEndCharacter &&
				token.IsWhitespaceChar(runes[2])
		}) && t.lookbehind(isNonWordTypedToken) {
			tok.Origin = string([]rune{r, t.ra.Next(), t.ra.Next()})
			t.pos.Column += 2
			tok.End = t.pos
			tok.Type = token.DirectiveEndType
			return tok, true
		}
	case token.MappingStartCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			t.ctxStack = append(t.ctxStack, flowObjectContext)
			tok.End = t.pos
			tok.Type = token.MappingStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.CommentCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			t.ctxStack = append(t.ctxStack, commentContext)
			tok.End = t.pos
			tok.Type = token.CommentType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.AnchorCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			tok.End = t.pos
			tok.Type = token.AnchorType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.AliasCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			tok.End = t.pos
			tok.Type = token.AliasType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.TagCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			tok.End = t.pos
			tok.Type = token.TagType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.LiteralCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			t.ctxStack = append(t.ctxStack, multilineBlockStartContext)
			tok.End = t.pos
			tok.Type = token.LiteralType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.FoldedCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			t.ctxStack = append(t.ctxStack, multilineBlockStartContext)
			tok.End = t.pos
			tok.Type = token.FoldedType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.SingleQuoteCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			t.ctxStack = append(t.ctxStack, singleQuoteContext)
			tok.End = t.pos
			tok.Type = token.SingleQuoteType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.DoubleQuoteCharacter:
		if t.lookbehind(isNonWordTypedToken) {
			t.ctxStack = append(t.ctxStack, doubleQuoteContext)
			tok.End = t.pos
			tok.Type = token.DoubleQuoteType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.DirectiveCharacter:
		tok.End = t.pos
		tok.Type = token.DirectiveType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.CarriageReturnCharacter:
		origin := []rune{r}
		if t.lookahead(1, func(runes []rune) bool {
			return runes[0] == token.LineFeedCharacter
		}) {
			origin = append(origin, t.ra.Next())
			t.pos.Column++
		}
		tok.End = t.pos
		t.pos.Column = 0
		t.pos.Row++
		tok.Type = token.LineBreakType
		tok.Origin = string(origin)
		return tok, true
	case token.LineFeedCharacter:
		tok.End = t.pos
		t.pos.Column = 0
		t.pos.Row++
		tok.Type = token.LineBreakType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.SpaceCharacter:
		tok.End = t.pos
		tok.Type = token.SpaceType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.TabCharacter:
		tok.End = t.pos
		tok.Type = token.TabType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.DocumentEndCharacter:
		if t.lookahead(3, func(runes []rune) bool {
			return runes[0] == runes[1] && runes[1] == token.DocumentEndCharacter &&
				token.IsWhitespaceChar(runes[2])
		}) && t.lookbehind(isNonWordTypedToken) {
			tok.Origin = string([]rune{r, t.ra.Next(), t.ra.Next()})
			t.pos.Column += 2
			tok.End = t.pos
			tok.Type = token.DocumentEndType
		}
	}
	return token.Token{}, false
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

func isNonWordTypedToken(t token.Token) bool {
	switch t.Type {
	case token.SpaceType, token.TabType, token.LineBreakType, token.UnknownType:
		return true
	default:
		return false
	}
}

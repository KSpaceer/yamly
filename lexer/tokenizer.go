package lexer

import (
	"github.com/KSpaceer/fastyaml/cpaccessor"
	"github.com/KSpaceer/fastyaml/token"
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
	ra            cpaccessor.CheckpointingAccessor[rune]
	ctxStack      []scanningContext
	pos           token.Position
	currentRune   rune
	lookaheadBuf  []rune
	lookbehindTok token.Token
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
	t.currentRune = t.ra.Next()
	return t
}

func (t *tokenizer) Next() token.Token {
	ctx := baseContext
	if len(t.ctxStack) > 0 {
		ctx = t.ctxStack[len(t.ctxStack)-1]
	}
	t.pos.Column++
	switch ctx {
	case baseContext:
		return t.emitBaseContextToken()
	}
	return token.Token{}
}

func (t *tokenizer) emitBaseContextToken() token.Token {
	tok := token.Token{
		Type:   0,
		Start:  t.pos,
		End:    token.Position{},
		Origin: "",
	}
	r := t.currentRune
	for {
		switch r {
		case EOF:
			tok.End = t.pos
			tok.Type = token.EOFType
			return tok
		case token.ByteOrderMarkCharacter:
			tok.End = t.pos
			tok.Type = token.BOMType
			tok.Origin = string([]rune{r})
			return tok
		case token.SequenceEntryCharacter:
			lookaheadPred := func(runes []rune) bool {
				return token.IsWhitespaceChar(runes[0]) || token.IsLineBreakChar(runes[0])
			}

			if t.lookahead(1, lookaheadPred) && t.lookbehind(isNonWordTypedToken) {
				t.ctxStack = append(t.ctxStack, blockObjectContext)
				tok.End = t.pos
				tok.Type = token.SequenceEntryType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.MappingKeyCharacter:
			if t.lookahead(1, func(runes []rune) bool {
				return token.IsWhitespaceChar(runes[0])
			}) && t.lookbehind(isNonWordTypedToken) {
				tok.End = t.pos
				tok.Type = token.MappingKeyType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.MappingValueCharacter:
			if t.lookahead(1, func(runes []rune) bool {
				return token.IsWhitespaceChar(runes[0])
			}) {
				tok.End = t.pos
				tok.Type = token.MappingValueType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.SequenceStartCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				t.ctxStack = append(t.ctxStack, flowObjectContext)
				tok.End = t.pos
				tok.Type = token.SequenceStartType
				tok.Origin = string([]rune{r})
				return tok
			}

			if t.lookahead(3, func(runes []rune) bool {
				return runes[0] == runes[1] && runes[1] == token.DirectiveEndCharacter &&
					token.IsWhitespaceChar(runes[2])
			}) && t.lookbehind(isNonWordTypedToken) {
				tok.Origin = string([]rune{r, t.ra.Next(), t.ra.Next()})
				t.pos.Column += 2
				tok.End = t.pos
				tok.Type = token.DirectiveEndType
			}

		case token.MappingStartCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				t.ctxStack = append(t.ctxStack, flowObjectContext)
				tok.End = t.pos
				tok.Type = token.MappingStartType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.CommentCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				t.ctxStack = append(t.ctxStack, commentContext)
				tok.End = t.pos
				tok.Type = token.CommentType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.AnchorCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				tok.End = t.pos
				tok.Type = token.AnchorType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.AliasCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				tok.End = t.pos
				tok.Type = token.AliasType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.TagCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				tok.End = t.pos
				tok.Type = token.TagType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.LiteralCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				t.ctxStack = append(t.ctxStack, multilineBlockStartContext)
				tok.End = t.pos
				tok.Type = token.LiteralType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.FoldedCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				t.ctxStack = append(t.ctxStack, multilineBlockStartContext)
				tok.End = t.pos
				tok.Type = token.FoldedType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.SingleQuoteCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				t.ctxStack = append(t.ctxStack, singleQuoteContext)
				tok.End = t.pos
				tok.Type = token.SingleQuoteType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.DoubleQuoteCharacter:
			if t.lookbehind(isNonWordTypedToken) {
				t.ctxStack = append(t.ctxStack, doubleQuoteContext)
				tok.End = t.pos
				tok.Type = token.DoubleQuoteType
				tok.Origin = string([]rune{r})
				return tok
			}
		case token.DirectiveCharacter:
			tok.End = t.pos
			tok.Type = token.DirectiveType
			tok.Origin = string([]rune{r})
			return tok
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
			return tok
		case token.LineFeedCharacter:
			tok.End = t.pos
			t.pos.Column = 0
			t.pos.Row++
			tok.Type = token.LineBreakType
			tok.Origin = string([]rune{r})
			return tok
		case token.SpaceCharacter:
			tok.End = t.pos
			tok.Type = token.SpaceType
			tok.Origin = string([]rune{r})
			return tok
		case token.TabCharacter:
			tok.End = t.pos
			tok.Type = token.TabType
			tok.Origin = string([]rune{r})
			return tok
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
		default:
			return token.Token{}
		}
	}

	return token.Token{}
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

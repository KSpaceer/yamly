package lexer

import (
	"github.com/KSpaceer/yamly/engines/yayamls/chars"
	"github.com/KSpaceer/yamly/engines/yayamls/token"
)

type contextType int8

const (
	blockContextType contextType = iota
	flowContextType
	commentContextType
	multilineBlockStartContextType
	singleQuoteContextType
	doubleQuoteContextType
	tagContextType
)

const contextStackPreallocationSize = 4

type context struct {
	ctxStack []contextType
	escaped  bool
	rawMode  bool
}

func newContext() context {
	return context{
		ctxStack: make([]contextType, 0, contextStackPreallocationSize),
	}
}

func (c *context) setRawModeValue(value bool) {
	c.rawMode = value
}

func (c *context) switchContext(ctxType contextType) {
	c.ctxStack = append(c.ctxStack, ctxType)
}

func (c *context) currentType() contextType {
	if len(c.ctxStack) > 0 {
		return c.ctxStack[len(c.ctxStack)-1]
	}
	return blockContextType
}

func (c *context) revertContext() {
	if len(c.ctxStack) > 0 {
		c.ctxStack = c.ctxStack[:len(c.ctxStack)-1]
	}
}

func (c *context) whitespaceRevertContext() {
	ctxType := c.currentType()
	for {
		switch ctxType {
		case tagContextType:
		default:
			return
		}
		c.revertContext()
		ctxType = c.currentType()
	}
}

func (c *context) lineBreakRevertContext() {
	ctxType := c.currentType()
	for {
		switch ctxType {
		case blockContextType,
			flowContextType,
			doubleQuoteContextType,
			singleQuoteContextType:
			return
		}

		c.revertContext()
		ctxType = c.currentType()
	}
}

func (c *context) matchSpecialToken(t *Tokenizer, r rune) (token.Token, bool) {
	if c.rawMode {
		return c.rawMatching(t, r)
	}

	switch c.currentType() {
	case blockContextType:
		return c.blockMatching(t, r)
	case flowContextType:
		return c.flowMatching(t, r)
	case commentContextType:
		return c.commentMatching(t, r)
	case multilineBlockStartContextType:
		return c.multilineBlockStartMatching(t, r)
	case singleQuoteContextType:
		return c.singleQuoteMatching(t, r)
	case doubleQuoteContextType:
		return c.doubleQuoteMatching(t, r)
	case tagContextType:
		return c.tagMatching(t, r)
	default:
		return c.baseMatching(t, r)
	}
}

func (c *context) blockMatching(t *Tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	switch r {
	case chars.SequenceEntryCharacter:
		lookaheadPred := func(runes []rune) bool {
			return chars.IsBlankChar(runes[0]) || runes[0] == EOF
		}

		if t.lookahead(1, lookaheadPred) && t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.SequenceEntryType
			tok.Origin = string([]rune{r})
			return tok, true
		}

		if t.lookahead(3, func(runes []rune) bool {
			return runes[0] == runes[1] && runes[1] == chars.DirectiveEndCharacter &&
				(chars.IsBlankChar(runes[2]) || runes[2] == EOF)
		}) && t.lookbehind(token.MayPrecedeWord) {
			tok.Origin = string([]rune{r, t.ra.Next(), t.ra.Next()})
			t.pos.Column += 2
			tok.End = t.pos
			tok.Type = token.DirectiveEndType
			return tok, true
		}
	case chars.MappingKeyCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return chars.IsBlankChar(runes[0]) || runes[0] == EOF
		}) && t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.MappingKeyType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.MappingValueCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return chars.IsBlankChar(runes[0]) || runes[0] == EOF
		}) {
			tok.End = t.pos
			tok.Type = token.MappingValueType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.SequenceStartCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(flowContextType)
			tok.End = t.pos
			tok.Type = token.SequenceStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}

	case chars.MappingStartCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(flowContextType)
			tok.End = t.pos
			tok.Type = token.MappingStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.CommentCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(commentContextType)
			tok.End = t.pos
			tok.Type = token.CommentType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.AnchorCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.AnchorType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.AliasCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.AliasType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.TagCharacter:
		c.switchContext(tagContextType)
		tok.End = t.pos
		tok.Type = token.TagType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.LiteralCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(multilineBlockStartContextType)
			tok.End = t.pos
			tok.Type = token.LiteralType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.FoldedCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(multilineBlockStartContextType)
			tok.End = t.pos
			tok.Type = token.FoldedType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.SingleQuoteCharacter:
		c.switchContext(singleQuoteContextType)
		tok.End = t.pos
		tok.Type = token.SingleQuoteType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.DoubleQuoteCharacter:
		c.switchContext(doubleQuoteContextType)
		tok.End = t.pos
		tok.Type = token.DoubleQuoteType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.DirectiveCharacter:
		tok.End = t.pos
		tok.Type = token.DirectiveType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.DocumentEndCharacter:
		if t.lookahead(3, func(runes []rune) bool {
			return runes[0] == runes[1] && runes[1] == chars.DocumentEndCharacter &&
				(chars.IsBlankChar(runes[2]) || runes[2] == EOF)
		}) && t.lookbehind(token.MayPrecedeWord) {
			tok.Origin = string([]rune{r, t.ra.Next(), t.ra.Next()})
			t.pos.Column += 2
			tok.End = t.pos
			tok.Type = token.DocumentEndType
			return tok, true
		}
	}
	return c.baseMatching(t, r)
}

func (c *context) flowMatching(t *Tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}

	switch r {
	case chars.MappingKeyCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return chars.IsBlankChar(runes[0]) || runes[0] == EOF
		}) && t.lookbehind(mayPrecedeWordInFlow) {
			tok.End = t.pos
			tok.Type = token.MappingKeyType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.MappingValueCharacter:
		canBeAdjacent := func(tok token.Token) bool {
			return token.IsClosingFlowIndicator(tok) || tok.Type == token.DoubleQuoteType ||
				tok.Type == token.SingleQuoteType
		}

		if t.lookahead(1, func(runes []rune) bool {
			return chars.IsBlankChar(runes[0]) || chars.IsFlowIndicatorChar(runes[0]) || runes[0] == EOF
		}) || t.lookbehind(canBeAdjacent) {
			tok.End = t.pos
			tok.Type = token.MappingValueType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.SequenceStartCharacter:
		if t.lookbehind(mayPrecedeWordInFlow) {
			c.switchContext(flowContextType)
			tok.End = t.pos
			tok.Type = token.SequenceStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.MappingStartCharacter:
		if t.lookbehind(mayPrecedeWordInFlow) {
			c.switchContext(flowContextType)
			tok.End = t.pos
			tok.Type = token.MappingStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.SequenceEndCharacter:
		c.revertContext()
		tok.End = t.pos
		tok.Type = token.SequenceEndType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.MappingEndCharacter:
		c.revertContext()
		tok.End = t.pos
		tok.Type = token.MappingEndType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.AnchorCharacter:
		if t.lookbehind(mayPrecedeWordInFlow) {
			tok.End = t.pos
			tok.Type = token.AnchorType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.AliasCharacter:
		if t.lookbehind(mayPrecedeWordInFlow) {
			tok.End = t.pos
			tok.Type = token.AliasType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.TagCharacter:
		c.switchContext(tagContextType)
		tok.End = t.pos
		tok.Type = token.TagType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.SingleQuoteCharacter:
		if t.lookbehind(func(tok token.Token) bool {
			return mayPrecedeWordInFlow(tok) || tok.Type == token.MappingValueType
		}) {
			c.switchContext(singleQuoteContextType)
			tok.End = t.pos
			tok.Type = token.SingleQuoteType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.DoubleQuoteCharacter:
		if t.lookbehind(func(tok token.Token) bool {
			return mayPrecedeWordInFlow(tok) || tok.Type == token.MappingValueType
		}) {
			c.switchContext(doubleQuoteContextType)
			tok.End = t.pos
			tok.Type = token.DoubleQuoteType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case chars.CollectEntryCharacter:
		tok.End = t.pos
		tok.Type = token.CollectEntryType
		tok.Origin = string([]rune{r})
		return tok, true
	}
	return c.baseMatching(t, r)
}

func mayPrecedeWordInFlow(tok token.Token) bool {
	return token.MayPrecedeWord(tok) || token.IsOpeningFlowIndicator(tok)
}

func (c *context) commentMatching(t *Tokenizer, r rune) (token.Token, bool) {
	return c.baseMatching(t, r)
}

func (c *context) multilineBlockStartMatching(t *Tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	switch r {
	case chars.StripChompingCharacter:
		tok.End = t.pos
		tok.Type = token.StripChompingType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.KeepChompingCharacter:
		tok.End = t.pos
		tok.Type = token.KeepChompingType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.CommentCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(commentContextType)
			tok.End = t.pos
			tok.Type = token.CommentType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	}
	return c.baseMatching(t, r)
}

func (c *context) singleQuoteMatching(t *Tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	var escaped bool
	if r == chars.SingleQuoteCharacter {
		if !c.escaped && t.lookahead(1, func(runes []rune) bool {
			return runes[0] != chars.SingleQuoteCharacter
		}) {
			c.revertContext()
			tok.End = t.pos
			tok.Type = token.SingleQuoteType
			tok.Origin = string([]rune{r})
			return tok, true
		}
		escaped = !c.escaped
	}
	c.escaped = escaped
	return c.baseMatching(t, r)
}

func (c *context) doubleQuoteMatching(t *Tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	var escaped bool
	switch r {
	case chars.EscapeCharacter:
		if !c.escaped {
			escaped = true
		}
	case chars.SpaceCharacter:
		if c.escaped {
			c.escaped = false
			return token.Token{}, false
		}
	case chars.DoubleQuoteCharacter:
		if !c.escaped {
			c.revertContext()
			tok.End = t.pos
			tok.Type = token.DoubleQuoteType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	}
	c.escaped = escaped
	return c.baseMatching(t, r)
}

func (c *context) tagMatching(t *Tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	if r == chars.TagCharacter {
		tok.End = t.pos
		tok.Type = token.TagType
		tok.Origin = string([]rune{r})
		return tok, true
	}
	return c.baseMatching(t, r)
}

func (c *context) rawMatching(t *Tokenizer, r rune) (token.Token, bool) {
	return c.baseMatching(t, r)
}

func (c *context) baseMatching(t *Tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	switch r {
	case EOF:
		tok.End = t.pos
		tok.Type = token.EOFType
		return tok, true
	case chars.ByteOrderMarkCharacter:
		tok.End = t.pos
		tok.Type = token.BOMType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.CarriageReturnCharacter:
		c.lineBreakRevertContext()
		origin := []rune{r}
		if t.lookahead(1, func(runes []rune) bool {
			return runes[0] == chars.LineFeedCharacter
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
	case chars.LineFeedCharacter:
		c.lineBreakRevertContext()
		tok.End = t.pos
		t.pos.Column = 0
		t.pos.Row++
		tok.Type = token.LineBreakType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.SpaceCharacter:
		c.whitespaceRevertContext()
		tok.End = t.pos
		tok.Type = token.SpaceType
		tok.Origin = string([]rune{r})
		return tok, true
	case chars.TabCharacter:
		c.whitespaceRevertContext()
		tok.End = t.pos
		tok.Type = token.TabType
		tok.Origin = string([]rune{r})
		return tok, true
	}
	return token.Token{}, false
}

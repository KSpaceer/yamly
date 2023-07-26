package lexer

import "github.com/KSpaceer/yayamls/token"

type contextType int8

const (
	blockContextType contextType = iota
	flowContextType
	commentContextType
	multilineBlockStartContextType
	singleQuoteContextType
	doubleQuoteContextType
	rawContextType
)

const contextStackPreallocationSize = 4

type context struct {
	ctxStack []contextType
	escaped  bool
}

func newContext() context {
	return context{
		ctxStack: make([]contextType, 0, contextStackPreallocationSize),
	}
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

func (c *context) lineBreakRevertContext() {
	ctxType := c.currentType()
	for {
		switch ctxType {
		case blockContextType, flowContextType, doubleQuoteContextType, singleQuoteContextType, rawContextType:
			return
		}

		c.revertContext()
		ctxType = c.currentType()
	}
}

func (c *context) matchSpecialToken(t *tokenizer, r rune) (token.Token, bool) {
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
	case rawContextType:
		return c.rawMatching(t, r)
	default:
		return c.baseMatching(t, r)
	}
}

func (c *context) blockMatching(t *tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	switch r {
	case token.SequenceEntryCharacter:
		lookaheadPred := func(runes []rune) bool {
			return token.IsWhitespaceChar(runes[0]) || token.IsLineBreakChar(runes[0])
		}

		if t.lookahead(1, lookaheadPred) && t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.SequenceEntryType
			tok.Origin = string([]rune{r})
			return tok, true
		}

		if t.lookahead(3, func(runes []rune) bool {
			return runes[0] == runes[1] && runes[1] == token.DirectiveEndCharacter &&
				token.IsWhitespaceChar(runes[2])
		}) && t.lookbehind(token.MayPrecedeWord) {
			tok.Origin = string([]rune{r, t.ra.Next(), t.ra.Next()})
			t.pos.Column += 2
			tok.End = t.pos
			tok.Type = token.DirectiveEndType
			return tok, true
		}
	case token.MappingKeyCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return token.IsWhitespaceChar(runes[0])
		}) && t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.MappingKeyType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.MappingValueCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return token.IsWhitespaceChar(runes[0]) || token.IsLineBreakChar(runes[0])
		}) {
			tok.End = t.pos
			tok.Type = token.MappingValueType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.SequenceStartCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(flowContextType)
			tok.End = t.pos
			tok.Type = token.SequenceStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}

	case token.MappingStartCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(flowContextType)
			tok.End = t.pos
			tok.Type = token.MappingStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.CommentCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(commentContextType)
			tok.End = t.pos
			tok.Type = token.CommentType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.AnchorCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.AnchorType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.AliasCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.AliasType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.TagCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.TagType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.LiteralCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(multilineBlockStartContextType)
			tok.End = t.pos
			tok.Type = token.LiteralType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.FoldedCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(multilineBlockStartContextType)
			tok.End = t.pos
			tok.Type = token.FoldedType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.SingleQuoteCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(singleQuoteContextType)
			tok.End = t.pos
			tok.Type = token.SingleQuoteType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.DoubleQuoteCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(doubleQuoteContextType)
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
	case token.DocumentEndCharacter:
		if t.lookahead(3, func(runes []rune) bool {
			return runes[0] == runes[1] && runes[1] == token.DocumentEndCharacter &&
				token.IsWhitespaceChar(runes[2])
		}) && t.lookbehind(token.MayPrecedeWord) {
			tok.Origin = string([]rune{r, t.ra.Next(), t.ra.Next()})
			t.pos.Column += 2
			tok.End = t.pos
			tok.Type = token.DocumentEndType
		}
	}
	return c.baseMatching(t, r)
}

func (c *context) flowMatching(t *tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}

	switch r {
	case token.MappingKeyCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return token.IsWhitespaceChar(runes[0])
		}) && t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.MappingKeyType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.MappingValueCharacter:
		if t.lookahead(1, func(runes []rune) bool {
			return token.IsWhitespaceChar(runes[0])
		}) || t.lookbehind(token.IsClosingFlowIndicator) {
			tok.End = t.pos
			tok.Type = token.MappingValueType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.SequenceStartCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(flowContextType)
			tok.End = t.pos
			tok.Type = token.SequenceStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.MappingStartCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			c.switchContext(flowContextType)
			tok.End = t.pos
			tok.Type = token.MappingStartType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.SequenceEndCharacter:
		c.revertContext()
		tok.End = t.pos
		tok.Type = token.SequenceEndType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.MappingEndCharacter:
		c.revertContext()
		tok.End = t.pos
		tok.Type = token.MappingEndType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.AnchorCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.AnchorType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.AliasCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.AliasType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.TagCharacter:
		if t.lookbehind(token.MayPrecedeWord) {
			tok.End = t.pos
			tok.Type = token.TagType
			tok.Origin = string([]rune{r})
			return tok, true
		}
	case token.SingleQuoteCharacter:
		c.switchContext(singleQuoteContextType)
		tok.End = t.pos
		tok.Type = token.SingleQuoteType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.DoubleQuoteCharacter:
		c.switchContext(singleQuoteContextType)
		tok.End = t.pos
		tok.Type = token.DoubleQuoteType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.CollectEntryCharacter:
		tok.End = t.pos
		tok.Type = token.CollectEntryType
		tok.Origin = string([]rune{r})
		return tok, true
	}
	return c.baseMatching(t, r)
}

func (c *context) commentMatching(t *tokenizer, r rune) (token.Token, bool) {
	return c.baseMatching(t, r)
}

func (c *context) multilineBlockStartMatching(t *tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	switch r {
	case token.StripChompingCharacter:
		tok.End = t.pos
		tok.Type = token.StripChompingType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.KeepChompingCharacter:
		tok.End = t.pos
		tok.Type = token.KeepChompingType
		tok.Origin = string([]rune{r})
		return tok, true
	case token.CommentCharacter:
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

func (c *context) singleQuoteMatching(t *tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	var escaped bool
	if r == token.SingleQuoteCharacter {
		if !c.escaped && t.lookahead(1, func(runes []rune) bool {
			return runes[0] != token.SingleQuoteCharacter
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

func (c *context) doubleQuoteMatching(t *tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
	var escaped bool
	switch r {
	case token.EscapeCharacter:
		if !c.escaped {
			escaped = true
		}
	case token.DoubleQuoteCharacter:
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

func (c *context) rawMatching(t *tokenizer, r rune) (token.Token, bool) {
	return c.baseMatching(t, r)
}

func (c *context) baseMatching(t *tokenizer, r rune) (token.Token, bool) {
	tok := token.Token{Start: t.pos}
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
	case token.CarriageReturnCharacter:
		c.lineBreakRevertContext()
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
		c.lineBreakRevertContext()
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
	}
	return token.Token{}, false
}

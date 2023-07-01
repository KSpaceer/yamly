package parser

import (
	"github.com/KSpaceer/fastyaml/lexer"
	"github.com/KSpaceer/fastyaml/token"
)

type TokenAccessor struct {
	ts               lexer.TokenStream
	buf              []token.Token
	saved            token.Token
	bufIndex         int
	checkpointsStack []int
}

const (
	tokenBufferPreallocationSize      = 8
	checkpointsStackPreallocationSize = 2
)

func NewTokenAccessor(ts lexer.TokenStream) TokenAccessor {
	return TokenAccessor{
		ts:               ts,
		buf:              make([]token.Token, 0, tokenBufferPreallocationSize),
		bufIndex:         -1,
		checkpointsStack: make([]int, 0, checkpointsStackPreallocationSize),
	}
}

func (a *TokenAccessor) Next() token.Token {
	var tok token.Token
	if a.bufIndex != -1 && a.bufIndex != len(a.buf) {
		tok = a.buf[a.bufIndex]
		a.bufIndex++
		if a.bufIndex == len(a.buf) {
			if len(a.checkpointsStack) == 0 {
				a.buf = a.buf[:0]
			}
			a.bufIndex = -1
		}
	} else {
		tok = a.ts.Next()
		if len(a.checkpointsStack) > 0 {
			a.buf = append(a.buf, tok)
		} else {
			a.saved = tok
		}
	}

	return tok
}

func (a *TokenAccessor) SetCheckpoint() {
	if a.bufIndex != -1 {
		a.checkpointsStack = append(a.checkpointsStack, a.bufIndex)
	} else {
		a.checkpointsStack = append(a.checkpointsStack, len(a.buf)-1)
	}
}

func (a *TokenAccessor) Rollback() token.Token {
	switch stackLen := len(a.checkpointsStack); stackLen {
	case 0:
		return a.saved
	default:
		a.bufIndex = a.checkpointsStack[stackLen-1]
		var restoredToken token.Token
		if a.bufIndex == -1 {
			restoredToken = a.saved
			if len(a.buf) > 0 {
				a.bufIndex = 0
			}
		} else {
			restoredToken = a.buf[a.bufIndex]
			a.bufIndex++
			if a.bufIndex == len(a.buf) && len(a.checkpointsStack) == 0 {
				a.buf = a.buf[:0]
				a.bufIndex = -1
			}
		}
		a.checkpointsStack = a.checkpointsStack[:stackLen-1]
		return restoredToken
	}
}

func (a *TokenAccessor) Commit() {
	switch stackLen := len(a.checkpointsStack); stackLen {
	case 0:
	case 1:
		a.buf = a.buf[:0]
		a.bufIndex = -1
		fallthrough
	default:
		a.checkpointsStack = a.checkpointsStack[:stackLen-1]
	}
}

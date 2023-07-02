package parser

import (
	"github.com/KSpaceer/fastyaml/lexer"
	"github.com/KSpaceer/fastyaml/token"
)

type TokenAccessor struct {
	ts               lexer.TokenStream
	buf              []token.Token
	saved            token.Token
	bufIndicator     int
	checkpointsStack []int
}

const (
	tokenBufferPreallocationSize      = 8
	checkpointsStackPreallocationSize = 2

	withoutBuffer = -1
)

func NewTokenAccessor(ts lexer.TokenStream) TokenAccessor {
	return TokenAccessor{
		ts:               ts,
		buf:              make([]token.Token, 0, tokenBufferPreallocationSize),
		bufIndicator:     withoutBuffer,
		checkpointsStack: make([]int, 0, checkpointsStackPreallocationSize),
	}
}

func (a *TokenAccessor) Next() token.Token {
	var tok token.Token
	if a.bufIndicator == withoutBuffer {
		tok = a.ts.Next()
		if len(a.checkpointsStack) > 0 {
			a.buf = append(a.buf, tok)
		} else {
			a.saved = tok
		}
	} else {
		tok = a.buf[a.bufIndicator]
		a.bufIndicator++
		if a.bufIndicator == len(a.buf) {
			if len(a.checkpointsStack) == 0 {
				a.buf = a.buf[:0]
			}
			a.bufIndicator = withoutBuffer
		}
	}
	return tok
}

func (a *TokenAccessor) SetCheckpoint() {
	if a.bufIndicator == withoutBuffer {
		a.checkpointsStack = append(a.checkpointsStack, len(a.buf)-1)
	} else {
		a.checkpointsStack = append(a.checkpointsStack, a.bufIndicator-1)
	}
}

func (a *TokenAccessor) Rollback() token.Token {
	switch stackLen := len(a.checkpointsStack); stackLen {
	case 0:
		return a.saved
	default:
		a.bufIndicator = a.checkpointsStack[stackLen-1]
		var restoredTok token.Token
		if a.bufIndicator == withoutBuffer {
			restoredTok = a.saved
			if len(a.buf) > 0 {
				a.bufIndicator = 0
			}
		} else {
			restoredTok = a.buf[a.bufIndicator]
			a.bufIndicator++
			if a.bufIndicator == len(a.buf) {
				if len(a.checkpointsStack) == 0 {
					a.buf = a.buf[:0]
				}
				a.bufIndicator = withoutBuffer
			}
		}
		a.checkpointsStack = a.checkpointsStack[:stackLen-1]
		return restoredTok
	}
}

func (a *TokenAccessor) Commit() {
	switch stackLen := len(a.checkpointsStack); stackLen {
	case 0:
	case 1:
		if a.bufIndicator == withoutBuffer && len(a.buf) != 0 {
			a.saved = a.buf[len(a.buf)-1]
		}
		a.bufIndicator = -1
		a.buf = a.buf[:0]
		fallthrough
	default:
		a.checkpointsStack = a.checkpointsStack[:stackLen-1]
	}
}

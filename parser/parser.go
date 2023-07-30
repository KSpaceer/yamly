package parser

import (
	"bytes"
	"errors"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/lexer"
	"github.com/KSpaceer/yayamls/token"
	"sync"
)

type Context int8

const (
	NoContext Context = iota
	BlockInContext
	BlockOutContext
	BlockKeyContext
	FlowInContext
	FlowOutContext
	FlowKeyContext
)

var bufsPool = sync.Pool{
	New: func() any { return bytes.NewBuffer(nil) },
}

type IndentationMode int8

const (
	Unknown IndentationMode = iota
	StrictEquality
	WithLowerBound
)

type indentation struct {
	value int
	mode  IndentationMode
}

type parser struct {
	tokSrc      *tokenSource
	tok         token.Token
	savedStates []state
	state
	errors []error
}

type state struct {
	startOfLine bool
}

func newParser(tokSrc *tokenSource) *parser {
	return &parser{
		tokSrc: tokSrc,
		state: state{
			startOfLine: true,
		},
	}
}

func ParseTokenStream(cts ConfigurableTokenStream) (ast.Node, error) {
	p := newParser(newTokenSource(cts))
	return p.Parse()
}

func ParseTokens(tokens []token.Token) (ast.Node, error) {
	tokSrc := newTokenSource(newSimpleTokenStream(tokens))
	p := newParser(tokSrc)
	return p.Parse()
}

func ParseString(src string, opts ...ParseOption) (ast.Node, error) {
	o := applyOptions(opts...)
	var cts ConfigurableTokenStream
	if o.tokenStreamConstructor != nil {
		cts = o.tokenStreamConstructor(src)
	} else {
		cts = lexer.NewTokenizer(src)
	}
	return ParseTokenStream(cts)
}

func ParseBytes(src []byte, opts ...ParseOption) (ast.Node, error) {
	return ParseString(string(src), opts...)
}

func Parse(cts ConfigurableTokenStream) (ast.Node, error) {
	p := newParser(newTokenSource(cts))
	return p.Parse()
}

func (p *parser) Parse() (ast.Node, error) {
	p.next()
	p.startOfLine = true
	result := p.parseStream()
	return result, p.error()
}

func (p *parser) next() {
	p.startOfLine = isStartOfLine(p.startOfLine, p.tok)
	p.tok = p.tokSrc.Next()
}

func isStartOfLine(startOfLine bool, tok token.Token) bool {
	switch tok.Type {
	case token.LineBreakType:
		return true
	case token.BOMType:
		return startOfLine
	default:
		return false
	}
}

func (p *parser) appendError(err error) {
	p.errors = append(p.errors, err)
}

func (p *parser) hasErrors() bool {
	return len(p.errors) > 0
}

func (p *parser) error() error {
	return errors.Join(p.errors...)
}

func (p *parser) setCheckpoint() {
	p.tokSrc.SetCheckpoint()
	p.savedStates = append(p.savedStates, state{
		startOfLine: p.startOfLine,
	})
}

func (p *parser) commit() {
	p.tokSrc.Commit()
	if savedStatesLen := len(p.savedStates); savedStatesLen > 0 {
		p.savedStates = p.savedStates[:savedStatesLen-1]
	}
}

func (p *parser) rollback() {
	p.tok = p.tokSrc.Rollback()
	if savedStatesLen := len(p.savedStates); savedStatesLen > 0 {
		p.state = p.savedStates[savedStatesLen-1]
		p.savedStates = p.savedStates[:savedStatesLen-1]
	}
}

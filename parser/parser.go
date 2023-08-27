package parser

import (
	"bytes"
	"errors"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/lexer"
	"github.com/KSpaceer/yayamls/parser/internal/balancecheck"
	"github.com/KSpaceer/yayamls/parser/internal/deadend"
	"github.com/KSpaceer/yayamls/token"
	"sync"
)

type context int8

const (
	noContext context = iota
	blockInContext
	blockOutContext
	blockKeyContext
	flowInContext
	flowOutContext
	flowKeyContext
)

var bufsPool = sync.Pool{
	New: func() any { return bytes.NewBuffer(nil) },
}

type indentationMode int8

const (
	unknownIndentationMode indentationMode = iota
	strictEqualityIndentationMode
	withLowerBoundIndentationMode
)

type indentation struct {
	value int
	mode  indentationMode
}

type parser struct {
	tokSrc      *tokenSource
	tok         token.Token
	savedStates []state
	state
	errors         []error
	balanceChecker balancecheck.BalanceChecker
	deadEndFinder  deadend.Finder
}

type state struct {
	startOfLine         bool
	balanceCheckMemento balancecheck.BalanceCheckerMemento
}

func newParser(tokSrc *tokenSource) *parser {
	return &parser{
		tokSrc: tokSrc,
		state: state{
			startOfLine: true,
		},
		balanceChecker: balancecheck.NewBalanceChecker([][2]token.Type{
			{token.SequenceStartType, token.SequenceEndType},
			{token.MappingStartType, token.MappingEndType},
		}),
		deadEndFinder: deadend.NewFinder(),
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
	tree, err := ParseTokenStream(cts)
	if err != nil {
		return nil, err
	}
	if o.omitStream && tree.Type() == ast.StreamType {
		stream := tree.(*ast.StreamNode)
		if len(stream.Documents()) == 1 {
			tree = stream.Documents()[0]
		}
	}
	return tree, nil
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
	switch p.tok.Type {
	case token.EOFType:
		if !p.balanceChecker.IsBalanced() {
			unbalanced, _ := p.balanceChecker.PeekLastUnbalanced()
			p.appendError(UnbalancedOpeningParenthesisError{
				Type:        tokenTypeToParenthesesType(unbalanced),
				ExpectedPos: p.tok.Start,
			})
		}
	case token.MappingStartType, token.SequenceStartType, token.SingleQuoteType, token.DoubleQuoteType:
		p.balanceChecker.Add(p.tok.Type)
	case token.MappingEndType, token.SequenceEndType:
		if !p.balanceChecker.Add(p.tok.Type) {
			p.appendError(UnbalancedClosingParenthesisError{p.tok})
		}
	}
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
		startOfLine:         p.startOfLine,
		balanceCheckMemento: p.balanceChecker.Memento(),
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
		p.balanceChecker.SetMemento(p.state.balanceCheckMemento)
	}
}

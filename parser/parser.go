package parser

import (
	"bytes"
	"errors"
	"github.com/KSpaceer/yamly/ast"
	"github.com/KSpaceer/yamly/lexer"
	"github.com/KSpaceer/yamly/parser/internal/balancecheck"
	"github.com/KSpaceer/yamly/parser/internal/deadend"
	"github.com/KSpaceer/yamly/pkg/strslice"
	"github.com/KSpaceer/yamly/token"
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

var parserPool = sync.Pool{}

func getParser(tokSrc *tokenSource) *parser {
	if p, ok := parserPool.Get().(*parser); ok {
		p.tokSrc = tokSrc
		return p
	}
	return newParser(tokSrc)
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
	defer p.tokSrc.release()
	return p.Parse()
}

func ParseTokens(tokens []token.Token) (ast.Node, error) {
	tokSrc := newTokenSource(newSimpleTokenStream(tokens))
	p := newParser(tokSrc)
	defer p.release()
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
	return ParseString(strslice.BytesSliceToString(src), opts...)
}

func Parse(cts ConfigurableTokenStream) (ast.Node, error) {
	p := newParser(newTokenSource(cts))
	defer p.release()
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

func (p *parser) release() {
	p.tokSrc.release()
	p.tokSrc = nil
	p.savedStates = p.savedStates[:0]
	p.balanceChecker.Reset()
	p.deadEndFinder.Reset()
	p.errors = p.errors[:0]
	p.state = state{startOfLine: true}
	parserPool.Put(p)
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

func newContentNode(properties ast.Node, content ast.Node) ast.Node {
	if ast.ValidNode(properties) {
		content = ast.NewContentNode(properties, content)
	}
	return content
}

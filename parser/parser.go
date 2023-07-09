package parser

import (
	"bytes"
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/cpaccessor"
	"github.com/KSpaceer/fastyaml/lexer"
	"github.com/KSpaceer/fastyaml/token"
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
	ta          cpaccessor.CheckpointingAccessor[token.Token]
	tok         token.Token
	savedStates []state
	state
}

type state struct {
	startOfLine bool
}

func Parse(ts lexer.TokenStream) ast.Node {
	p := newParser(ts)
	return p.Parse()
}

func newParser(ts lexer.TokenStream) *parser {
	return &parser{
		ta: cpaccessor.NewCheckpointingAccessor[token.Token](ts),
		state: state{
			startOfLine: true,
		},
	}
}

func (p *parser) Parse() ast.Node {
	p.next()
	p.startOfLine = true
	return p.parseStream()
}

func (p *parser) next() {
	p.startOfLine = isStartOfLine(p.startOfLine, p.tok)
	p.tok = p.ta.Next()
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

func (p *parser) setCheckpoint() {
	p.ta.SetCheckpoint()
	p.savedStates = append(p.savedStates, state{
		startOfLine: p.startOfLine,
	})
}

func (p *parser) commit() {
	p.ta.Commit()
	if savedStatesLen := len(p.savedStates); savedStatesLen > 0 {
		p.savedStates = p.savedStates[:savedStatesLen-1]
	}
}

func (p *parser) rollback() {
	p.tok = p.ta.Rollback()
	if savedStatesLen := len(p.savedStates); savedStatesLen > 0 {
		p.state = p.savedStates[savedStatesLen-1]
		p.savedStates = p.savedStates[:savedStatesLen-1]
	}
}

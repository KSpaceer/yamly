package parser

import (
	"bytes"

	"github.com/KSpaceer/yamly/engines/yayamls/ast"
	"github.com/KSpaceer/yamly/engines/yayamls/token"
)

// YAML specification: [80] s-separate
func (p *parser) parseSeparate(ind *indentation, ctx context) ast.Node {
	switch ctx {
	case blockInContext, blockOutContext, flowInContext, flowOutContext:
		return p.parseSeparateLines(ind)
	case blockKeyContext, flowKeyContext:
		return p.parseSeparateInLine()
	}
	return ast.NewInvalidNode()
}

// YAML specification: [81] s-separate-lines
func (p *parser) parseSeparateLines(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	if ast.ValidNode(p.parseComments()) && ast.ValidNode(p.parseFlowLinePrefix(ind)) {
		p.commit()
		return ast.NewBasicNode(ast.IndentType)
	}
	p.rollback()
	if !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.IndentType)
}

// YAML specification: [66] s-separate-in-line
func (p *parser) parseSeparateInLine() ast.Node {
	if p.hasErrors() || !token.IsWhiteSpace(p.tok) && !p.startOfLine {
		return ast.NewInvalidNode()
	}
	for token.IsWhiteSpace(p.tok) {
		p.next()
	}
	return ast.NewBasicNode(ast.IndentType)
}

// YAML specification: [63] s-indent
func (p *parser) parseIndent(ind *indentation) ast.Node {
	switch ind.mode {
	case strictEqualityIndentationMode:
		return p.parseIndentWithStrictEquality(ind.value)
	case withLowerBoundIndentationMode:
		node, ok := p.parseIndentWithLowerBound(ind.value).(*ast.IndentNode)
		if !ok || !ast.ValidNode(node) {
			return ast.NewInvalidNode()
		}
		ind.mode = strictEqualityIndentationMode
		ind.value = node.Indent()
		return node
	default:
		return ast.NewInvalidNode()
	}
}

func (p *parser) parseIndentWithStrictEquality(indentation int) ast.Node {
	if p.hasErrors() || indentation < 0 {
		return ast.NewInvalidNode()
	}
	for i := indentation; i > 0; i-- {
		if p.tok.Type != token.SpaceType {
			return ast.NewInvalidNode()
		}
		p.next()
	}
	return ast.NewIndentNode(indentation)
}

func (p *parser) parseIndentWithLowerBound(lowerBound int) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	var indent int
	for ; indent < lowerBound; indent++ {
		if p.tok.Type != token.SpaceType {
			return ast.NewInvalidNode()
		}
		p.next()
	}

	for p.tok.Type == token.SpaceType {
		indent++
		p.next()
	}

	return ast.NewIndentNode(indent)
}

// YAML specification: [64] s-indent-less-than
func (p *parser) parseIndentLessThan(indentation int) ast.Node {
	return p.parseBorderedIndent(indentation, 1)
}

// YAML specification: [65] s-indent-less-or-equal
func (p *parser) parseIndentLessThanOrEqual(indentation int) ast.Node {
	return p.parseBorderedIndent(indentation, 0)
}

func (p *parser) parseBorderedIndent(indentation, lowBorder int) ast.Node {
	if p.hasErrors() || indentation < lowBorder {
		return ast.NewInvalidNode()
	}
	var currentIndent int

	for indentation > lowBorder {
		if p.tok.Type != token.SpaceType {
			return ast.NewIndentNode(currentIndent)
		}
		p.next()
		currentIndent++
		indentation--
	}

	return ast.NewIndentNode(currentIndent)
}

// YAML specification: [70] l-empty
func (p *parser) parseEmpty(ind *indentation, ctx context, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	lp := p.parseLinePrefix(ind, ctx)
	if !ast.ValidNode(lp) {
		p.rollback()
		lp = p.parseIndentLessThan(ind.value)
		if !ast.ValidNode(lp) {
			return ast.NewInvalidNode()
		}
	} else {
		p.commit()
	}
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode()
	}
	buf.WriteString(p.tok.Origin)
	p.next()
	return ast.NewBasicNode(ast.IndentType)
}

// YAML specification: [68] s-block-line-prefix
func (p *parser) parseBlockLinePrefix(ind *indentation) ast.Node {
	return p.parseIndent(ind)
}

// YAML specification: [67] s-line-prefix
func (p *parser) parseLinePrefix(ind *indentation, ctx context) ast.Node {
	switch ctx {
	case blockOutContext, blockInContext:
		return p.parseBlockLinePrefix(ind)
	case flowOutContext, flowInContext:
		return p.parseFlowLinePrefix(ind)
	default:
		return ast.NewInvalidNode()
	}
}

// YAML specification: [69] s-flow-line-prefix
func (p *parser) parseFlowLinePrefix(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	indent := p.parseIndent(ind)
	if !ast.ValidNode(indent) {
		return ast.NewInvalidNode()
	}

	p.setCheckpoint()
	if !ast.ValidNode(p.parseSeparateInLine()) {
		p.rollback()
	} else {
		p.commit()
	}
	return ast.NewBasicNode(ast.IndentType)
}

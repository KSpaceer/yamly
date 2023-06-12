package parser

import (
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/lexer"
	"github.com/KSpaceer/fastyaml/token"
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

type parser struct {
	ts          lexer.TokenStream
	tok         token.Token
	startOfLine bool
}

func NewParser(ts lexer.TokenStream) *parser {
	return &parser{
		ts:          ts,
		startOfLine: true,
	}
}

func (p *parser) Parse() ast.Node {
	p.next()

}

func (p *parser) next() {
	p.tok = p.ts.Next()
}

func (p *parser) parseStream() ast.Node {
	start := p.tok.Start

	for ast.ValidNode(p.parseDocumentPrefix()) {
	}

	documents := make([]ast.Node, 0)
	if doc := p.parseAnyDocument(); ast.ValidNode(doc) {
		documents = append(documents, doc)
	}

docs:
	for {
		switch {
		case ast.ValidNode(p.parseDocumentSuffix()):
			for ast.ValidNode(p.parseDocumentSuffix()) {
			}
			for ast.ValidNode(p.parseDocumentPrefix()) {
			}
			doc := p.parseAnyDocument()
			if ast.ValidNode(doc) {
				documents = append(documents, doc)
			}
		case ast.ValidNode(p.parseComment()):
		default:
			doc := p.parseExplicitDoc()
			if ast.ValidNode(doc) {
				documents = append(documents, doc)
			} else if p.tok.Type != token.EOFType {
				return ast.NewInvalidNode(start, p.tok.Start)
			} else {
				break docs
			}
		}
	}

	return &ast.StreamNode{
		StartPos:  start,
		EndPos:    p.tok.End,
		Documents: documents,
	}
}

func (p *parser) parseAnyDocument() ast.Node {
	start := p.tok.Start
	doc := p.parseDirectiveDocument()
	if ast.ValidNode(doc) {
		return doc
	}

	doc = p.parseExplicitDocument()
	if ast.ValidNode(doc) {
		return doc
	}

	doc = p.parseBareDocument()
	if ast.ValidNode(doc) {
		return doc
	}

	return ast.NewInvalidNode(start, p.tok.End)
}

func (p *parser) parseDirectiveDocument() ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseDirective()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	for ast.ValidNode(p.parseDirective()) {
	}

	return p.parseExplicitDocument()
}

func (p *parser) parseExplicitDocument() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.DirectiveEndType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	p.next()

	doc := p.parseBareDocument()
	if ast.ValidNode(doc) {
		return doc
	}
	for ast.ValidNode(p.parseComment()) {
	}
	return ast.NewBasicNode(start, p.tok.End, ast.DocumentType)
}

func (p *parser) parseBareDocument() ast.Node {
	return p.parseBlockNode(-1, BlockInContext)
}

func (p *parser) parseBlockNode(indentation int, ctx Context) ast.Node {
	start := p.tok.Start
	blockInBlock := p.parseBlockInBlock(indentation, ctx)
	if ast.ValidNode(blockInBlock) {
		return blockInBlock
	}
	flowInBlock := p.parseFlowInBlock(indentation)
	if ast.ValidNode(flowInBlock) {
		return flowInBlock
	}
	return ast.NewInvalidNode(start, p.tok.End)
}

func (p *parser) parseBlockInBlock(indentation int, ctx Context) ast.Node {
	start := p.tok.Start
	scalar := p.parseBlockScalar(indentation, ctx)
	if ast.ValidNode(scalar) {
		return scalar
	}
	collection := p.parseBlockCollection(indentation, ctx)
	if ast.ValidNode(collection) {
		return collection
	}
	return ast.NewInvalidNode(start, p.tok.End)
}

func (p *parser) parseBlockScalar(indentation int, ctx Context) ast.Node {
	start := p.tok.Start
	p.parseSeparate(indentation+1, ctx)
	properties := p.parseProperties(indentation+1, ctx)

	switch p.tok.Type {
	case token.LiteralType:
	case token.FoldedType:
	default:
		return ast.NewInvalidNode(start, p.tok.End)
	}
}

func (p *parser) parseProperties(indentation int, ctx Context) ast.Node {
	var (
		tag, anchor ast.Node
		start       = p.tok.Start
	)
	switch p.tok.Type {
	case token.TagType:
		tag = ast.NewTagNode(p.tok)
		p.next()
	case token.AnchorType:
		anchor = ast.NewAnchorNode(p.tok)
		p.next()
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}

	if !ast.ValidNode(p.parseSeparate(indentation, ctx)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	switch {
	case p.tok.Type == token.TagType && tag == nil:
		tag = ast.NewTagNode(p.tok)
		p.next()
		if !ast.ValidNode(p.parseSeparate(indentation, ctx)) {
			return ast.NewInvalidNode(start, p.tok.End)
		}
	case p.tok.Type == token.AnchorType && anchor == nil:
		anchor = ast.NewAnchorNode(p.tok)
		p.next()
		if !ast.ValidNode(p.parseSeparate(indentation, ctx)) {
			return ast.NewInvalidNode(start, p.tok.End)
		}
	}

	return ast.NewPropertiesNode(start, p.tok.Start, tag, anchor)
}

func (p *parser) parseSeparate(indentation int, ctx Context) ast.Node {
	switch ctx {
	case BlockInContext, BlockOutContext, FlowInContext, FlowOutContext:
		return p.parseSeparateLines(indentation)
	case BlockKeyContext, FlowKeyContext:
		for token.IsWhiteSpace(p.tok) {
			p.next()
		}
	}
	return ast.NewInvalidNode(p.tok.Start, p.tok.End)
}

func (p *parser) parseSeparateLines(indentation int) ast.Node {
	start := p.tok.Start
	for ast.ValidNode(p.parseComment()) {
	}
	p.parseIndent(indentation)
	for token.IsWhiteSpace(p.tok) {
		p.next()
	}
	return ast.NewBasicNode(start, p.tok.End, ast.IndentType)
}

func (p *parser) parseIndent(indentation int) ast.Node {
	start := p.tok.Start
	switch {
	case indentation == 0:
		return ast.NewBasicNode(token.Position{}, token.Position{}, ast.IndentType)
	case indentation > 0:
		if p.tok.Type != token.SpaceType {
			return ast.NewInvalidNode(start, p.tok.End)
		}
		p.next()
		return p.parseIndent(indentation - 1)
	default:
		return ast.NewInvalidNode(start, p.tok.End)
	}
}

func (p *parser) parseDirective() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.DirectiveType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	p.next()
	var directiveNode ast.Node
	switch p.tok.Origin {
	case token.YAMLDirective:
		p.next()
		directiveNode = p.parseYAMLDirective()
	case token.TagDirective:
		directiveNode = p.parseTagDirective()
	default:
		directiveNode = p.parseReservedDirective()
	}
	if !ast.ValidNode(directiveNode) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	for ast.ValidNode(p.parseComment()) {
	}
	return ast.NewBasicNode(start, p.tok.End, ast.DirectiveType)
}

func (p *parser) parseReservedDirective() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.UnquotedStringType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	for {
		for token.IsWhiteSpace(p.tok) {
			p.next()
		}
		if p.tok.Type != token.UnquotedStringType {
			break
		}
		p.next()
	}
	return ast.NewBasicNode(start, p.tok.Start, ast.DirectiveType)
}

func (p *parser) parseTagDirective() ast.Node {
	start := p.tok.Start
	for token.IsWhiteSpace(p.tok) {
		p.next()
	}
	if !ast.ValidNode(p.parseTagHandle()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	for token.IsWhiteSpace(p.tok) {
		p.next()
	}
	if !ast.ValidNode(p.parseTagPrefix()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.DirectiveType)
}

func (p *parser) parseTagHandle() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.TagType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TagType)
}

func (p *parser) parseTagPrefix() ast.Node {
	switch p.tok.Type {
	case token.TagType:
		return ast.NewBasicNode(p.tok.Start, p.tok.End, ast.TagType)
	// TODO: global tags support (maybe)
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

func (p *parser) parseYAMLDirective() ast.Node {
	start := p.tok.Start
	for token.IsWhiteSpace(p.tok) {
		p.next()
	}
	if !ast.ValidNode(p.parseYAMLVersion()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.DirectiveType)
}

func (p *parser) parseYAMLVersion() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.DecimalNumberType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	p.next()
	if p.tok.Type != token.DotType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	if p.tok.Type != token.DecimalNumberType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.FloatNumberType)
}

func (p *parser) parseDocumentPrefix() ast.Node {
	start := p.tok.Start
	if p.tok.Type == token.BOMType {
		p.next()
	}
	for ast.ValidNode(p.parseComment()) {
	}
	return ast.NewBasicNode(start, p.tok.Start, ast.DocumentPrefixType)
}

func (p *parser) parseDocumentSuffix() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.DocumentEndType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	p.next()
	if !ast.ValidNode(p.parseComment()) {
		return ast.NewInvalidNode(start, p.tok.Start)
	}
	for ast.ValidNode(p.parseComment()) {
	}
	return ast.NewBasicNode(start, p.tok.Start, ast.DocumentSuffixType)
}

func (p *parser) parseComment() ast.Node {
	start := p.tok.Start

	for token.IsWhiteSpace(p.tok) {
		p.next()
	}

	if p.tok.Type == token.CommentType {
		p.next()
	}

	if p.tok.Type == token.LineBreakType {
		p.startOfLine = true
		p.next()
	} else {
		return ast.NewInvalidNode(start, p.tok.Start)
	}

	return ast.NewBasicNode(start, p.tok.Start, ast.CommentType)
}

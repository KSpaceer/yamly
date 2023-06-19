package parser

import (
	"bytes"
	"fmt"
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/lexer"
	"github.com/KSpaceer/fastyaml/token"
	"strconv"
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
	ta          TokenAccessor
	tok         token.Token
	startOfLine bool
}

func NewParser(ts lexer.TokenStream) *parser {
	return &parser{
		ta:          NewTokenAccessor(ts),
		startOfLine: true,
	}
}

func (p *parser) Parse() ast.Node {
	p.next()
	return ast.NewInvalidNode(token.Position{}, token.Position{})
}

func (p *parser) next() {
	p.tok = p.ta.Next()
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
			// TODO: doc := p.parseExplicitDoc()
			var doc ast.Node
			if ast.ValidNode(doc) {
				documents = append(documents, doc)
			} else if p.tok.Type != token.EOFType {
				return ast.NewInvalidNode(start, p.tok.Start)
			} else {
				break docs
			}
		}
	}

	return &ast.StreamNode{}
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
	// TODO: !!!
	// flowInBlock := p.parseFlowInBlock(indentation)
	var flowInBlock ast.Node
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
	// TODO: !!!
	// collection := p.parseBlockCollection(indentation, ctx)
	var collection ast.Node
	if ast.ValidNode(collection) {
		return collection
	}
	return ast.NewInvalidNode(start, p.tok.End)
}

func (p *parser) parseBlockScalar(indentation int, ctx Context) ast.Node {
	start := p.tok.Start
	p.parseSeparate(indentation+1, ctx)
	properties := p.parseProperties(indentation+1, ctx)
	_ = properties

	switch p.tok.Type {
	case token.LiteralType:
		return p.parseLiteral(indentation)
	case token.FoldedType:
		// TODO: !!!
		// return p.parseFolded(indentation)
		return ast.StreamNode{}
	default:
		return ast.NewInvalidNode(start, p.tok.End)
	}
}

func (p *parser) parseLiteral(indentation int) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.LiteralType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	header := p.parseBlockHeader()
	if !ast.ValidNode(header) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	castedHeader, ok := header.(ast.BlockHeaderNode)
	if !ok {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	content := p.parseLiteralContent(indentation+castedHeader.IndentationIndicator(), castedHeader.ChompingIndicator())
	// TODO: !!!
	_ = content
	return content
}

func (p *parser) parseLiteralContent(indentation int, chomping ast.ChompingType) ast.Node {
	start := p.tok.Start
	var literalBuf bytes.Buffer
	txt := p.parseLiteralText(indentation, &literalBuf)
	if ast.ValidNode(txt) {
		for ast.ValidNode(p.parseLiteralNext(indentation, &literalBuf)) {
		}
		if !ast.ValidNode(p.parseChompedLast(chomping, &literalBuf)) {
			return ast.NewInvalidNode(start, p.tok.End)
		}
	}
	chompedEmpty := p.parseChompedEmpty(indentation, chomping, &literalBuf)
	// TODO: !!!
	_ = chompedEmpty
	return chompedEmpty
}

func (p *parser) parseChompedLast(chomping ast.ChompingType, buf *bytes.Buffer) ast.Node {
	switch chomping {
	case ast.StripChompingType:
		switch p.tok.Type {
		case token.LineBreakType, token.EOFType:
		default:
			return ast.NewInvalidNode(p.tok.Start, p.tok.End)
		}
	case ast.ClipChompingType, ast.KeepChompingType:
		switch p.tok.Type {
		case token.LineBreakType:
			buf.WriteString(p.tok.Origin)
		case token.EOFType:
		default:
			return ast.NewInvalidNode(p.tok.Start, p.tok.End)
		}
	}
	start := p.tok.Start
	p.next()
	return ast.NewBasicNode(start, p.tok.Start, ast.TextType)
}

func (p *parser) parseChompedEmpty(indentation int, chomping ast.ChompingType, buf *bytes.Buffer) ast.Node {
	switch chomping {
	case ast.ClipChompingType, ast.StripChompingType:
		return p.parseStripEmpty(indentation)
	case ast.KeepChompingType:
		// TODO: !!!
		// return p.parseKeepEmpty(indentation, buf)
		return ast.StreamNode{}
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

func (p *parser) parseStripEmpty(indentation int) ast.Node {
	if !ast.ValidNode(p.parseIndentLessThan(indentation)) {
		return ast.NewTextNode(token.Position{}, token.Position{}, nil)
	}
	for {
		if token.IsWhiteSpace(p.tok) {
			p.next()
		}
		if p.tok.Type != token.LineBreakType {
			break
		}
		p.next()
		if !ast.ValidNode(p.parseIndentLessThan(indentation)) {
			break
		}
	}
	// TODO: !!!
	return ast.StreamNode{}
}

func (p *parser) parseLiteralNext(indentation int, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	return p.parseLiteralText(indentation, buf)
}

func (p *parser) parseLiteralText(indentation int, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	for ast.ValidNode(p.parseEmpty(indentation, BlockInContext, buf)) {
	}

	if !ast.ValidNode(p.parseIndent(indentation)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	if p.tok.Type != token.StringType && !token.IsWhiteSpace(p.tok) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	for p.tok.Type == token.StringType || token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

func (p *parser) parseEmpty(indentation int, ctx Context, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	lp := p.parseLinePrefix(indentation, ctx)
	if !ast.ValidNode(lp) {
		lp = p.parseIndentLessThan(indentation)
		if !ast.ValidNode(lp) {
			return ast.NewInvalidNode(start, p.tok.End)
		}
	}
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf.WriteString(p.tok.Origin)
	p.next()
	return ast.NewBasicNode(start, p.tok.Start, ast.IndentType)
}

func (p *parser) parseLinePrefix(indentation int, ctx Context) ast.Node {
	switch ctx {
	case BlockOutContext, BlockInContext:
		return p.parseBlockLinePrefix(indentation)
	case FlowOutContext, FlowInContext:
		return p.parseFlowLinePrefix(indentation)
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

func (p *parser) parseBlockLinePrefix(indentation int) ast.Node {
	return p.parseIndent(indentation)
}

func (p *parser) parseFlowLinePrefix(indentation int) ast.Node {
	start := p.tok.Start
	indent := p.parseIndent(indentation)
	if !ast.ValidNode(indent) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	for token.IsWhiteSpace(p.tok) {
		p.next()
	}
	return ast.NewBasicNode(start, p.tok.End, ast.IndentType)
}

func (p *parser) parseBlockHeader() ast.Node {
	start := p.tok.Start

	chompingIndicator := p.parseChompingIndicator()
	indentationIndicator, err := p.parseIndentationIndicator()
	if err != nil {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if chompingIndicator == ast.ClipChompingType {
		chompingIndicator = p.parseChompingIndicator()
	}
	if !ast.ValidNode(p.parseComment()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBlockHeaderNode(start, p.tok.Start, chompingIndicator, indentationIndicator)
}

func (p *parser) parseChompingIndicator() ast.ChompingType {
	result := ast.TokenChompingType(p.tok)
	if result == ast.UnknownChompingType {
		return ast.ClipChompingType
	}
	p.next()
	return result
}

func (p *parser) parseIndentationIndicator() (int, error) {
	if p.tok.Type != token.StringType || !p.tok.ConformsCharSet(token.DecimalCharSetType) {
		return 0, nil
	}

	indentation, err := strconv.Atoi(p.tok.Origin)
	switch {
	case err != nil:
		return 0, fmt.Errorf("failed to parse indentation indicator node: %w", err)
	case indentation <= 0:
		return 0, fmt.Errorf("failed to parse indentation indicator node: " +
			"indentation must be omitted or greater than 0")
	default:
		p.next()
		return indentation, nil
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

func (p *parser) parseIndentLessThan(indentation int) ast.Node {
	start := p.tok.Start
	switch {
	case indentation == 1:
		return ast.NewBasicNode(token.Position{}, token.Position{}, ast.IndentType)
	case indentation > 1:
		if p.tok.Type == token.SpaceType {
			p.next()
			return p.parseIndentLessThan(indentation - 1)
		}
		return ast.NewBasicNode(token.Position{}, token.Position{}, ast.IndentType)
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
	if p.tok.Type != token.StringType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	for {
		for token.IsWhiteSpace(p.tok) {
			p.next()
		}
		if p.tok.Type != token.StringType {
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
	if p.tok.Type != token.StringType || !p.tok.ConformsCharSet(token.DecimalCharSetType) {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	p.next()
	if p.tok.Type != token.DotType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	if p.tok.Type != token.StringType || !p.tok.ConformsCharSet(token.DecimalCharSetType) {
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

package parser

import (
	"bytes"
	"fmt"
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/lexer"
	"github.com/KSpaceer/fastyaml/token"
	"strconv"
	"strings"
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
	savedStates []state
	state
}

type state struct {
	startOfLine bool
}

func NewParser(ts lexer.TokenStream) *parser {
	return &parser{
		ta: NewTokenAccessor(ts),
		state: state{
			startOfLine: true,
		},
	}
}

func (p *parser) Parse() ast.Node {
	p.next()
	return ast.NewInvalidNode(token.Position{}, token.Position{})
}

func (p *parser) next() {
	p.tok = p.ta.Next()
	p.startOfLine = p.tok.Type == token.LineBreakType
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
	p.ta.Rollback()
	if savedStatesLen := len(p.savedStates); savedStatesLen > 0 {
		p.state = p.savedStates[savedStatesLen-1]
		p.savedStates = p.savedStates[:savedStatesLen-1]
	}
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
	collection := p.parseBlockCollection(indentation, ctx)
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
		return p.parseFolded(indentation)
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
	castedContent, ok := content.(ast.TextNode)
	if !ok {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, castedContent.Text())
}

func (p *parser) parseLiteralContent(indentation int, chomping ast.ChompingType) ast.Node {
	start := p.tok.Start
	var literalBuf bytes.Buffer

	p.setCheckpoint()
	if ast.ValidNode(p.parseLiteralText(indentation, &literalBuf)) {
		for {
			p.setCheckpoint()
			if !ast.ValidNode(p.parseLiteralNext(indentation, &literalBuf)) {
				p.rollback()
				break
			}
			p.commit()
		}
		if !ast.ValidNode(p.parseChompedLast(chomping, &literalBuf)) {
			p.rollback()
		} else {
			p.commit()
		}
	} else {
		p.rollback()
	}
	if !ast.ValidNode(p.parseChompedEmpty(indentation, chomping, &literalBuf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, literalBuf.Bytes())
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
		return p.parseKeepEmpty(indentation, buf)
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

func (p *parser) parseKeepEmpty(indentation int, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(indentation, BlockInContext, buf)) {
			p.rollback()
			break
		}
		p.commit()
	}
	p.setCheckpoint()
	if !ast.ValidNode(p.parseTrailComments(indentation)) {
		p.rollback()
	} else {
		p.commit()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

func (p *parser) parseStripEmpty(indentation int) ast.Node {
	start := p.tok.Start
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseIndentLessThanOrEqual(indentation)) {
			p.rollback()
			break
		}
		if p.tok.Type != token.LineBreakType {
			p.rollback()
			break
		}
		p.commit()
	}

	p.setCheckpoint()
	if !ast.ValidNode(p.parseTrailComments(indentation)) {
		p.rollback()
	} else {
		p.commit()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

func (p *parser) parseTrailComments(indentation int) ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseIndentLessThan(indentation)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if !ast.ValidNode(p.parseCommentText()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if p.tok.Type != token.LineBreakType && p.tok.Type != token.EOFType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseCommentLine()) {
			p.rollback()
			break
		}
		p.commit()
	}
	return ast.NewBasicNode(start, p.tok.End, ast.CommentType)
}

func (p *parser) parseLiteralNext(indentation int, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	var localBuf bytes.Buffer
	localBuf.WriteString(p.tok.Origin)
	p.next()
	if !ast.ValidNode(p.parseLiteralText(indentation, &localBuf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf.Grow(localBuf.Len())
	buf.Write(localBuf.Bytes())
	return ast.NewTextNode(start, p.tok.End, nil)
}

func (p *parser) parseLiteralText(indentation int, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(indentation, BlockInContext, buf)) {
			p.rollback()
			break
		}
		p.commit()
	}
	if !ast.ValidNode(p.parseIndent(indentation)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if p.tok.Type != token.StringType && p.tok.Type != token.IsWhiteSpace(p.tok) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	for p.tok.Type == token.StringType || p.tok.Type == token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

func (p *parser) parseEmpty(indentation int, ctx Context, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	lp := p.parseLinePrefix(indentation, ctx)
	if !ast.ValidNode(lp) {
		p.rollback()
		lp = p.parseIndentLessThan(indentation)
		if !ast.ValidNode(lp) {
			return ast.NewInvalidNode(start, p.tok.End)
		}
	} else {
		p.commit()
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

	p.setCheckpoint()
	if !ast.ValidNode(p.parseSeparateInLine()) {
		p.rollback()
	} else {
		p.commit()
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
	p.setCheckpoint()
	if !ast.ValidNode(p.parseSeparatedComment()) {
		p.rollback()
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.commit()
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
		tag = p.parseTagProperty()
	case token.AnchorType:
		anchor = p.parseAnchorProperty()
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}

	if !ast.ValidNode(tag) && !ast.ValidNode(anchor) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparate(indentation, ctx)) {
		switch {
		case tag != nil:
			anchor = p.parseAnchorProperty()
		case anchor != nil:
			tag = p.parseTagProperty()
		}
	}
	if ast.ValidNode(tag) && ast.ValidNode(anchor) {
		p.commit()
	} else {
		p.rollback()
	}

	return ast.NewPropertiesNode(start, p.tok.Start, tag, anchor)
}

func (p *parser) parseAnchorProperty() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.AnchorType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.setCheckpoint()
	p.next()
	if p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.AnchorCharSetType) {
		anchor := ast.NewAnchorNode(start, p.tok.End, p.tok.Origin)
		p.next()
		p.commit()
		return anchor
	}

	p.rollback()
	return ast.NewInvalidNode(start, p.tok.End)
}

func (p *parser) parseTagProperty() ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	// shorthand tag
	if ast.ValidNode(p.parseTagHandle()) && p.tok.Type == token.StringType &&
		p.tok.ConformsCharSet(token.TagCharSetType) {
		p.commit()
		text := p.tok.Origin
		p.next()
		return ast.NewTagNode(start, p.tok.End, text)
	}
	p.rollback()

	if p.tok.Type != token.TagType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()

	// verbatim tag
	if p.tok.Type == token.StringType && strings.HasPrefix(p.tok.Origin, "<") && len(p.tok.Origin) > 2 {
		cutToken := token.Token{
			Type:   token.StringType,
			Start:  p.tok.Start,
			End:    p.tok.End,
			Origin: p.tok.Origin[1 : len(p.tok.Origin)-1],
		}
		if len(cutToken.Origin) > 0 && cutToken.ConformsCharSet(token.URICharSetType) &&
			p.tok.Origin[len(p.tok.Origin)-1] == '>' {
			p.next()
			return ast.NewTagNode(start, p.tok.End, cutToken.Origin)
		}
	}

	// non specific tag
	return ast.NewTagNode(start, p.tok.End, "")
}

func (p *parser) parseSeparate(indentation int, ctx Context) ast.Node {
	switch ctx {
	case BlockInContext, BlockOutContext, FlowInContext, FlowOutContext:
		return p.parseSeparateLines(indentation)
	case BlockKeyContext, FlowKeyContext:
		return p.parseSeparateInLine()
	}
	return ast.NewInvalidNode(p.tok.Start, p.tok.End)
}

func (p *parser) parseSeparateLines(indentation int) ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	if ast.ValidNode(p.parseComments()) && ast.ValidNode(p.parseFlowLinePrefix(indentation)) {
		p.commit()
		return ast.NewBasicNode(start, p.tok.End, ast.IndentType)
	}
	p.rollback()
	if !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.IndentType)
}

func (p *parser) parseSeparateInLine() ast.Node {
	if p.tok.Type != token.SpaceType && !p.startOfLine {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	start := p.tok.Start
	for p.tok.Type == token.SpaceType {
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
	return p.parseIndentWithLowerBound(indentation, 1)
}

func (p *parser) parseIndentLessThanOrEqual(indentation int) ast.Node {
	return p.parseIndentWithLowerBound(indentation, 0)
}

func (p *parser) parseIndentWithLowerBound(indentation int, lowerBound int) ast.Node {
	start := p.tok.Start
	switch {
	case indentation == lowerBound:
		return ast.NewBasicNode(token.Position{}, token.Position{}, ast.IndentType)
	case indentation > lowerBound:
		if p.tok.Type == token.SpaceType {
			p.next()
			return p.parseIndentWithLowerBound(indentation-1, lowerBound)
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
		p.next()
		directiveNode = p.parseTagDirective()
	default:
		directiveNode = p.parseReservedDirective()
	}
	if !ast.ValidNode(directiveNode) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode(start, p.tok.End)
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
		p.setCheckpoint()
		if !ast.ValidNode(p.parseSeparateInLine()) {
			p.rollback()
			break
		}
		if p.tok.Type != token.StringType {
			p.rollback()
			break
		}
		p.next()
		p.commit()
	}

	return ast.NewBasicNode(start, p.tok.Start, ast.DirectiveType)
}

func (p *parser) parseTagDirective() ast.Node {
	start := p.tok.Start
	for token.IsWhiteSpace(p.tok) {
		p.next()
	}
	if !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if !ast.ValidNode(p.parseTagHandle()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode(start, p.tok.End)
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
	p.next()
	if p.tok.Type == token.StringType {
		// either named or secondary tag handle
		p.setCheckpoint()
		// trying named handle
		cutToken := token.Token{
			Type:   token.StringType,
			Start:  p.tok.Start,
			End:    p.tok.End,
			Origin: p.tok.Origin[:len(p.tok.Origin)-1],
		}
		if cutToken.ConformsCharSet(token.WordCharSetType) || p.tok.Origin == "!" {
			p.commit()
			p.next()
		}
		// else - primary
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TagType)
}

func (p *parser) parseTagPrefix() ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	if ast.ValidNode(p.parseLocalTagPrefix()) {
		p.commit()
	} else {
		p.rollback()
		// trying global tag
		if !(p.tok.Type == token.StringType && len(p.tok.Origin) == 1 && p.tok.ConformsCharSet(token.TagCharSetType)) {
			return ast.NewInvalidNode(start, p.tok.End)
		}
		p.next()
		if p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.URICharSetType) {
			p.next()
		}
	}

	return ast.NewBasicNode(start, p.tok.End, ast.TagType)
}

func (p *parser) parseLocalTagPrefix() ast.Node {
	if p.tok.Type != token.TagType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	start := p.tok.Start
	p.next()
	if p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.URICharSetType) {
		p.next()
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TagType)
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
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseCommentLine()) {
			p.rollback()
			break
		}
		p.commit()
	}
	return ast.NewBasicNode(start, p.tok.Start, ast.DocumentPrefixType)
}

func (p *parser) parseDocumentSuffix() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.DocumentEndType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	p.next()
	p.setCheckpoint()
	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.Start, ast.DocumentSuffixType)
}

func (p *parser) parseComments() ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseSeparatedComment()) {
		p.rollback()
		if !p.startOfLine {
			return ast.NewInvalidNode(start, p.tok.End)
		}
	} else {
		p.commit()
	}
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseCommentLine()) {
			p.rollback()
			break
		}
		p.commit()
	}
	return ast.NewBasicNode(start, p.tok.End, ast.CommentType)
}

func (p *parser) parseSeparatedComment() ast.Node {
	start := p.tok.Start

	p.setCheckpoint()

	if !ast.ValidNode(p.parseSeparateInLine()) {
		p.rollback()
	} else {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseCommentText()) {
			p.rollback()
		} else {
			p.commit()
		}
		p.commit()
	}

	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.CommentType)
}

func (p *parser) parseCommentLine() ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.setCheckpoint()
	if !ast.ValidNode(p.parseCommentText()) {
		p.rollback()
	} else {
		p.commit()
	}
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.CommentType)
}

func (p *parser) parseCommentText() ast.Node {
	if p.tok.Type != token.CommentType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	start := p.tok.Start
	p.next()

	for p.tok.Type == token.StringType || p.tok.Type == token.SpaceType {
		p.next()
	}
	return ast.NewBasicNode(start, p.tok.Start, ast.CommentType)
}

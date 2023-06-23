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

// YAML specification: [211] l-yaml-stream
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

// YAML specification: [210] l-any-document
func (p *parser) parseAnyDocument() ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	doc := p.parseDirectiveDocument()
	if ast.ValidNode(doc) {
		p.commit()
		return doc
	}
	p.rollback()
	p.setCheckpoint()

	doc = p.parseExplicitDocument()
	if ast.ValidNode(doc) {
		p.commit()
		return doc
	}
	p.rollback()
	p.setCheckpoint()

	doc = p.parseBareDocument()
	if ast.ValidNode(doc) {
		p.commit()
		return doc
	}
	p.rollback()

	return ast.NewInvalidNode(start, p.tok.End)
}

// YAML specification: [209] l-directive-document
func (p *parser) parseDirectiveDocument() ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseDirective()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	for ast.ValidNode(p.parseDirective()) {
	}

	return p.parseExplicitDocument()
}

// YAML specification: [208] l-explicit-document
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

// YAML specification: [207] l-bare-document
func (p *parser) parseBareDocument() ast.Node {
	return p.parseBlockNode(&indentation{value: -1, mode: StrictEquality}, BlockInContext)
}

// YAML specification: [196] s-l+block-node
func (p *parser) parseBlockNode(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	blockInBlock := p.parseBlockInBlock(ind, ctx)
	if ast.ValidNode(blockInBlock) {
		p.commit()
		return blockInBlock
	}
	p.rollback()
	flowInBlock := p.parseFlowInBlock(ind)
	var flowInBlock ast.Node
	if ast.ValidNode(flowInBlock) {
		return flowInBlock
	}
	return ast.NewInvalidNode(start, p.tok.End)
}

// YAML specification: [197] s-l+flow-in-block
func (p *parser) parseFlowInBlock(ind *indentation) ast.Node {
	start := p.tok.Start

	localInd := indentation{
		value: ind.value + 1,
		mode:  StrictEquality,
	}
	if !ast.ValidNode(p.parseSeparate(&localInd, FlowOutContext)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	node := p.parseFlowNode(&localInd, FlowOutContext)
	if !ast.ValidNode(node) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return node
}

// YAML specification: [161] ns-flow-node
func (p *parser) parseFlowNode(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	node := p.parseAliasNode()
	if ast.ValidNode(node) {
		p.commit()
		return node
	}

	p.rollback()
	p.setCheckpoint()

	node = p.parseFlowContent(ind, ctx)
	if ast.ValidNode(node) {
		p.commit()
		return node
	}

	p.rollback()

	properties := p.parseProperties(ind, ctx)
	if !ast.ValidNode(properties) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	contentPos := p.tok.Start

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparate(ind, ctx)) {
		node = p.parseFlowContent(ind, ctx)
	}

	if ast.ValidNode(node) {
		p.commit()
		return ast.NewScalarNode(start, p.tok.End, properties, node)
	}

	p.rollback()
	return ast.NewScalarNode(start, p.tok.End, properties, ast.NewNullNode(contentPos))
}

// YAML specification: [198] s-l+block-in-block
func (p *parser) parseBlockInBlock(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	scalar := p.parseBlockScalar(ind, ctx)
	if ast.ValidNode(scalar) {
		p.commit()
		return scalar
	}
	p.rollback()

	collection := p.parseBlockCollection(ind, ctx)
	if ast.ValidNode(collection) {
		return collection
	}
	return ast.NewInvalidNode(start, p.tok.End)
}

// YAML specification: [200] s-l+block-collection
func (p *parser) parseBlockCollection(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start
	p.setCheckpoint()

	var properties ast.Node

	ind.value++
	if ast.ValidNode(p.parseSeparate(ind, ctx)) {
		properties = p.parseProperties(ind, ctx)
		if !ast.ValidNode(properties) {
			p.rollback()
			properties = nil
		}
	} else {
		p.rollback()
	}
	p.commit()
	ind.value--

	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	p.setCheckpoint()
	collection := p.parseSeqSpace(ind, ctx)
	if ast.ValidNode(collection) {
		p.commit()
		return ast.NewCollectionNode(start, p.tok.End, properties, collection)
	}
	p.rollback()
	collection = p.parseBlockMapping(ind)
	if !ast.ValidNode(collection) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewCollectionNode(start, p.tok.End, properties, collection)
}

// YAML specification: [187] l+block-mapping
func (p *parser) parseBlockMapping(ind *indentation) ast.Node {
	start := p.tok.Start
	localInd := indentation{
		value: ind.value + 1,
		mode:  WithLowerBound,
	}
	if !ast.ValidNode(p.parseIndent(&localInd)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entry := p.parseBlockMappingEntry(&localInd)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entries := []ast.Node{entry}

	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseIndent(&localInd)) {
			p.rollback()
			break
		}
		entry = p.parseBlockMappingEntry(&localInd)
		if !ast.ValidNode(entry) {
			p.rollback()
			break
		}
		p.commit()
		entries = append(entries, entry)
	}

	return ast.NewMappingNode(start, p.tok.End, entries)
}

func (p *parser) parseCompactMapping(ind *indentation) ast.Node {
	start := p.tok.Start

	entry := p.parseBlockMappingEntry(ind)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entries := []ast.Node{entry}

	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseIndent(ind)) {
			p.rollback()
			break
		}
		entry = p.parseBlockMappingEntry(ind)
		if !ast.ValidNode(entry) {
			p.rollback()
			break
		}
		p.commit()
		entries = append(entries, entry)
	}

	return ast.NewMappingNode(start, p.tok.End, entries)
}

// YAML specification: [188] ns-l-block-map-entry
func (p *parser) parseBlockMappingEntry(ind *indentation) ast.Node {

	switch p.tok.Type {
	case token.MappingKeyType:
		return p.parseBlockMappingExplicitEntry(ind)
	default:
		return p.parseBlockMappingImplicitEntry(ind)
	}
}

// YAML specification: [192] ns-l-block-map-implicit-entry
func (p *parser) parseBlockMappingImplicitEntry(ind *indentation) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	key := p.parseBlockMappingImplicitKey()
	if !ast.ValidNode(key) {
		p.rollback()
		key = ast.NewNullNode(start)
	} else {
		p.commit()
	}
	value := p.parseBlockMappingImplicitValue(ind)
	if !ast.ValidNode(value) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewMappingEntryNode(start, p.tok.End, key, value)
}

// YAML specification: [193] ns-s-block-map-implicit-key
func (p *parser) parseBlockMappingImplicitKey() ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	key := p.parseImplicitJSONKey(BlockKeyContext)
	if ast.ValidNode(key) {
		p.commit()
		return key
	}
	p.rollback()
	return p.parseImplicitYAMLKey(BlockKeyContext)
}

// YAML specification: [194] c-l-block-map-implicit-value
func (p *parser) parseBlockMappingImplicitValue(ind *indentation) ast.Node {
	if p.tok.Type != token.MappingValueType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	start := p.tok.Start
	p.next()

	p.setCheckpoint()
	value := p.parseBlockNode(ind, BlockOutContext)
	if ast.ValidNode(value) {
		p.commit()
		return value
	}
	p.rollback()
	value = ast.NewNullNode(p.tok.Start)
	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return value
}

// YAML specification: [189] c-l-block-map-explicit-entry
func (p *parser) parseBlockMappingExplicitEntry(ind *indentation) ast.Node {
	start := p.tok.Start

	key := p.parseBlockMappingExplicitKey(ind)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	valueStart := p.tok.Start
	p.setCheckpoint()
	value := p.parseBlockMappingExplicitValue(ind)
	if !ast.ValidNode(value) {
		p.rollback()
		value = ast.NewNullNode(valueStart)
	} else {
		p.commit()
	}

	return ast.NewMappingEntryNode(start, p.tok.End, key, value)
}

// YAML specification: [189] c-l-block-map-explicit-key
func (p *parser) parseBlockMappingExplicitKey(ind *indentation) ast.Node {
	if p.tok.Type != token.MappingKeyType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	p.next()

	return p.parseBlockIndented(ind, BlockOutContext)
}

// YAML specification: [189] l-block-map-explicit-value
func (p *parser) parseBlockMappingExplicitValue(ind *indentation) ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseIndent(ind)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if p.tok.Type != token.MappingValueType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	return p.parseBlockIndented(ind, BlockOutContext)
}

// YAML specification: [201] seq-space
func (p *parser) parseSeqSpace(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case BlockInContext:
		return p.parseBlockSequence(ind)
	case BlockOutContext:
		ind.value--
		node := p.parseBlockSequence(ind)
		ind.value++
		return node
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

// YAML specification: [183] l+block-sequence
func (p *parser) parseBlockSequence(ind *indentation) ast.Node {
	var entries []ast.Node

	start := p.tok.Start
	localInd := indentation{
		value: ind.value + 1,
		mode:  WithLowerBound,
	}
	if !ast.ValidNode(p.parseIndent(&localInd)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entry := p.parseBlockSequenceEntry(&localInd)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entries = append(entries, entry)

	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseIndent(&localInd)) {
			p.rollback()
			break
		}
		entry = p.parseBlockSequenceEntry(&localInd)
		if !ast.ValidNode(entry) {
			p.rollback()
			break
		}
		entries = append(entries, entry)
		p.commit()
	}

	return ast.NewSequenceNode(start, p.tok.End, entries)
}

// YAML specification: [184] c-l-block-seq-entry
func (p *parser) parseBlockSequenceEntry(ind *indentation) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.SequenceEntryType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	switch p.tok.Type {
	case token.SpaceType, token.TabType, token.LineBreakType:
		return p.parseBlockIndented(ind, BlockInContext)
	default:
		return ast.NewInvalidNode(start, p.tok.End)
	}
}

// YAML specification: [185] s-l+block-indented
func (p *parser) parseBlockIndented(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	localInd := indentation{
		value: 0,
		mode:  WithLowerBound,
	}
	// because we have "opportunistic" indentation starting with 0,
	// parseIndent will never return invalid node
	p.parseIndent(&localInd)
	p.setCheckpoint()
	mergedInd := indentation{
		value: ind.value + 1 + localInd.value,
		mode:  StrictEquality,
	}
	content := p.parseCompactSequence(&mergedInd)
	if ast.ValidNode(content) {
		p.commit()
		p.commit()
		return ast.NewBlockNode(start, p.tok.End, content)
	}
	p.rollback()

	content = p.parseCompactMapping(&mergedInd)
	if ast.ValidNode(content) {
		p.commit()
		return ast.NewBlockNode(start, p.tok.End, content)
	}
	p.rollback()

	p.setCheckpoint()
	content = p.parseBlockNode(ind, ctx)
	if ast.ValidNode(content) {
		p.commit()
		return content
	}
	p.rollback()

	nullStart := p.tok.Start

	if ast.ValidNode(p.parseComments()) {
		return ast.NewNullNode(nullStart)
	}
	return ast.NewInvalidNode(start, p.tok.End)
}

// YAML specification: [186] ns-l-compact-sequence
func (p *parser) parseCompactSequence(ind *indentation) ast.Node {
	start := p.tok.Start
	entry := p.parseBlockSequenceEntry(ind)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entries := []ast.Node{entry}

	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseIndent(ind)) {
			p.rollback()
			break
		}
		entry = p.parseBlockSequenceEntry(ind)
		if !ast.ValidNode(entry) {
			p.rollback()
			break
		}
		p.commit()
		entries = append(entries, entry)
	}

	return ast.NewSequenceNode(start, p.tok.End, entries)
}

// YAML specification: [199] s-l+block-scalar
func (p *parser) parseBlockScalar(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start
	ind.value++
	if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
		ind.value--
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.setCheckpoint()

	properties := p.parseProperties(ind, ctx)
	if ast.ValidNode(properties) {
		if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
			properties = nil
			p.rollback()
		} else {
			p.commit()
		}
	} else {
		p.rollback()
		properties = nil
	}

	var content ast.Node

	ind.value--
	switch p.tok.Type {
	case token.LiteralType:
		content = p.parseLiteral(ind)
	case token.FoldedType:
		content = p.parseFolded(ind)
	default:
		return ast.NewInvalidNode(start, p.tok.End)
	}

	if !ast.ValidNode(content) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewScalarNode(start, p.tok.End, properties, content)
}

// YAML specification: [182] c-l+folded
func (p *parser) parseFolded(ind *indentation) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.FoldedType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	header := p.parseBlockHeader()
	if !ast.ValidNode(header) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	castedHeader, ok := header.(ast.BlockHeaderNode)
	if !ok {
		ast.NewInvalidNode(start, p.tok.End)
	}
	content := p.parseFoldedContent(
		&indentation{
			value: ind.value + castedHeader.IndentationIndicator(),
			mode:  WithLowerBound,
		},
		castedHeader.ChompingIndicator(),
	)
	castedContent, ok := content.(ast.TextNode)
	if !ok {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, castedContent.Text())
}

// YAML specification: [182] l-folded-content
func (p *parser) parseFoldedContent(ind *indentation, chomping ast.ChompingType) ast.Node {
	start := p.tok.Start
	var foldedBuf bytes.Buffer

	p.setCheckpoint()
	var conditionalBuf bytes.Buffer
	if ast.ValidNode(p.parseDiffLines(ind, &conditionalBuf)) {
		if !ast.ValidNode(p.parseChompedLast(chomping, &conditionalBuf)) {
			p.rollback()
		} else {
			foldedBuf = conditionalBuf
			p.commit()
		}
	} else {
		p.rollback()
	}
	if !ast.ValidNode(p.parseChompedEmpty(ind, chomping, &foldedBuf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, foldedBuf.Bytes())
}

// YAML specification: [181] l-nb-diff-lines
func (p *parser) parseDiffLines(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseSameLines(ind, buf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	savedLen := buf.Len()
	for {
		p.setCheckpoint()
		if p.tok.Type != token.LineBreakType {
			buf.Truncate(savedLen)
			p.rollback()
			break
		}
		buf.WriteByte(byte(token.LineFeedCharacter))
		if !ast.ValidNode(p.parseSameLines(ind, buf)) {
			buf.Truncate(savedLen)
			p.rollback()
			break
		}
		p.commit()
		savedLen = buf.Len()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [180] l-nb-same-lines
func (p *parser) parseSameLines(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start

	savedLen := buf.Len()
	localLen := savedLen
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, BlockInContext, buf)) {
			p.rollback()
			buf.Truncate(localLen)
			break
		}
		localLen = buf.Len()
		p.commit()
	}
	p.setCheckpoint()

	if !ast.ValidNode(p.parseFoldedLines(ind, buf)) {
		buf.Truncate(localLen)
		p.rollback()
	} else {
		p.commit()
		return ast.NewTextNode(start, p.tok.End, nil)
	}

	if !ast.ValidNode(p.parseSpacedLines(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}

	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [179] l-nb-spaced-lines
func (p *parser) parseSpacedLines(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseSpacedText(ind, buf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	savedLen := buf.Len()
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseSpacedLineBreak(ind, buf)) {
			p.rollback()
			buf.Truncate(savedLen)
			break
		}
		if !ast.ValidNode(p.parseSpacedText(ind, buf)) {
			p.rollback()
			buf.Truncate(savedLen)
			break
		}
		p.commit()
		savedLen = buf.Len()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [177] s-nb-spaced-text
func (p *parser) parseSpacedText(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseIndent(ind)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if !token.IsWhiteSpace(p.tok) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf.WriteString(p.tok.Origin)
	p.next()
	for p.tok.Type == token.StringType || token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [178] b-l-spaced
func (p *parser) parseSpacedLineBreak(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf.WriteByte(byte(token.LineFeedCharacter))

	savedLen := buf.Len()

	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, BlockInContext, buf)) {
			p.rollback()
			buf.Truncate(savedLen)
			break
		}
		p.commit()
		savedLen = buf.Len()
	}

	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [176] l-nb-folded-lines
func (p *parser) parseFoldedLines(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseFoldedText(ind, buf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	savedLen := buf.Len()
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseFoldedLineBreak(ind, BlockInContext, buf)) {
			p.rollback()
			buf.Truncate(savedLen)
			break
		}
		if !ast.ValidNode(p.parseFoldedText(ind, buf)) {
			p.rollback()
			buf.Truncate(savedLen)
			break
		}
		p.commit()
		savedLen = buf.Len()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [175] s-nb-folded-text
func (p *parser) parseFoldedText(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseIndent(ind)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if p.tok.Type != token.StringType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf.WriteString(p.tok.Origin)
	p.next()
	for p.tok.Type == token.StringType || token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [73] b-l-folded
func (p *parser) parseFoldedLineBreak(ind *indentation, ctx Context, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start

	savedLen := buf.Len()
	p.setCheckpoint()

	if ast.ValidNode(p.parseTrimmed(ind, ctx, buf)) {
		p.commit()
		return ast.NewTextNode(start, p.tok.End, nil)
	}

	p.rollback()
	buf.Truncate(savedLen)

	// linebreak as space
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf.WriteByte(byte(token.SpaceCharacter))
	p.next()

	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [71] b-l-trimmed
func (p *parser) parseTrimmed(ind *indentation, ctx Context, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()

	savedLen := buf.Len()

	if !ast.ValidNode(p.parseEmpty(ind, ctx, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}

	savedLen = buf.Len()

	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, ctx, buf)) {
			buf.Truncate(savedLen)
			p.rollback()
			break
		}
		p.commit()
		savedLen = buf.Len()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [170] c-l+literal
func (p *parser) parseLiteral(ind *indentation) ast.Node {
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
	content := p.parseLiteralContent(
		&indentation{
			value: ind.value + castedHeader.IndentationIndicator(),
			mode:  WithLowerBound,
		},
		castedHeader.ChompingIndicator(),
	)
	castedContent, ok := content.(ast.TextNode)
	if !ok {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, castedContent.Text())
}

// YAML specification: [173] l-literal-content
func (p *parser) parseLiteralContent(ind *indentation, chomping ast.ChompingType) ast.Node {
	start := p.tok.Start
	var literalBuf bytes.Buffer

	p.setCheckpoint()
	var conditionalBuf bytes.Buffer
	if ast.ValidNode(p.parseLiteralText(ind, &conditionalBuf)) {
		for {
			p.setCheckpoint()
			if !ast.ValidNode(p.parseLiteralNext(ind, &conditionalBuf)) {
				p.rollback()
				break
			}
			p.commit()
		}
		if !ast.ValidNode(p.parseChompedLast(chomping, &conditionalBuf)) {
			p.rollback()
		} else {
			literalBuf = conditionalBuf
			p.commit()
		}
	} else {
		p.rollback()
	}
	if !ast.ValidNode(p.parseChompedEmpty(ind, chomping, &literalBuf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, literalBuf.Bytes())
}

// YAML specification: [165] b-chomped-last
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
			buf.WriteByte(byte(token.LineFeedCharacter))
		case token.EOFType:
		default:
			return ast.NewInvalidNode(p.tok.Start, p.tok.End)
		}
	}
	start := p.tok.Start
	p.next()
	return ast.NewBasicNode(start, p.tok.Start, ast.TextType)
}

// YAML specification: [166] l-chomped-empty
func (p *parser) parseChompedEmpty(ind *indentation, chomping ast.ChompingType, buf *bytes.Buffer) ast.Node {
	switch chomping {
	case ast.ClipChompingType, ast.StripChompingType:
		return p.parseStripEmpty(ind)
	case ast.KeepChompingType:
		return p.parseKeepEmpty(ind, buf)
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

// YAML specification: [168] l-keep-empty
func (p *parser) parseKeepEmpty(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start

	savedLen := buf.Len()
	localLen := savedLen
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, BlockInContext, buf)) {
			p.rollback()
			buf.Truncate(localLen)
			break
		}
		p.commit()
		localLen = buf.Len()
	}
	p.setCheckpoint()
	if !ast.ValidNode(p.parseTrailComments(ind)) {
		buf.Truncate(savedLen)
		p.rollback()
	} else {
		p.commit()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [167] l-strip-empty
func (p *parser) parseStripEmpty(ind *indentation) ast.Node {
	start := p.tok.Start
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseIndentLessThanOrEqual(ind.value)) {
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
	if !ast.ValidNode(p.parseTrailComments(ind)) {
		p.rollback()
	} else {
		p.commit()
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [169] l-trail-comments
func (p *parser) parseTrailComments(ind *indentation) ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseIndentLessThan(ind.value)) {
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

// YAML specification: [172] l-nb-literal-next
func (p *parser) parseLiteralNext(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	savedLen := buf.Len()
	buf.WriteByte(byte(token.LineFeedCharacter))
	p.next()
	if !ast.ValidNode(p.parseLiteralText(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, nil)
}

// YAML specification: [171] l-nb-literal-text
func (p *parser) parseLiteralText(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, BlockInContext, buf)) {
			p.rollback()
			break
		}
		p.commit()
	}
	if !ast.ValidNode(p.parseIndent(ind)) {
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

// YAML specification: [70] l-empty
func (p *parser) parseEmpty(ind *indentation, ctx Context, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	lp := p.parseLinePrefix(ind, ctx)
	if !ast.ValidNode(lp) {
		p.rollback()
		lp = p.parseIndentLessThan(ind.value)
		if !ast.ValidNode(lp) {
			return ast.NewInvalidNode(start, p.tok.End)
		}
	} else {
		p.commit()
	}
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf.WriteByte(byte(token.LineFeedCharacter))
	p.next()
	return ast.NewBasicNode(start, p.tok.Start, ast.IndentType)
}

// YAML specification: [67] s-line-prefix
func (p *parser) parseLinePrefix(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case BlockOutContext, BlockInContext:
		return p.parseBlockLinePrefix(ind)
	case FlowOutContext, FlowInContext:
		return p.parseFlowLinePrefix(ind)
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

// YAML specification: [68] s-block-line-prefix
func (p *parser) parseBlockLinePrefix(ind *indentation) ast.Node {
	return p.parseIndent(ind)
}

// YAML specification: [69] s-flow-line-prefix
func (p *parser) parseFlowLinePrefix(ind *indentation) ast.Node {
	start := p.tok.Start
	indent := p.parseIndent(ind)
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

// YAML specification: [162] c-b-block-header
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

// YAML specification: [164] c-chomping-indicator
func (p *parser) parseChompingIndicator() ast.ChompingType {
	result := ast.TokenChompingType(p.tok)
	if result == ast.UnknownChompingType {
		return ast.ClipChompingType
	}
	p.next()
	return result
}

// YAML specification: [163] c-indentation-indicator
func (p *parser) parseIndentationIndicator() (int, error) {
	if p.tok.Type != token.StringType || !p.tok.ConformsCharSet(token.DecimalCharSetType) {
		return 0, nil
	}

	ind, err := strconv.Atoi(p.tok.Origin)
	switch {
	case err != nil:
		return 0, fmt.Errorf("failed to parse indentation indicator node: %w", err)
	case ind <= 0:
		return 0, fmt.Errorf("failed to parse indentation indicator node: " +
			"indentation must be omitted or greater than 0")
	default:
		p.next()
		return ind, nil
	}
}

// YAML specification: [96] c-ns-properties
func (p *parser) parseProperties(ind *indentation, ctx Context) ast.Node {
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
	if ast.ValidNode(p.parseSeparate(ind, ctx)) {
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

// YAML specification: [101] c-ns-anchor-property
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

// YAML specification: [97] c-ns-tag-property
func (p *parser) parseTagProperty() ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	// shorthand tag
	// YAML specification: [99] c-ns-shorthand-tag
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
	// YAML specification: [98] c-verbatim-tag
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
	// YAML specification: [100] c-non-specific-tag
	return ast.NewTagNode(start, p.tok.End, "")
}

// YAML specification: [80] s-separate
func (p *parser) parseSeparate(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case BlockInContext, BlockOutContext, FlowInContext, FlowOutContext:
		return p.parseSeparateLines(ind)
	case BlockKeyContext, FlowKeyContext:
		return p.parseSeparateInLine()
	}
	return ast.NewInvalidNode(p.tok.Start, p.tok.End)
}

// YAML specification: [81] s-separate-lines
func (p *parser) parseSeparateLines(ind *indentation) ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	if ast.ValidNode(p.parseComments()) && ast.ValidNode(p.parseFlowLinePrefix(ind)) {
		p.commit()
		return ast.NewBasicNode(start, p.tok.End, ast.IndentType)
	}
	p.rollback()
	if !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.IndentType)
}

// YAML specification: [66] s-separate-in-line
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

// YAML specification: [63] s-indent
func (p *parser) parseIndent(ind *indentation) ast.Node {
	switch ind.mode {
	case StrictEquality:
		return p.parseIndentWithStrictEquality(ind.value)
	case WithLowerBound:
		node, ok := p.parseIndentWithLowerBound(ind.value).(ast.IndentNode)
		if !ok || !ast.ValidNode(node) {
			return ast.NewInvalidNode(node.Start(), node.End())
		}
		ind.mode = StrictEquality
		ind.value = node.Indent()
		return node
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

func (p *parser) parseIndentWithStrictEquality(indentation int) ast.Node {
	start := p.tok.Start
	if indentation < 0 {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	for i := indentation; i > 0; i-- {
		if p.tok.Type != token.SpaceType {
			return ast.NewInvalidNode(start, p.tok.End)
		}
		p.next()
	}
	return ast.NewIndentNode(start, p.tok.End, indentation)
}

func (p *parser) parseIndentWithLowerBound(lowerBound int) ast.Node {
	start := p.tok.Start
	var indent int
	for ; indent < lowerBound; indent++ {
		if p.tok.Type != token.SpaceType {
			return ast.NewInvalidNode(start, p.tok.End)
		}
		p.next()
	}

	for p.tok.Type == token.SpaceType {
		indent++
		p.next()
	}

	return ast.NewIndentNode(start, p.tok.End, indent)
}

// YAML specification: [64] s-indent-less-than
func (p *parser) parseIndentLessThan(indentation int) ast.Node {
	return p.parseBorderedIndent(indentation, 1)
}

// YAML specification: [65] s-indent-less-or-equal
func (p *parser) parseIndentLessThanOrEqual(indentation int) ast.Node {
	return p.parseBorderedIndent(indentation, 0)
}

func (p *parser) parseBorderedIndent(indentation int, lowBorder int) ast.Node {
	start := p.tok.Start
	if indentation < lowBorder {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	var currentIndent int

	for indentation > lowBorder {
		if p.tok.Type != token.SpaceType {
			return ast.NewIndentNode(start, p.tok.End, currentIndent)
		}
		p.next()
		currentIndent++
		indentation--
	}

	return ast.NewIndentNode(start, p.tok.End, currentIndent)
}

// YAML specification: [82] l-directive
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

// YAML specification: [83] ns-reserved-directive
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

// YAML specification: [88] ns-tag-directive
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

// YAML specification: [89] c-tag-handle
func (p *parser) parseTagHandle() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.TagType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	if p.tok.Type == token.StringType {
		// either named or secondary tag handle
		// YAML specification: [91] c-secondary-tag-handle
		// YAML specification: [92] c-named-tag-handle
		cutToken := token.Token{
			Type:   token.StringType,
			Start:  p.tok.Start,
			End:    p.tok.End,
			Origin: p.tok.Origin[:len(p.tok.Origin)-1],
		}
		// \w*!
		if p.tok.Origin[len(p.tok.Origin)-1] == byte(token.TagCharacter) &&
			(cutToken.ConformsCharSet(token.WordCharSetType) || len(cutToken.Origin) == 0) {
			p.next()
		}
		// else - primary
		// YAML specification: [90] c-primary-tag-handle
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TagType)
}

// YAML specification: [93] ns-tag-prefix
func (p *parser) parseTagPrefix() ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	if ast.ValidNode(p.parseLocalTagPrefix()) {
		p.commit()
	} else {
		p.rollback()
		// trying global tag
		// YAML specification: [95] ns-global-tag-prefix
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

// YAML specification: [94] c-ns-local-tag-prefix
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

// YAML specification: [86] ns-yaml-directive
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

// YAML specification: [87] ns-yaml-version
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

// YAML specification: [202] l-document-prefix
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

// YAML specification: [205] l-document-suffix
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

// YAML specification: [79] s-l-comments
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

// YAML specification: [77] s-b-comment
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

	if p.tok.Type != token.LineBreakType && p.tok.Type != token.EOFType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.CommentType)
}

// YAML specification: [78] l-comment
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

// YAML specification: [75] c-nb-comment-text
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

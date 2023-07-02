package parser

import (
	"bytes"
	"fmt"
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/lexer"
	"github.com/KSpaceer/fastyaml/token"
	"strconv"
	"strings"
	"sync"
	"unicode"
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
	ta          TokenAccessor
	tok         token.Token
	savedStates []state
	state
}

type state struct {
	startOfLine bool
}

func Parse(ts lexer.TokenStream) ast.Node {
	p := NewParser(ts)
	return p.Parse()
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
	p.startOfLine = true
	return p.parseStream()
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
	p.tok = p.ta.Rollback()
	if savedStatesLen := len(p.savedStates); savedStatesLen > 0 {
		p.state = p.savedStates[savedStatesLen-1]
		p.savedStates = p.savedStates[:savedStatesLen-1]
	}
}

// YAML specification: [211] l-yaml-stream
func (p *parser) parseStream() ast.Node {
	start := p.tok.Start

	for {
		p.setCheckpoint()
		if prefix := p.parseDocumentPrefix(); !ast.ValidNode(prefix) || prefix.Type() == ast.NullType {
			p.rollback()
			break
		}
		p.commit()
	}

	var docs []ast.Node

	p.setCheckpoint()
	doc := p.parseAnyDocument()
	if !ast.ValidNode(doc) {
		p.rollback()
	} else {
		p.commit()
		docs = append(docs, doc)
	}

	for {
		p.setCheckpoint()
		doc = p.parseDocumentWithSuffixesAndPrefixes()
		if ast.ValidNode(doc) {
			p.commit()
			docs = append(docs, doc)
			continue
		}
		p.rollback()

		if p.tok.Type == token.BOMType {
			continue
		}

		p.setCheckpoint()
		if ast.ValidNode(p.parseCommentLine()) {
			p.commit()
			continue
		}
		p.rollback()

		p.setCheckpoint()
		doc = p.parseExplicitDocument()
		if !ast.ValidNode(doc) {
			p.rollback()
			break
		}
		p.commit()
		docs = append(docs, doc)
	}

	return ast.NewStreamNode(start, p.tok.End, docs)
}

func (p *parser) parseDocumentWithSuffixesAndPrefixes() ast.Node {
	start := p.tok.Start
	if !ast.ValidNode(p.parseDocumentSuffix()) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseDocumentSuffix()) {
			p.rollback()
			break
		}
		p.commit()
	}

	p.setCheckpoint()
	doc := p.parseAnyDocument()
	if !ast.ValidNode(doc) {
		p.rollback()
		doc = ast.NewNullNode(p.tok.Start)
	} else {
		p.commit()
	}
	return doc
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

	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseDirective()) {
			p.rollback()
			break
		}
		p.commit()
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

	p.setCheckpoint()
	doc := p.parseBareDocument()
	if ast.ValidNode(doc) {
		p.commit()
		return doc
	}
	p.rollback()

	p.setCheckpoint()
	if !ast.ValidNode(p.parseComments()) {
		p.rollback()
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.commit()
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
	node := p.parseBlockInBlock(ind, ctx)
	if ast.ValidNode(node) {
		p.commit()
		return node
	}
	p.rollback()
	node = p.parseFlowInBlock(ind)
	if ast.ValidNode(node) {
		return node
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

// YAML specification: [159] ns-flow-yaml-node
func (p *parser) parseFlowYAMLNode(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	node := p.parseAliasNode()
	if ast.ValidNode(node) {
		p.commit()
		return node
	}

	p.rollback()
	p.setCheckpoint()

	node = p.parsePlain(ind, ctx)
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
		node = p.parsePlain(ind, ctx)
	}

	if ast.ValidNode(node) {
		p.commit()
		return ast.NewScalarNode(start, p.tok.End, properties, node)
	}

	p.rollback()
	return ast.NewScalarNode(start, p.tok.End, properties, ast.NewNullNode(contentPos))
}

// YAML specification: [160] c-flow-json-node
func (p *parser) parseFlowJSONNode(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	properties := p.parseProperties(ind, ctx)
	if ast.ValidNode(properties) && ast.ValidNode(p.parseSeparate(ind, ctx)) {
		p.commit()
	} else {
		p.rollback()
	}
	content := p.parseFlowJSONContent(ind, ctx)
	if !ast.ValidNode(content) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewScalarNode(start, p.tok.End, properties, content)
}

// YAML specification: [158] ns-flow-content
func (p *parser) parseFlowContent(ind *indentation, ctx Context) ast.Node {
	p.setCheckpoint()
	// YAML specification: [156] ns-flow-yaml-content
	content := p.parsePlain(ind, ctx)
	if ast.ValidNode(content) {
		p.commit()
		return content
	}
	p.rollback()
	return p.parseFlowJSONContent(ind, ctx)
}

// YAML specification: [157] c-flow-json-content
func (p *parser) parseFlowJSONContent(ind *indentation, ctx Context) ast.Node {
	switch p.tok.Type {
	case token.SequenceStartType:
		return p.parseFlowSequence(ind, ctx)
	case token.MappingStartType:
		return p.parseFlowMapping(ind, ctx)
	case token.SingleQuoteType:
		return p.parseSingleQuoted(ind, ctx)
	case token.DoubleQuoteType:
		return p.parseDoubleQuoted(ind, ctx)
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

// YAML specification: [109] c-double-quoted
func (p *parser) parseDoubleQuoted(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.DoubleQuoteType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	text := p.parseDoubleText(ind, ctx)
	if !ast.ValidNode(text) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	if p.tok.Type != token.SingleQuoteType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	return text
}

// YAML specification: [121] nb-single-text
func (p *parser) parseDoubleText(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case FlowInContext, FlowOutContext:
		return p.parseDoubleMultiLine(ind)
	case BlockKeyContext, FlowKeyContext:
		return p.parseDoubleOneLine()
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

// YAML specification: [111] nb-double-one-line
func (p *parser) parseDoubleOneLine() ast.Node {
	start := p.tok.Start

	buf := bufsPool.Get().(*bytes.Buffer)
	// escape sequences are handled by scanner
	for (p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.DoubleQuotedCharSetType)) ||
		token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
}

// YAML specification: [116] nb-double-multi-line
func (p *parser) parseDoubleMultiLine(ind *indentation) ast.Node {
	start := p.tok.Start

	buf := bufsPool.Get().(*bytes.Buffer)
	for (p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.DoubleQuotedCharSetType)) ||
		token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}

	savedLen := buf.Len()
	p.setCheckpoint()
	if !ast.ValidNode(p.parseDoubleNextLine(ind, buf)) {
		p.rollback()
		buf.Truncate(savedLen)

		p.next()
		for token.IsWhiteSpace(p.tok) {
			buf.WriteString(p.tok.Origin)
			p.next()
		}
	} else {
		p.commit()
	}

	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
}

// YAML specification: [115] s-double-next-line
func (p *parser) parseDoubleNextLine(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	savedLen := buf.Len()

	if !ast.ValidNode(p.parseDoubleBreak(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}

	p.setCheckpoint()

	if p.tok.Type != token.StringType || !p.tok.ConformsCharSet(token.DoubleQuotedCharSetType) {
		p.rollback()
		return ast.NewBasicNode(start, p.tok.End, ast.TextType)
	}
	p.commit()
	buf.WriteString(p.tok.Origin)
	p.next()

	for (p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.DoubleQuotedCharSetType)) ||
		token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}

	p.setCheckpoint()
	savedLen = buf.Len()

	if !ast.ValidNode(p.parseDoubleNextLine(ind, buf)) {
		p.rollback()
		buf.Truncate(savedLen)
		p.next()
		for token.IsWhiteSpace(p.tok) {
			buf.WriteString(p.tok.Origin)
			p.next()
		}
	} else {
		p.commit()
	}

	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
}

// YAML specification: [113] s-double-break
func (p *parser) parseDoubleBreak(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	p.setCheckpoint()
	savedLen := buf.Len()
	if ast.ValidNode(p.parseDoubleEscaped(ind, buf)) {
		p.commit()
		return ast.NewBasicNode(start, p.tok.End, ast.TextType)
	}
	p.rollback()
	buf.Truncate(savedLen)
	return p.parseFlowFolded(ind, buf)
}

// YAML specification: [112] s-double-escaped
func (p *parser) parseDoubleEscaped(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	savedLen := buf.Len()

	for token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}

	if p.tok.Type != token.StringType || p.tok.Origin != "\\" {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	if p.tok.Type != token.LineBreakType {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	for {
		p.setCheckpoint()
		savedLen := buf.Len()
		if !ast.ValidNode(p.parseEmpty(ind, FlowInContext, buf)) {
			buf.Truncate(savedLen)
			break
		}
		p.commit()
	}
	if !ast.ValidNode(p.parseFlowLinePrefix(ind)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
}

// YAML specification: [140] c-single-quoted
func (p *parser) parseSingleQuoted(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.SingleQuoteType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	text := p.parseSingleText(ind, ctx)
	if !ast.ValidNode(text) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	if p.tok.Type != token.SingleQuoteType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	return text
}

// YAML specification: [121] nb-single-text
func (p *parser) parseSingleText(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case FlowInContext, FlowOutContext:
		return p.parseSingleMultiLine(ind)
	case BlockKeyContext, FlowKeyContext:
		return p.parseSingleOneLine()
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

// YAML specification: [122] nb-single-one-line
func (p *parser) parseSingleOneLine() ast.Node {
	start := p.tok.Start

	buf := bufsPool.Get().(*bytes.Buffer)
	for (p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.SingleQuotedCharSetType)) ||
		token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
}

// YAML specification: [125] nb-single-multi-line
func (p *parser) parseSingleMultiLine(ind *indentation) ast.Node {
	start := p.tok.Start

	buf := bufsPool.Get().(*bytes.Buffer)
	for (p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.SingleQuotedCharSetType)) ||
		token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}

	savedLen := buf.Len()
	p.setCheckpoint()
	if !ast.ValidNode(p.parseSingleNextLine(ind, buf)) {
		p.rollback()
		buf.Truncate(savedLen)

		for token.IsWhiteSpace(p.tok) {
			buf.WriteString(p.tok.Origin)
			p.next()
		}
	} else {
		p.commit()
	}

	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
}

// YAML specification: [124] s-single-next-line
func (p *parser) parseSingleNextLine(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	savedLen := buf.Len()
	if !ast.ValidNode(p.parseFlowFolded(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.setCheckpoint()

	if p.tok.Type != token.StringType || !p.tok.ConformsCharSet(token.SingleQuotedCharSetType) {
		p.rollback()
		return ast.NewBasicNode(start, p.tok.End, ast.TextType)
	}
	p.commit()
	buf.WriteString(p.tok.Origin)
	p.next()

	for (p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.SingleQuotedCharSetType)) ||
		token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}

	p.setCheckpoint()
	savedLen = buf.Len()

	if !ast.ValidNode(p.parseSingleNextLine(ind, buf)) {
		p.rollback()
		buf.Truncate(savedLen)
		p.next()
		for token.IsWhiteSpace(p.tok) {
			buf.WriteString(p.tok.Origin)
			p.next()
		}
	} else {
		p.commit()
	}

	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
}

// YAML specification: [140] c-flow-mapping
func (p *parser) parseFlowMapping(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	if p.tok.Type != token.MappingStartType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparate(ind, ctx)) {
		p.commit()
	} else {
		p.rollback()
	}

	p.setCheckpoint()
	content := p.parseFlowMappingEntries(ind, ctx)
	if ast.ValidNode(content) {
		p.commit()
	} else {
		p.rollback()
		content = ast.NewNullNode(p.tok.Start)
	}

	if p.tok.Type != token.MappingEndType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return content
}

// YAML specification: [141] ns-s-flow-map-entries
func (p *parser) parseFlowMappingEntries(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	entry := p.parseFlowMappingEntry(ind, ctx)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entries := []ast.Node{entry}

	for {
		p.setCheckpoint()
		if ast.ValidNode(p.parseSeparate(ind, ctx)) {
			p.commit()
		} else {
			p.rollback()
		}

		if p.tok.Type != token.CollectEntryType {
			break
		}
		p.next()

		p.setCheckpoint()
		if ast.ValidNode(p.parseSeparate(ind, ctx)) {
			p.commit()
		} else {
			p.rollback()
		}

		p.setCheckpoint()
		entry = p.parseFlowMappingEntry(ind, ctx)
		if !ast.ValidNode(entry) {
			p.rollback()
			break
		}
		p.commit()
		entries = append(entries, entry)
	}

	return ast.NewMappingNode(start, p.tok.End, entries)
}

// YAML specification: [142] ns-flow-map-entry
func (p *parser) parseFlowMappingEntry(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	entry := p.parseFlowMappingImplicitEntry(ind, ctx)
	if ast.ValidNode(entry) {
		p.commit()
		return entry
	}
	p.rollback()

	if p.tok.Type != token.MappingKeyType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()

	if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entry = p.parseFlowMappingExplicitEntry(ind, ctx)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return entry
}

// YAML specification: [137] c-flow-sequence
func (p *parser) parseFlowSequence(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	if p.tok.Type != token.SequenceStartType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparate(ind, ctx)) {
		p.commit()
	} else {
		p.rollback()
	}

	p.setCheckpoint()
	content := p.parseInFlow(ind, ctx)
	if ast.ValidNode(content) {
		p.commit()
	} else {
		p.rollback()
		content = ast.NewNullNode(p.tok.Start)
	}

	if p.tok.Type != token.SequenceEndType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return content
}

// YAML specification: [136] in-flow
func (p *parser) parseInFlow(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case FlowInContext, FlowOutContext:
		ctx = FlowInContext
	case BlockKeyContext, FlowKeyContext:
		ctx = FlowKeyContext
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	return p.parseFlowSequenceEntries(ind, ctx)
}

// YAML specification: [138] ns-s-flow-seq-entries
func (p *parser) parseFlowSequenceEntries(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	entry := p.parseFlowSequenceEntry(ind, ctx)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	entries := []ast.Node{entry}

	for {
		p.setCheckpoint()
		if ast.ValidNode(p.parseSeparate(ind, ctx)) {
			p.commit()
		} else {
			p.rollback()
		}

		if p.tok.Type != token.CollectEntryType {
			break
		}
		p.next()

		p.setCheckpoint()
		if ast.ValidNode(p.parseSeparate(ind, ctx)) {
			p.commit()
		} else {
			p.rollback()
		}

		p.setCheckpoint()
		entry = p.parseFlowSequenceEntry(ind, ctx)
		if !ast.ValidNode(entry) {
			p.rollback()
			break
		}
		p.commit()
		entries = append(entries, entry)
	}

	return ast.NewSequenceNode(start, p.tok.End, entries)
}

// YAML specification: [139] ns-flow-seq-entry
func (p *parser) parseFlowSequenceEntry(ind *indentation, ctx Context) ast.Node {
	p.setCheckpoint()
	entry := p.parseFlowPair(ind, ctx)
	if ast.ValidNode(entry) {
		p.commit()
		return entry
	}
	p.rollback()
	return p.parseFlowNode(ind, ctx)
}

// YAML specification: [150] ns-flow-pair
func (p *parser) parseFlowPair(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	pair := p.parseFlowPairEntry(ind, ctx)
	if ast.ValidNode(pair) {
		p.commit()
		return pair
	}
	p.rollback()

	if p.tok.Type != token.MappingKeyType {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	p.next()
	if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	pair = p.parseFlowMappingExplicitEntry(ind, ctx)
	if !ast.ValidNode(pair) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return pair
}

// YAML specification: [151] ns-flow-pair-entry
func (p *parser) parseFlowPairEntry(ind *indentation, ctx Context) ast.Node {
	p.setCheckpoint()
	pair := p.parseFlowPairYAMLKeyEntry(ind, ctx)
	if ast.ValidNode(pair) {
		p.commit()
		return pair
	}

	p.rollback()
	p.setCheckpoint()

	pair = p.parseFlowMappingEmptyKeyEntry(ind, ctx)
	if ast.ValidNode(pair) {
		p.commit()
		return pair
	}

	p.rollback()
	return p.parseFlowPairJSONKeyEntry(ind, ctx)
}

// YAML specification: [153] c-ns-flow-pair-json-key-entry
func (p *parser) parseFlowPairJSONKeyEntry(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start
	key := p.parseImplicitJSONKey(FlowKeyContext)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	value := p.parseFlowMappingAdjacentValue(ind, ctx)
	if !ast.ValidNode(value) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewMappingEntryNode(start, p.tok.End, key, value)
}

// YAML specification: [155] c-s-implicit-json-key
func (p *parser) parseImplicitJSONKey(ctx Context) ast.Node {
	start := p.tok.Start

	localInd := indentation{
		value: 0,
		mode:  StrictEquality,
	}
	node := p.parseFlowJSONNode(&localInd, ctx)
	if !ast.ValidNode(node) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparateInLine()) {
		p.commit()
	} else {
		p.rollback()
	}
	return node
}

// YAML specification: [152] ns-flow-pair-yaml-key-entry
func (p *parser) parseFlowPairYAMLKeyEntry(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start
	key := p.parseImplicitYAMLKey(FlowKeyContext)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	value := p.parseFlowMappingSeparateValue(ind, ctx)
	if !ast.ValidNode(value) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewMappingEntryNode(start, p.tok.End, key, value)
}

// YAML specification: [154] ns-s-implicit-yaml-key
func (p *parser) parseImplicitYAMLKey(ctx Context) ast.Node {
	start := p.tok.Start

	localInd := indentation{
		value: 0,
		mode:  StrictEquality,
	}
	node := p.parseFlowYAMLNode(&localInd, ctx)
	if !ast.ValidNode(node) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparateInLine()) {
		p.commit()
	} else {
		p.rollback()
	}
	return node
}

// YAML specification: [143] ns-flow-map-explicit-entry
func (p *parser) parseFlowMappingExplicitEntry(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	entry := p.parseFlowMappingImplicitEntry(ind, ctx)
	if ast.ValidNode(entry) {
		p.commit()
		return entry
	}
	p.rollback()
	return ast.NewMappingEntryNode(
		start,
		start,
		ast.NewNullNode(start),
		ast.NewNullNode(start),
	)
}

// YAML specification: [144] ns-flow-map-implicit-entry
func (p *parser) parseFlowMappingImplicitEntry(ind *indentation, ctx Context) ast.Node {
	p.setCheckpoint()
	entry := p.parseFlowMappingYAMLKeyEntry(ind, ctx)
	if ast.ValidNode(entry) {
		p.commit()
		return entry
	}
	p.rollback()

	p.setCheckpoint()
	entry = p.parseFlowMappingJSONKeyEntry(ind, ctx)
	if ast.ValidNode(entry) {
		p.commit()
		return entry
	}
	p.rollback()

	return p.parseFlowMappingEmptyKeyEntry(ind, ctx)
}

// YAML specification: [146] c-ns-flow-map-empty-key-entry
func (p *parser) parseFlowMappingEmptyKeyEntry(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	value := p.parseFlowMappingSeparateValue(ind, ctx)
	if !ast.ValidNode(value) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewMappingEntryNode(start, p.tok.End, ast.NewNullNode(start), value)
}

// YAML specification: [148] c-ns-flow-map-json-key-entry
func (p *parser) parseFlowMappingJSONKeyEntry(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	key := p.parseFlowJSONNode(ind, ctx)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	p.setCheckpoint()

	p.setCheckpoint()
	if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
		p.rollback()
	} else {
		p.commit()
	}

	value := p.parseFlowMappingAdjacentValue(ind, ctx)
	if !ast.ValidNode(value) {
		p.rollback()
		value = ast.NewNullNode(p.tok.Start)
	} else {
		p.commit()
	}

	return ast.NewMappingEntryNode(start, p.tok.End, key, value)
}

// YAML specification: [149] c-ns-flow-map-adjacent-value
func (p *parser) parseFlowMappingAdjacentValue(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	if p.tok.Type != token.MappingValueType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()

	p.setCheckpoint()

	p.setCheckpoint()
	if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
		p.rollback()
	} else {
		p.commit()
	}
	value := p.parseFlowNode(ind, ctx)
	if ast.ValidNode(value) {
		p.commit()
	} else {
		p.rollback()
		value = ast.NewNullNode(p.tok.Start)
	}
	return value
}

// YAML specification: [145] ns-flow-map-yaml-key-entry
func (p *parser) parseFlowMappingYAMLKeyEntry(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	key := p.parseFlowYAMLNode(ind, ctx)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	p.setCheckpoint()

	p.setCheckpoint()
	if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
		p.rollback()
	} else {
		p.commit()
	}

	value := p.parseFlowMappingSeparateValue(ind, ctx)
	if !ast.ValidNode(value) {
		p.rollback()
		value = ast.NewNullNode(p.tok.Start)
	} else {
		p.commit()
	}
	return ast.NewMappingEntryNode(start, p.tok.End, key, value)
}

// YAML specification: [147] c-ns-flow-map-separate-value
func (p *parser) parseFlowMappingSeparateValue(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.MappingValueType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	// lookahead
	if isPlainSafeToken(p.tok, ctx) {
		return ast.NewInvalidNode(start, p.tok.End)
	}

	valueStart := p.tok.Start

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparate(ind, ctx)) {
		value := p.parseFlowNode(ind, ctx)
		if ast.ValidNode(value) {
			p.commit()
			return value
		}
	}
	p.rollback()

	return ast.NewNullNode(valueStart)
}

// YAML specification: [131] ns-plain
func (p *parser) parsePlain(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case FlowInContext, FlowOutContext:
		return p.parsePlainMultiLine(ind, ctx)
	case BlockKeyContext, FlowKeyContext:
		return p.parsePlainOneLine(ctx)
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

// YAML specification: [135] ns-plain-multi-line
func (p *parser) parsePlainMultiLine(ind *indentation, ctx Context) ast.Node {
	start := p.tok.Start

	firstLine, ok := p.parsePlainOneLine(ctx).(*ast.TextNode)
	if !ok || !ast.ValidNode(firstLine) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf := bufsPool.Get().(*bytes.Buffer)
	buf.WriteString(firstLine.Text())
	for {
		savedLen := buf.Len()
		p.setCheckpoint()
		if !ast.ValidNode(p.parsePlainNextLine(ind, ctx, buf)) {
			buf.Truncate(savedLen)
			p.rollback()
			break
		}
		p.commit()
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
}

// YAML specification: [134] s-ns-plain-next-line
func (p *parser) parsePlainNextLine(ind *indentation, ctx Context, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	savedLen := buf.Len()
	if !ast.ValidNode(p.parseFlowFolded(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	// checking that line has at least one plain safe string
	if !isPlainSafeToken(p.tok, ctx) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if !ast.ValidNode(p.parsePlainInLine(ctx, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
}

// YAML specification: [74] s-flow-folded
func (p *parser) parseFlowFolded(ind *indentation, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparateInLine()) {
		p.commit()
	} else {
		p.rollback()
	}

	savedLen := buf.Len()
	if !ast.ValidNode(p.parseFoldedLineBreak(ind, FlowInContext, buf)) ||
		!ast.ValidNode(p.parseFlowLinePrefix(ind)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
}

// YAML specification: [133] ns-plain-one-line
func (p *parser) parsePlainOneLine(ctx Context) ast.Node {
	start := p.tok.Start
	buf := bufsPool.Get().(*bytes.Buffer)
	if !ast.ValidNode(p.parsePlainFirst(ctx, buf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	if !ast.ValidNode(p.parsePlainInLine(ctx, buf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
}

// YAML specification: [132] nb-ns-plain-in-line
func (p *parser) parsePlainInLine(ctx Context, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start
	for {
		p.setCheckpoint()

		savedLen := buf.Len()

		for token.IsWhiteSpace(p.tok) {
			buf.WriteString(p.tok.Origin)
			p.next()
		}
		// YAML specification: [130] ns-plain-char
		// Only checking for plain safety conformity, because
		// violating cases for the plain char string are handled by scanner:
		// e.g.:
		// "#foo" transforms to tokens "#" (comment) and "foo"
		// "bar:" transforms to tokens "bar" and ":" (mapping value)
		if !isPlainSafeToken(p.tok, ctx) {
			buf.Truncate(savedLen)
			p.rollback()
			break
		}
		p.commit()
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
}

// YAML specification: [126] ns-plain-first
func (p *parser) parsePlainFirst(ctx Context, buf *bytes.Buffer) ast.Node {
	if p.tok.Type == token.StringType && isPlainSafeToken(p.tok, ctx) {
		// will be parsed as part of "plain in line"
		return ast.NewBasicNode(p.tok.Start, p.tok.End, ast.TextType)
	}
	switch p.tok.Type {
	case token.MappingKeyType, token.MappingValueType, token.SequenceEntryType:
		savedLen := buf.Len()
		buf.WriteString(p.tok.Origin)
		result := ast.NewBasicNode(p.tok.Start, p.tok.End, ast.TextType)
		p.next()
		// lookahead
		if isPlainSafeToken(p.tok, ctx) {
			buf.Truncate(savedLen)
			return ast.NewInvalidNode(result.Start(), result.End())
		}
		return result
	default:
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
}

func isPlainSafeToken(tok token.Token, ctx Context) bool {
	switch ctx {
	case FlowInContext, FlowOutContext:
		return tok.Type == token.StringType
	case BlockKeyContext, FlowKeyContext:
		return tok.Type == token.StringType && tok.ConformsCharSet(token.PlainSafeCharSetType)
	default:
		return false
	}
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
		} else {
			p.commit()
		}
	} else {
		p.rollback()
	}
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
		return content
	}
	p.rollback()

	content = p.parseCompactMapping(&mergedInd)
	if ast.ValidNode(content) {
		p.commit()
		return content
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
	castedHeader, ok := header.(*ast.BlockHeaderNode)
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
	castedContent, ok := content.(*ast.TextNode)
	if !ok {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, castedContent.Text())
}

// YAML specification: [182] l-folded-content
func (p *parser) parseFoldedContent(ind *indentation, chomping ast.ChompingType) ast.Node {
	start := p.tok.Start
	buf := bufsPool.Get().(*bytes.Buffer)

	p.setCheckpoint()
	if ast.ValidNode(p.parseDiffLines(ind, buf)) && ast.ValidNode(p.parseChompedLast(chomping, buf)) {
		p.commit()
	} else {
		buf.Reset()
		p.rollback()
	}
	if !ast.ValidNode(p.parseChompedEmpty(ind, chomping, buf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
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
		buf.WriteByte(token.LineFeedCharacter)
		if !ast.ValidNode(p.parseSameLines(ind, buf)) {
			buf.Truncate(savedLen)
			p.rollback()
			break
		}
		p.commit()
		savedLen = buf.Len()
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
		return ast.NewBasicNode(start, p.tok.End, ast.TextType)
	}

	if !ast.ValidNode(p.parseSpacedLines(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode(start, p.tok.End)
	}

	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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

	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
}

// YAML specification: [73] b-l-folded
func (p *parser) parseFoldedLineBreak(ind *indentation, ctx Context, buf *bytes.Buffer) ast.Node {
	start := p.tok.Start

	savedLen := buf.Len()
	p.setCheckpoint()

	if ast.ValidNode(p.parseTrimmed(ind, ctx, buf)) {
		p.commit()
		return ast.NewBasicNode(start, p.tok.End, ast.TextType)
	}

	p.rollback()
	buf.Truncate(savedLen)

	// linebreak as space
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	buf.WriteByte(byte(token.SpaceCharacter))
	p.next()

	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	castedHeader, ok := header.(*ast.BlockHeaderNode)
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
	castedContent, ok := content.(*ast.TextNode)
	if !ok {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	return ast.NewTextNode(start, p.tok.End, castedContent.Text())
}

// YAML specification: [173] l-literal-content
func (p *parser) parseLiteralContent(ind *indentation, chomping ast.ChompingType) ast.Node {
	start := p.tok.Start
	buf := bufsPool.Get().(*bytes.Buffer)

	p.setCheckpoint()
	if ast.ValidNode(p.parseLiteralText(ind, buf)) {
		for {
			p.setCheckpoint()
			savedLen := buf.Len()
			if !ast.ValidNode(p.parseLiteralNext(ind, buf)) {
				buf.Truncate(savedLen)
				p.rollback()
				break
			}
			p.commit()
		}
		if !ast.ValidNode(p.parseChompedLast(chomping, buf)) {
			buf.Reset()
			p.rollback()
		} else {
			p.commit()
		}
	} else {
		buf.Reset()
		p.rollback()
	}
	if !ast.ValidNode(p.parseChompedEmpty(ind, chomping, buf)) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(start, p.tok.End, text)
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
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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
	if p.tok.Type != token.StringType && !token.IsWhiteSpace(p.tok) {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	for p.tok.Type == token.StringType || token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
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

// YAML specification: [104] c-ns-alias-node
func (p *parser) parseAliasNode() ast.Node {
	start := p.tok.Start
	if p.tok.Type != token.AliasType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	if p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.AnchorCharSetType) {
		text := p.tok.Origin
		p.next()
		return ast.NewAliasNode(start, p.tok.End, text)
	}
	return ast.NewInvalidNode(start, p.tok.End)
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
		start := p.tok.Start
		node, ok := p.parseIndentWithLowerBound(ind.value).(*ast.IndentNode)
		if !ok || !ast.ValidNode(node) {
			return ast.NewInvalidNode(start, p.tok.End)
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

	// YAML specification: [91] c-secondary-tag-handle
	if p.tok.Type == token.TagType {
		p.next()
		return ast.NewBasicNode(start, p.tok.End, ast.TagType)
	}

	if p.tok.Type == token.StringType {
		// YAML specification: [92] c-named-tag-handle
		p.setCheckpoint()
		p.next()
		if p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.WordCharSetType) {
			p.next()
			if p.tok.Type == token.TagType {
				p.next()
				p.commit()
				return ast.NewBasicNode(start, p.tok.End, ast.TagType)
			}
		}
		p.rollback()
	}
	// else - primary
	// YAML specification: [90] c-primary-tag-handle
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
	if p.tok.Type != token.StringType || !isCorrectYAMLVersion(p.tok.Origin) {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	p.next()
	return ast.NewBasicNode(start, p.tok.End, ast.TextType)
}

func isCorrectYAMLVersion(s string) bool {
	const (
		start = iota
		metFirstPart
		metDot
		metSecondPart
	)
	currentState := start
	for _, c := range s {
		switch currentState {
		case start:
			if unicode.IsDigit(c) {
				currentState = metFirstPart
			} else {
				return false
			}
		case metFirstPart:
			if c == rune(token.DotCharacter) {
				currentState = metDot
			} else if !unicode.IsDigit(c) {
				return false
			}
		case metSecondPart:
			if !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return currentState == metSecondPart
}

// YAML specification: [202] l-document-prefix
func (p *parser) parseDocumentPrefix() ast.Node {
	start := p.tok.Start
	// document prefix without tokens (null prefix) is valid
	// so we have to break the endless loop of null prefixes
	var notNull bool

	if p.tok.Type == token.BOMType {
		notNull = true
		p.next()
	}
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseCommentLine()) {
			p.rollback()
			break
		}
		notNull = true
		p.commit()
	}
	if !notNull {
		return ast.NewNullNode(start)
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
	p.setCheckpoint()
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
	p.next()
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
	if p.tok.Type != token.LineBreakType && p.tok.Type != token.EOFType {
		return ast.NewInvalidNode(start, p.tok.End)
	}
	p.next()
	return ast.NewBasicNode(start, p.tok.End, ast.CommentType)
}

// YAML specification: [75] c-nb-comment-text
func (p *parser) parseCommentText() ast.Node {
	if p.tok.Type != token.CommentType {
		return ast.NewInvalidNode(p.tok.Start, p.tok.End)
	}
	start := p.tok.Start
	p.next()

	for p.tok.Type == token.StringType || token.IsWhiteSpace(p.tok) {
		p.next()
	}
	return ast.NewBasicNode(start, p.tok.Start, ast.CommentType)
}

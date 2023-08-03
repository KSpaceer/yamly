package parser

import (
	"bytes"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/token"
)

// YAML specification: [197] s-l+flow-in-block
func (p *parser) parseFlowInBlock(ind *indentation) ast.Node {
	localInd := indentation{
		value: ind.value + 1,
		mode:  StrictEquality,
	}
	if p.hasErrors() || !ast.ValidNode(p.parseSeparate(&localInd, FlowOutContext)) {
		return ast.NewInvalidNode()
	}
	node := p.parseFlowNode(&localInd, FlowOutContext)
	if !ast.ValidNode(node) {
		return ast.NewInvalidNode()
	}
	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode()
	}
	return node
}

func (p *parser) parseGenericFlowNode(
	ind *indentation,
	ctx Context,
	contentParse func(*indentation, Context) ast.Node,
) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	node := p.parseAliasNode()
	if ast.ValidNode(node) {
		p.commit()
		return node
	}

	p.rollback()
	p.setCheckpoint()

	node = contentParse(ind, ctx)
	if ast.ValidNode(node) {
		p.commit()
		return node
	}

	p.rollback()

	properties := p.parseProperties(ind, ctx)
	if !ast.ValidNode(properties) {
		return ast.NewInvalidNode()
	}

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparate(ind, ctx)) {
		node = contentParse(ind, ctx)
	}

	if ast.ValidNode(node) {
		p.commit()
		return ast.NewContentNode(properties, node)
	}

	p.rollback()
	return ast.NewContentNode(properties, ast.NewNullNode())
}

// YAML specification: [161] ns-flow-node
func (p *parser) parseFlowNode(ind *indentation, ctx Context) ast.Node {
	return p.parseGenericFlowNode(ind, ctx, p.parseFlowContent)
}

// YAML specification: [159] ns-flow-yaml-node
func (p *parser) parseFlowYAMLNode(ind *indentation, ctx Context) ast.Node {
	return p.parseGenericFlowNode(ind, ctx, p.parsePlain)
}

// YAML specification: [160] c-flow-json-node
func (p *parser) parseFlowJSONNode(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	properties := p.parseProperties(ind, ctx)
	if ast.ValidNode(properties) && ast.ValidNode(p.parseSeparate(ind, ctx)) {
		p.commit()
	} else {
		p.rollback()
	}
	content := p.parseFlowJSONContent(ind, ctx)
	if !ast.ValidNode(content) {
		return ast.NewInvalidNode()
	}
	return ast.NewContentNode(properties, content)
}

// YAML specification: [158] ns-flow-content
func (p *parser) parseFlowContent(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
		return ast.NewInvalidNode()
	}
}

// YAML specification: [109] c-double-quoted
func (p *parser) parseDoubleQuoted(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() || p.tok.Type != token.DoubleQuoteType {
		return ast.NewInvalidNode()
	}
	p.next()
	text := p.parseDoubleText(ind, ctx)
	if !ast.ValidNode(text) {
		return ast.NewInvalidNode()
	}

	if p.tok.Type != token.DoubleQuoteType {
		return ast.NewInvalidNode()
	}
	p.next()
	return text
}

// YAML specification: [121] nb-double-text
func (p *parser) parseDoubleText(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case FlowInContext, FlowOutContext:
		return p.parseDoubleMultiLine(ind)
	case BlockKeyContext, FlowKeyContext:
		return p.parseDoubleOneLine()
	default:
		return ast.NewInvalidNode()
	}
}

// YAML specification: [111] nb-double-one-line
func (p *parser) parseDoubleOneLine() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	buf := bufsPool.Get().(*bytes.Buffer)
	for (p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.DoubleQuotedCharSetType)) ||
		token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(text)
}

// YAML specification: [116] nb-double-multi-line
func (p *parser) parseDoubleMultiLine(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
	return ast.NewTextNode(text)
}

// YAML specification: [115] s-double-next-line
func (p *parser) parseDoubleNextLine(ind *indentation, buf *bytes.Buffer) ast.Node {
	savedLen := buf.Len()

	if p.hasErrors() || !ast.ValidNode(p.parseDoubleBreak(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
	}

	p.setCheckpoint()

	if p.tok.Type != token.StringType || !p.tok.ConformsCharSet(token.DoubleQuotedCharSetType) {
		p.rollback()
		return ast.NewBasicNode(ast.TextType)
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
		for token.IsWhiteSpace(p.tok) {
			buf.WriteString(p.tok.Origin)
			p.next()
		}
	} else {
		p.commit()
	}

	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [113] s-double-break
func (p *parser) parseDoubleBreak(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	savedLen := buf.Len()
	if ast.ValidNode(p.parseDoubleEscaped(ind, buf)) {
		p.commit()
		return ast.NewBasicNode(ast.TextType)
	}
	p.rollback()
	buf.Truncate(savedLen)
	return p.parseFlowFolded(ind, buf)
}

// YAML specification: [112] s-double-escaped
func (p *parser) parseDoubleEscaped(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	savedLen := buf.Len()

	for token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}

	if p.tok.Type != token.StringType || p.tok.Origin != "\\" {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
	}
	p.next()
	if p.tok.Type != token.LineBreakType {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
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
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [140] c-single-quoted
func (p *parser) parseSingleQuoted(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() || p.tok.Type != token.SingleQuoteType {
		return ast.NewInvalidNode()
	}
	p.next()
	text := p.parseSingleText(ind, ctx)
	if !ast.ValidNode(text) {
		return ast.NewInvalidNode()
	}

	if p.tok.Type != token.SingleQuoteType {
		return ast.NewInvalidNode()
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
		return ast.NewInvalidNode()
	}
}

// YAML specification: [122] nb-single-one-line
func (p *parser) parseSingleOneLine() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	buf := bufsPool.Get().(*bytes.Buffer)
	for (p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.SingleQuotedCharSetType)) ||
		token.IsWhiteSpace(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(text)
}

// YAML specification: [125] nb-single-multi-line
func (p *parser) parseSingleMultiLine(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
	return ast.NewTextNode(text)
}

// YAML specification: [124] s-single-next-line
func (p *parser) parseSingleNextLine(ind *indentation, buf *bytes.Buffer) ast.Node {
	savedLen := buf.Len()
	if p.hasErrors() || !ast.ValidNode(p.parseFlowFolded(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()

	if p.tok.Type != token.StringType || !p.tok.ConformsCharSet(token.SingleQuotedCharSetType) {
		p.rollback()
		return ast.NewBasicNode(ast.TextType)
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
		for token.IsWhiteSpace(p.tok) {
			buf.WriteString(p.tok.Origin)
			p.next()
		}
	} else {
		p.commit()
	}
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [140] c-flow-mapping
func (p *parser) parseFlowMapping(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() || p.tok.Type != token.MappingStartType {
		return ast.NewInvalidNode()
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
		content = ast.NewMappingNode(nil)
	}

	if p.tok.Type != token.MappingEndType {
		return ast.NewInvalidNode()
	}
	p.next()
	return content
}

// YAML specification: [141] ns-s-flow-map-entries
func (p *parser) parseFlowMappingEntries(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	entry := p.parseFlowMappingEntry(ind, ctx)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode()
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
		} else {
			entries = append(entries, entry)
			p.commit()
		}
	}

	return ast.NewMappingNode(entries)
}

// YAML specification: [142] ns-flow-map-entry
func (p *parser) parseFlowMappingEntry(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	if p.tok.Type == token.MappingKeyType {
		p.setCheckpoint()
		p.next()
		if ast.ValidNode(p.parseSeparate(ind, ctx)) {
			entry := p.parseFlowMappingExplicitEntry(ind, ctx)
			if ast.ValidNode(entry) {
				p.commit()
				return entry
			}
		}
		p.rollback()
	}
	entry := p.parseFlowMappingImplicitEntry(ind, ctx)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode()
	}
	return entry
}

// YAML specification: [137] c-flow-sequence
func (p *parser) parseFlowSequence(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() || p.tok.Type != token.SequenceStartType {
		return ast.NewInvalidNode()
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
		content = ast.NewSequenceNode(nil)
	}

	if p.tok.Type != token.SequenceEndType {
		return ast.NewInvalidNode()
	}
	p.next()
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
		return ast.NewInvalidNode()
	}
	return p.parseFlowSequenceEntries(ind, ctx)
}

// YAML specification: [138] ns-s-flow-seq-entries
func (p *parser) parseFlowSequenceEntries(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	entry := p.parseFlowSequenceEntry(ind, ctx)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode()
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

	return ast.NewSequenceNode(entries)
}

// YAML specification: [139] ns-flow-seq-entry
func (p *parser) parseFlowSequenceEntry(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	pair := p.parseFlowPairEntry(ind, ctx)
	if ast.ValidNode(pair) {
		p.commit()
		return pair
	}
	p.rollback()

	if p.tok.Type != token.MappingKeyType {
		return ast.NewInvalidNode()
	}

	p.next()
	if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
		return ast.NewInvalidNode()
	}

	pair = p.parseFlowMappingExplicitEntry(ind, ctx)
	if !ast.ValidNode(pair) {
		return ast.NewInvalidNode()
	}
	return pair
}

// YAML specification: [151] ns-flow-pair-entry
func (p *parser) parseFlowPairEntry(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	key := p.parseImplicitJSONKey(FlowKeyContext)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode()
	}
	value := p.parseFlowMappingAdjacentValue(ind, ctx)
	if !ast.ValidNode(value) {
		return ast.NewInvalidNode()
	}
	return ast.NewMappingEntryNode(key, value)
}

// YAML specification: [155] c-s-implicit-json-key
func (p *parser) parseImplicitJSONKey(ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	localInd := indentation{
		value: 0,
		mode:  StrictEquality,
	}
	node := p.parseFlowJSONNode(&localInd, ctx)
	if !ast.ValidNode(node) {
		return ast.NewInvalidNode()
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
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	key := p.parseImplicitYAMLKey(FlowKeyContext)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode()
	}
	value := p.parseFlowMappingSeparateValue(ind, ctx)
	if !ast.ValidNode(value) {
		return ast.NewInvalidNode()
	}
	return ast.NewMappingEntryNode(key, value)
}

// YAML specification: [154] ns-s-implicit-yaml-key
func (p *parser) parseImplicitYAMLKey(ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	localInd := indentation{
		value: 0,
		mode:  StrictEquality,
	}
	node := p.parseFlowYAMLNode(&localInd, ctx)
	if !ast.ValidNode(node) {
		return ast.NewInvalidNode()
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
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	entry := p.parseFlowMappingImplicitEntry(ind, ctx)
	if ast.ValidNode(entry) {
		p.commit()
		return entry
	}
	p.rollback()
	return ast.NewMappingEntryNode(ast.NewNullNode(), ast.NewNullNode())
}

// YAML specification: [144] ns-flow-map-implicit-entry
func (p *parser) parseFlowMappingImplicitEntry(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	value := p.parseFlowMappingSeparateValue(ind, ctx)
	if !ast.ValidNode(value) {
		return ast.NewInvalidNode()
	}
	return ast.NewMappingEntryNode(ast.NewNullNode(), value)
}

// YAML specification: [148] c-ns-flow-map-json-key-entry
func (p *parser) parseFlowMappingJSONKeyEntry(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	key := p.parseFlowJSONNode(ind, ctx)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode()
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
		value = ast.NewNullNode()
	} else {
		p.commit()
	}

	return ast.NewMappingEntryNode(key, value)
}

// YAML specification: [149] c-ns-flow-map-adjacent-value
func (p *parser) parseFlowMappingAdjacentValue(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() || p.tok.Type != token.MappingValueType {
		return ast.NewInvalidNode()
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
		value = ast.NewNullNode()
	}
	return value
}

// YAML specification: [145] ns-flow-map-yaml-key-entry
func (p *parser) parseFlowMappingYAMLKeyEntry(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	key := p.parseFlowYAMLNode(ind, ctx)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode()
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
		value = ast.NewNullNode()
	} else {
		p.commit()
	}
	return ast.NewMappingEntryNode(key, value)
}

// YAML specification: [147] c-ns-flow-map-separate-value
func (p *parser) parseFlowMappingSeparateValue(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() || p.tok.Type != token.MappingValueType {
		return ast.NewInvalidNode()
	}
	p.next()
	// lookahead
	if isPlainSafeToken(p.tok, ctx) {
		return ast.NewInvalidNode()
	}

	p.setCheckpoint()
	if ast.ValidNode(p.parseSeparate(ind, ctx)) {
		value := p.parseFlowNode(ind, ctx)
		if ast.ValidNode(value) {
			p.commit()
			return value
		}
	}
	p.rollback()

	return ast.NewNullNode()
}

// YAML specification: [131] ns-plain
func (p *parser) parsePlain(ind *indentation, ctx Context) ast.Node {
	switch ctx {
	case FlowInContext, FlowOutContext:
		return p.parsePlainMultiLine(ind, ctx)
	case BlockKeyContext, FlowKeyContext:
		return p.parsePlainOneLine(ctx)
	default:
		return ast.NewInvalidNode()
	}
}

// YAML specification: [135] ns-plain-multi-line
func (p *parser) parsePlainMultiLine(ind *indentation, ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	firstLine, ok := p.parsePlainOneLine(ctx).(*ast.TextNode)
	if !ok || !ast.ValidNode(firstLine) {
		return ast.NewInvalidNode()
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
	return ast.NewTextNode(text)
}

// YAML specification: [134] s-ns-plain-next-line
func (p *parser) parsePlainNextLine(ind *indentation, ctx Context, buf *bytes.Buffer) ast.Node {
	savedLen := buf.Len()
	if p.hasErrors() || !ast.ValidNode(p.parseFlowFolded(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
	}
	// checking that line has at least one plain safe string
	if !isPlainSafeToken(p.tok, ctx) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
	}
	if !ast.ValidNode(p.parsePlainInLine(ctx, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [74] s-flow-folded
func (p *parser) parseFlowFolded(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [133] ns-plain-one-line
func (p *parser) parsePlainOneLine(ctx Context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	buf := bufsPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufsPool.Put(buf)
	}()
	if !ast.ValidNode(p.parsePlainFirst(ctx, buf)) {
		if p.deadEndFinder.Mark(p.tok) {
			p.appendError(DeadEndError{Pos: p.tok.Start})
		}
		return ast.NewInvalidNode()
	}
	if !ast.ValidNode(p.parsePlainInLine(ctx, buf)) {
		return ast.NewInvalidNode()
	}
	text := buf.String()
	return ast.NewTextNode(text)
}

// YAML specification: [132] nb-ns-plain-in-line
func (p *parser) parsePlainInLine(ctx Context, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [126] ns-plain-first
func (p *parser) parsePlainFirst(ctx Context, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	if p.tok.Type == token.StringType && isPlainSafeToken(p.tok, ctx) {
		// will be parsed as part of "plain in line"
		return ast.NewBasicNode(ast.TextType)
	}
	switch p.tok.Type {
	case token.MappingKeyType, token.MappingValueType, token.SequenceEntryType:
		savedLen := buf.Len()
		p.setCheckpoint()
		buf.WriteString(p.tok.Origin)
		result := ast.NewBasicNode(ast.TextType)
		p.next()
		// lookahead
		if !isPlainSafeToken(p.tok, ctx) {
			p.rollback()
			buf.Truncate(savedLen)
			return ast.NewInvalidNode()
		}
		p.commit()
		return result
	default:
		return ast.NewInvalidNode()
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

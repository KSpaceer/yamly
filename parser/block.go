package parser

import (
	"bytes"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/token"
)

// YAML specification: [196] s-l+block-node
func (p *parser) parseBlockNode(ind *indentation, ctx context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
	return ast.NewInvalidNode()
}

// YAML specification: [198] s-l+block-in-block
func (p *parser) parseBlockInBlock(ind *indentation, ctx context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
	return ast.NewInvalidNode()
}

// YAML specification: [200] s-l+block-collection
func (p *parser) parseBlockCollection(ind *indentation, ctx context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
		return ast.NewInvalidNode()
	}

	p.setCheckpoint()
	collection := p.parseSeqSpace(ind, ctx)
	if ast.ValidNode(collection) {
		p.commit()
		return newContentNode(properties, collection)
	}
	p.rollback()
	collection = p.parseBlockMapping(ind)
	if !ast.ValidNode(collection) {
		return ast.NewInvalidNode()
	}
	return newContentNode(properties, collection)
}

// YAML specification: [187] l+block-mapping
func (p *parser) parseBlockMapping(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	localInd := indentation{
		value: ind.value + 1,
		mode:  withLowerBoundIndentationMode,
	}
	if !ast.ValidNode(p.parseIndent(&localInd)) {
		return ast.NewInvalidNode()
	}
	entry := p.parseBlockMappingEntry(&localInd)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode()
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

	return ast.NewMappingNode(entries)
}

func (p *parser) parseCompactMapping(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	entry := p.parseBlockMappingEntry(ind)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode()
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

	return ast.NewMappingNode(entries)
}

// YAML specification: [188] ns-l-block-map-entry
func (p *parser) parseBlockMappingEntry(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}

	switch p.tok.Type {
	case token.MappingKeyType:
		return p.parseBlockMappingExplicitEntry(ind)
	default:
		return p.parseBlockMappingImplicitEntry(ind)
	}
}

// YAML specification: [192] ns-l-block-map-implicit-entry
func (p *parser) parseBlockMappingImplicitEntry(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}

	p.setCheckpoint()
	key := p.parseBlockMappingImplicitKey()
	if !ast.ValidNode(key) {
		p.rollback()
		key = ast.NewNullNode()
	} else {
		p.commit()
	}
	value := p.parseBlockMappingImplicitValue(ind)
	if !ast.ValidNode(value) {
		return ast.NewInvalidNode()
	}
	return ast.NewMappingEntryNode(key, value)
}

// YAML specification: [193] ns-s-block-map-implicit-key
func (p *parser) parseBlockMappingImplicitKey() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	key := p.parseImplicitJSONKey(blockKeyContext)
	if ast.ValidNode(key) {
		p.commit()
		return key
	}
	p.rollback()
	return p.parseImplicitYAMLKey(blockKeyContext)
}

// YAML specification: [194] c-l-block-map-implicit-value
func (p *parser) parseBlockMappingImplicitValue(ind *indentation) ast.Node {
	if p.hasErrors() || p.tok.Type != token.MappingValueType {
		return ast.NewInvalidNode()
	}
	p.next()

	p.setCheckpoint()
	value := p.parseBlockNode(ind, blockOutContext)
	if ast.ValidNode(value) {
		p.commit()
		return value
	}
	p.rollback()
	value = ast.NewNullNode()
	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode()
	}
	return value
}

// YAML specification: [189] c-l-block-map-explicit-entry
func (p *parser) parseBlockMappingExplicitEntry(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	key := p.parseBlockMappingExplicitKey(ind)
	if !ast.ValidNode(key) {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	value := p.parseBlockMappingExplicitValue(ind)
	if !ast.ValidNode(value) {
		p.rollback()
		value = ast.NewNullNode()
	} else {
		p.commit()
	}

	return ast.NewMappingEntryNode(key, value)
}

// YAML specification: [189] c-l-block-map-explicit-key
func (p *parser) parseBlockMappingExplicitKey(ind *indentation) ast.Node {
	if p.hasErrors() || p.tok.Type != token.MappingKeyType {
		return ast.NewInvalidNode()
	}
	p.next()

	return p.parseBlockIndented(ind, blockOutContext)
}

// YAML specification: [189] l-block-map-explicit-value
func (p *parser) parseBlockMappingExplicitValue(ind *indentation) ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseIndent(ind)) {
		return ast.NewInvalidNode()
	}
	if p.tok.Type != token.MappingValueType {
		return ast.NewInvalidNode()
	}
	p.next()
	return p.parseBlockIndented(ind, blockOutContext)
}

// YAML specification: [201] seq-space
func (p *parser) parseSeqSpace(ind *indentation, ctx context) ast.Node {
	switch ctx {
	case blockInContext:
		return p.parseBlockSequence(ind)
	case blockOutContext:
		ind.value--
		node := p.parseBlockSequence(ind)
		ind.value++
		return node
	default:
		return ast.NewInvalidNode()
	}
}

// YAML specification: [183] l+block-sequence
func (p *parser) parseBlockSequence(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}

	localInd := indentation{
		value: ind.value + 1,
		mode:  withLowerBoundIndentationMode,
	}
	if !ast.ValidNode(p.parseIndent(&localInd)) {
		return ast.NewInvalidNode()
	}
	entry := p.parseBlockSequenceEntry(&localInd)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode()
	}
	entries := []ast.Node{entry}

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

	return ast.NewSequenceNode(entries)
}

// YAML specification: [184] c-l-block-seq-entry
func (p *parser) parseBlockSequenceEntry(ind *indentation) ast.Node {
	if p.hasErrors() || p.tok.Type != token.SequenceEntryType {
		return ast.NewInvalidNode()
	}
	p.next()
	switch p.tok.Type {
	case token.SpaceType, token.TabType, token.LineBreakType:
		return p.parseBlockIndented(ind, blockInContext)
	default:
		return ast.NewInvalidNode()
	}
}

// YAML specification: [185] s-l+block-indented
func (p *parser) parseBlockIndented(ind *indentation, ctx context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	localInd := indentation{
		value: 0,
		mode:  withLowerBoundIndentationMode,
	}
	// because we have "opportunistic" indentation starting with 0,
	// parseIndent will never return invalid node
	p.parseIndent(&localInd)
	p.setCheckpoint()
	mergedInd := indentation{
		value: ind.value + 1 + localInd.value,
		mode:  strictEqualityIndentationMode,
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

	if ast.ValidNode(p.parseComments()) {
		return ast.NewNullNode()
	}
	return ast.NewInvalidNode()
}

// YAML specification: [186] ns-l-compact-sequence
func (p *parser) parseCompactSequence(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	entry := p.parseBlockSequenceEntry(ind)
	if !ast.ValidNode(entry) {
		return ast.NewInvalidNode()
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

	return ast.NewSequenceNode(entries)
}

// YAML specification: [199] s-l+block-scalar
func (p *parser) parseBlockScalar(ind *indentation, ctx context) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	ind.value++
	if !ast.ValidNode(p.parseSeparate(ind, ctx)) {
		ind.value--
		return ast.NewInvalidNode()
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
		return ast.NewInvalidNode()
	}

	if !ast.ValidNode(content) {
		return ast.NewInvalidNode()
	}
	return newContentNode(properties, content)
}

// YAML specification: [182] c-l+folded
func (p *parser) parseFolded(ind *indentation) ast.Node {
	if p.hasErrors() || p.tok.Type != token.FoldedType {
		return ast.NewInvalidNode()
	}
	p.next()
	header := p.parseBlockHeader()
	if !ast.ValidNode(header) {
		return ast.NewInvalidNode()
	}
	castedHeader, ok := header.(*ast.BlockHeaderNode)
	if !ok {
		ast.NewInvalidNode()
	}
	foldedInd := indentation{
		value: ind.value,
		mode:  withLowerBoundIndentationMode,
	}
	if indVal := castedHeader.IndentationIndicator(); indVal != 0 {
		foldedInd.value = ind.value + indVal
		foldedInd.mode = strictEqualityIndentationMode
	}
	content := p.parseFoldedContent(&foldedInd, castedHeader.ChompingIndicator())
	return content
}

// YAML specification: [182] l-folded-content
func (p *parser) parseFoldedContent(ind *indentation, chomping ast.ChompingType) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	buf := bufsPool.Get().(*bytes.Buffer)

	p.setCheckpoint()
	if ast.ValidNode(p.parseDiffLines(ind, buf)) && ast.ValidNode(p.parseChompedLast(chomping, buf)) {
		p.commit()
	} else {
		buf.Reset()
		p.rollback()
	}
	if !ast.ValidNode(p.parseChompedEmpty(ind, chomping, buf)) {
		return ast.NewInvalidNode()
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(text)
}

// YAML specification: [181] l-nb-diff-lines
func (p *parser) parseDiffLines(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseSameLines(ind, buf)) {
		return ast.NewInvalidNode()
	}
	savedLen := buf.Len()
	for {
		p.setCheckpoint()
		if p.tok.Type != token.LineBreakType {
			buf.Truncate(savedLen)
			p.rollback()
			break
		}
		buf.WriteString(p.tok.Origin)
		p.next()
		if !ast.ValidNode(p.parseSameLines(ind, buf)) {
			buf.Truncate(savedLen)
			p.rollback()
			break
		}
		p.commit()
		savedLen = buf.Len()
	}
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [180] l-nb-same-lines
func (p *parser) parseSameLines(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	savedLen := buf.Len()
	localLen := savedLen
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, blockInContext, buf)) {
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
		return ast.NewBasicNode(ast.TextType)
	}

	if !ast.ValidNode(p.parseSpacedLines(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
	}

	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [179] l-nb-spaced-lines
func (p *parser) parseSpacedLines(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseSpacedText(ind, buf)) {
		return ast.NewInvalidNode()
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
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [177] s-nb-spaced-text
func (p *parser) parseSpacedText(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	if !ast.ValidNode(p.parseIndent(ind)) {
		return ast.NewInvalidNode()
	}
	if !token.IsWhiteSpace(p.tok) {
		return ast.NewInvalidNode()
	}
	buf.WriteString(p.tok.Origin)
	p.next()
	p.tokSrc.SetRawMode()
	for token.IsNonBreak(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	p.tokSrc.UnsetRawMode()
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [178] b-l-spaced
func (p *parser) parseSpacedLineBreak(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() || p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode()
	}
	buf.WriteString(p.tok.Origin)

	savedLen := buf.Len()

	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, blockInContext, buf)) {
			p.rollback()
			buf.Truncate(savedLen)
			break
		}
		p.commit()
		savedLen = buf.Len()
	}

	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [176] l-nb-folded-lines
func (p *parser) parseFoldedLines(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseFoldedText(ind, buf)) {
		return ast.NewInvalidNode()
	}

	savedLen := buf.Len()
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseFoldedLineBreak(ind, blockInContext, buf)) {
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
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [175] s-nb-folded-text
func (p *parser) parseFoldedText(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseIndent(ind)) {
		return ast.NewInvalidNode()
	}
	if !token.IsNonBreak(p.tok) || token.IsWhiteSpace(p.tok) {
		return ast.NewInvalidNode()
	}
	buf.WriteString(p.tok.Origin)
	p.next()
	p.tokSrc.SetRawMode()
	for token.IsNonBreak(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	p.tokSrc.UnsetRawMode()
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [73] b-l-folded
func (p *parser) parseFoldedLineBreak(ind *indentation, ctx context, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	savedLen := buf.Len()
	p.setCheckpoint()

	if ast.ValidNode(p.parseTrimmed(ind, ctx, buf)) {
		p.commit()
		return ast.NewBasicNode(ast.TextType)
	}

	p.rollback()
	buf.Truncate(savedLen)

	// linebreak as space
	if p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode()
	}
	buf.WriteRune(' ')
	p.next()

	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [71] b-l-trimmed
func (p *parser) parseTrimmed(ind *indentation, ctx context, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() || p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode()
	}
	p.next()

	savedLen := buf.Len()

	if !ast.ValidNode(p.parseEmpty(ind, ctx, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
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
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [170] c-l+literal
func (p *parser) parseLiteral(ind *indentation) ast.Node {
	if p.hasErrors() || p.tok.Type != token.LiteralType {
		return ast.NewInvalidNode()
	}
	p.next()
	header := p.parseBlockHeader()
	if !ast.ValidNode(header) {
		return ast.NewInvalidNode()
	}
	castedHeader, ok := header.(*ast.BlockHeaderNode)
	if !ok {
		return ast.NewInvalidNode()
	}
	literalInd := indentation{
		value: ind.value,
		mode:  withLowerBoundIndentationMode,
	}
	if indVal := castedHeader.IndentationIndicator(); indVal != 0 {
		literalInd.value = ind.value + indVal
		literalInd.mode = strictEqualityIndentationMode
	}
	content := p.parseLiteralContent(&literalInd, castedHeader.ChompingIndicator())
	return content
}

// YAML specification: [173] l-literal-content
func (p *parser) parseLiteralContent(ind *indentation, chomping ast.ChompingType) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
		return ast.NewInvalidNode()
	}
	text := buf.String()
	buf.Reset()
	bufsPool.Put(buf)
	return ast.NewTextNode(text)
}

// YAML specification: [165] b-chomped-last
func (p *parser) parseChompedLast(chomping ast.ChompingType, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	switch chomping {
	case ast.StripChompingType:
		switch p.tok.Type {
		case token.LineBreakType, token.EOFType:
		default:
			return ast.NewInvalidNode()
		}
	case ast.ClipChompingType, ast.KeepChompingType:
		switch p.tok.Type {
		case token.LineBreakType:
			buf.WriteString(p.tok.Origin)
		case token.EOFType:
		default:
			return ast.NewInvalidNode()
		}
	}
	p.next()
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [166] l-chomped-empty
func (p *parser) parseChompedEmpty(ind *indentation, chomping ast.ChompingType, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	switch chomping {
	case ast.ClipChompingType, ast.StripChompingType:
		return p.parseStripEmpty(ind)
	case ast.KeepChompingType:
		return p.parseKeepEmpty(ind, buf)
	default:
		return ast.NewInvalidNode()
	}
}

// YAML specification: [168] l-keep-empty
func (p *parser) parseKeepEmpty(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	savedLen := buf.Len()
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, blockInContext, buf)) {
			p.rollback()
			buf.Truncate(savedLen)
			break
		}
		p.commit()
		savedLen = buf.Len()
	}
	p.setCheckpoint()
	if !ast.ValidNode(p.parseTrailComments(ind)) {
		buf.Truncate(savedLen)
		p.rollback()
	} else {
		p.commit()
	}
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [167] l-strip-empty
func (p *parser) parseStripEmpty(ind *indentation) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
		p.next()
		p.commit()
	}

	p.setCheckpoint()
	if !ast.ValidNode(p.parseTrailComments(ind)) {
		p.rollback()
	} else {
		p.commit()
	}
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [169] l-trail-comments
func (p *parser) parseTrailComments(ind *indentation) ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseIndentLessThan(ind.value)) {
		return ast.NewInvalidNode()
	}
	if !ast.ValidNode(p.parseCommentText()) {
		return ast.NewInvalidNode()
	}
	if p.tok.Type != token.LineBreakType && p.tok.Type != token.EOFType {
		return ast.NewInvalidNode()
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
	return ast.NewBasicNode(ast.CommentType)
}

// YAML specification: [172] l-nb-literal-next
func (p *parser) parseLiteralNext(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() || p.tok.Type != token.LineBreakType {
		return ast.NewInvalidNode()
	}
	savedLen := buf.Len()
	buf.WriteString(p.tok.Origin)
	p.next()
	if !ast.ValidNode(p.parseLiteralText(ind, buf)) {
		buf.Truncate(savedLen)
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.TextType)
}

// YAML specification: [171] l-nb-literal-text
func (p *parser) parseLiteralText(ind *indentation, buf *bytes.Buffer) ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseEmpty(ind, blockInContext, buf)) {
			p.rollback()
			break
		}
		p.commit()
	}
	if !ast.ValidNode(p.parseIndent(ind)) {
		return ast.NewInvalidNode()
	}
	if !token.IsNonBreak(p.tok) {
		return ast.NewInvalidNode()
	}
	p.tokSrc.SetRawMode()
	for token.IsNonBreak(p.tok) {
		buf.WriteString(p.tok.Origin)
		p.next()
	}
	p.tokSrc.UnsetRawMode()
	return ast.NewBasicNode(ast.TextType)
}

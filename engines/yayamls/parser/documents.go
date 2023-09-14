package parser

import (
	"github.com/KSpaceer/yamly/engines/yayamls/ast"
	"github.com/KSpaceer/yamly/engines/yayamls/chars"
	"github.com/KSpaceer/yamly/engines/yayamls/token"
	"unicode"
)

// YAML specification: [211] l-yaml-stream
func (p *parser) parseStream() ast.Node {
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
		if ast.ValidNode(p.parseSuffixesAndPrefixes()) {
			p.commit()
			p.setCheckpoint()
			doc = p.parseAnyDocument()
			if !ast.ValidNode(doc) {
				p.rollback()
			} else {
				docs = append(docs, doc)
				p.commit()
			}
			continue
		}
		p.rollback()

		if p.tok.Type == token.BOMType {
			p.next()
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

	return ast.NewStreamNode(docs)
}

func (p *parser) parseSuffixesAndPrefixes() ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseDocumentSuffix()) {
		return ast.NewInvalidNode()
	}
	for {
		p.setCheckpoint()
		if !ast.ValidNode(p.parseDocumentSuffix()) {
			p.rollback()
			break
		}
		p.commit()
	}

	for {
		p.setCheckpoint()
		if prefix := p.parseDocumentPrefix(); !ast.ValidNode(prefix) || prefix.Type() == ast.NullType {
			p.rollback()
			break
		}
		p.commit()
	}
	return ast.NewBasicNode(ast.DocumentPrefixType)
}

// YAML specification: [210] l-any-document
func (p *parser) parseAnyDocument() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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

	return ast.NewInvalidNode()
}

// YAML specification: [209] l-directive-document
func (p *parser) parseDirectiveDocument() ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseDirective()) {
		return ast.NewInvalidNode()
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
	if p.hasErrors() || p.tok.Type != token.DirectiveEndType {
		return ast.NewInvalidNode()
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
		return ast.NewInvalidNode()
	}
	p.commit()
	return ast.NewNullNode()
}

// YAML specification: [207] l-bare-document
func (p *parser) parseBareDocument() ast.Node {
	return p.parseBlockNode(&indentation{value: -1, mode: strictEqualityIndentationMode}, blockInContext)
}

// YAML specification: [82] l-directive
func (p *parser) parseDirective() ast.Node {
	if p.hasErrors() || p.tok.Type != token.DirectiveType {
		return ast.NewInvalidNode()
	}
	// directive may have any name
	p.tokSrc.SetRawMode()
	p.next()
	p.tokSrc.UnsetRawMode()
	var directiveNode ast.Node
	switch p.tok.Origin {
	case chars.YAMLDirective:
		p.next()
		directiveNode = p.parseYAMLDirective()
	case chars.TagDirective:
		p.next()
		directiveNode = p.parseTagDirective()
	default:
		directiveNode = p.parseReservedDirective()
	}
	if !ast.ValidNode(directiveNode) {
		return ast.NewInvalidNode()
	}

	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.DirectiveType)
}

// YAML specification: [83] ns-reserved-directive
func (p *parser) parseReservedDirective() ast.Node {
	if p.hasErrors() || p.tok.Type != token.StringType {
		return ast.NewInvalidNode()
	}
	// reserved directive may have any parameters
	p.tokSrc.SetRawMode()
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
	p.tokSrc.UnsetRawMode()

	return ast.NewBasicNode(ast.DirectiveType)
}

// YAML specification: [88] ns-tag-directive
func (p *parser) parseTagDirective() ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode()
	}
	if !ast.ValidNode(p.parseTagHandle()) {
		return ast.NewInvalidNode()
	}
	// tag prefix may have almost all possible characters
	p.tokSrc.SetRawMode()
	defer p.tokSrc.UnsetRawMode()
	if !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode()
	}
	if !ast.ValidNode(p.parseTagPrefix()) {
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.DirectiveType)
}

// YAML specification: [89] c-tag-handle
func (p *parser) parseTagHandle() ast.Node {
	if p.hasErrors() || p.tok.Type != token.TagType {
		return ast.NewInvalidNode()
	}
	p.next()

	// YAML specification: [91] c-secondary-tag-handle
	if p.tok.Type == token.TagType {
		p.next()
		return ast.NewBasicNode(ast.TagType)
	}

	// YAML specification: [92] c-named-tag-handle
	p.setCheckpoint()
	if p.tok.Type == token.StringType && p.tok.ConformsCharSet(chars.WordCharSetType) {
		p.next()
		if p.tok.Type == token.TagType {
			p.next()
			p.commit()
			return ast.NewBasicNode(ast.TagType)
		}
	}
	p.rollback()

	// else - primary
	// YAML specification: [90] c-primary-tag-handle
	return ast.NewBasicNode(ast.TagType)
}

// YAML specification: [93] ns-tag-prefix
func (p *parser) parseTagPrefix() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	if ast.ValidNode(p.parseLocalTagPrefix()) {
		p.commit()
	} else {
		p.rollback()
		// trying global tag
		// YAML specification: [95] ns-global-tag-prefix
		if p.tok.Type != token.StringType || len(p.tok.Origin) == 0 {
			return ast.NewInvalidNode()
		}
		if !chars.ConformsCharSet(p.tok.Origin[:1], chars.TagCharSetType) ||
			!p.tok.ConformsCharSet(chars.URICharSetType) {
			return ast.NewInvalidNode()
		}
		p.next()
	}

	return ast.NewBasicNode(ast.TagType)
}

// YAML specification: [94] c-ns-local-tag-prefix
func (p *parser) parseLocalTagPrefix() ast.Node {
	if p.hasErrors() || p.tok.Type != token.TagType {
		return ast.NewInvalidNode()
	}
	p.next()
	if p.tok.Type == token.StringType && p.tok.ConformsCharSet(chars.URICharSetType) {
		p.next()
	}
	return ast.NewBasicNode(ast.TagType)
}

// YAML specification: [86] ns-yaml-directive
func (p *parser) parseYAMLDirective() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	for token.IsWhiteSpace(p.tok) {
		p.next()
	}
	if !ast.ValidNode(p.parseYAMLVersion()) {
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.DirectiveType)
}

// YAML specification: [87] ns-yaml-version
func (p *parser) parseYAMLVersion() ast.Node {
	if p.hasErrors() || p.tok.Type != token.StringType || !isCorrectYAMLVersion(p.tok.Origin) {
		return ast.NewInvalidNode()
	}
	p.next()
	return ast.NewBasicNode(ast.TextType)
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
			if c == '.' {
				currentState = metDot
			} else if !unicode.IsDigit(c) {
				return false
			}
		case metDot, metSecondPart:
			if !unicode.IsDigit(c) {
				return false
			}
			currentState = metSecondPart
		}
	}
	return currentState == metSecondPart
}

// YAML specification: [202] l-document-prefix
func (p *parser) parseDocumentPrefix() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
		return ast.NewNullNode()
	}
	return ast.NewBasicNode(ast.DocumentPrefixType)
}

// YAML specification: [205] l-document-suffix
func (p *parser) parseDocumentSuffix() ast.Node {
	if p.hasErrors() || p.tok.Type != token.DocumentEndType {
		return ast.NewInvalidNode()
	}
	p.next()
	if !ast.ValidNode(p.parseComments()) {
		return ast.NewInvalidNode()
	}
	return ast.NewBasicNode(ast.DocumentSuffixType)
}

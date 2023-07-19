package parser

import (
	"fmt"
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/token"
	"strconv"
	"strings"
)

// YAML specification: [162] c-b-block-header
func (p *parser) parseBlockHeader() ast.Node {
	chompingIndicator := p.parseChompingIndicator()
	indentationIndicator, err := p.parseIndentationIndicator()
	if err != nil {
		return ast.NewInvalidNode()
	}
	if chompingIndicator == ast.ClipChompingType {
		chompingIndicator = p.parseChompingIndicator()
	}
	p.setCheckpoint()
	if !ast.ValidNode(p.parseSeparatedComment()) {
		p.rollback()
		return ast.NewInvalidNode()
	}
	p.commit()
	return ast.NewBlockHeaderNode(chompingIndicator, indentationIndicator)
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
	var tag, anchor ast.Node
	switch p.tok.Type {
	case token.TagType:
		tag = p.parseTagProperty()
	case token.AnchorType:
		anchor = p.parseAnchorProperty()
	default:
		return ast.NewInvalidNode()
	}

	if !ast.ValidNode(tag) && !ast.ValidNode(anchor) {
		return ast.NewInvalidNode()
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

	return ast.NewPropertiesNode(tag, anchor)
}

// YAML specification: [104] c-ns-alias-node
func (p *parser) parseAliasNode() ast.Node {
	if p.tok.Type != token.AliasType {
		return ast.NewInvalidNode()
	}
	p.next()
	if p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.AnchorCharSetType) {
		text := p.tok.Origin
		p.next()
		return ast.NewAliasNode(text)
	}
	return ast.NewInvalidNode()
}

// YAML specification: [101] c-ns-anchor-property
func (p *parser) parseAnchorProperty() ast.Node {
	if p.tok.Type != token.AnchorType {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	p.next()
	if p.tok.Type == token.StringType && p.tok.ConformsCharSet(token.AnchorCharSetType) {
		anchor := ast.NewAnchorNode(p.tok.Origin)
		p.next()
		p.commit()
		return anchor
	}

	p.rollback()
	return ast.NewInvalidNode()
}

// YAML specification: [97] c-ns-tag-property
func (p *parser) parseTagProperty() ast.Node {
	p.setCheckpoint()
	// shorthand tag
	// YAML specification: [99] c-ns-shorthand-tag
	if ast.ValidNode(p.parseTagHandle()) && p.tok.Type == token.StringType &&
		p.tok.ConformsCharSet(token.TagCharSetType) {
		p.commit()
		text := p.tok.Origin
		p.next()
		return ast.NewTagNode(text)
	}
	p.rollback()

	if p.tok.Type != token.TagType {
		return ast.NewInvalidNode()
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
			return ast.NewTagNode(cutToken.Origin)
		}
	}

	// non specific tag
	// YAML specification: [100] c-non-specific-tag
	return ast.NewTagNode("")
}
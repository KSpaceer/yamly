package parser

import (
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/token"
)

// YAML specification: [79] s-l-comments
func (p *parser) parseComments() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	if !ast.ValidNode(p.parseSeparatedComment()) {
		p.rollback()
		if !p.startOfLine {
			return ast.NewInvalidNode()
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
	return ast.NewBasicNode(ast.CommentType)
}

// YAML specification: [77] s-b-comment
func (p *parser) parseSeparatedComment() ast.Node {
	if p.hasErrors() {
		return ast.NewInvalidNode()
	}
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
		return ast.NewInvalidNode()
	}
	p.next()
	return ast.NewBasicNode(ast.CommentType)
}

// YAML specification: [78] l-comment
func (p *parser) parseCommentLine() ast.Node {
	if p.hasErrors() || !ast.ValidNode(p.parseSeparateInLine()) {
		return ast.NewInvalidNode()
	}
	p.setCheckpoint()
	if !ast.ValidNode(p.parseCommentText()) {
		p.rollback()
	} else {
		p.commit()
	}
	if p.tok.Type != token.LineBreakType && p.tok.Type != token.EOFType {
		return ast.NewInvalidNode()
	}
	p.next()
	return ast.NewBasicNode(ast.CommentType)
}

// YAML specification: [75] c-nb-comment-text
func (p *parser) parseCommentText() ast.Node {
	if p.hasErrors() || p.tok.Type != token.CommentType {
		return ast.NewInvalidNode()
	}
	p.next()

	for token.IsNonBreak(p.tok) {
		p.next()
	}
	return ast.NewBasicNode(ast.CommentType)
}

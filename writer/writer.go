package writer

import (
	"bytes"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/chars"
)

const (
	defaultBasicIndentation = 0
	defaultIndendationDelta = 2
)

const nullValue = "null"

type writer struct {
	buf              bytes.Buffer
	errors           []error
	indentation      int
	indentationDelta int
}

func (w *writer) appendError(err error) {
	w.errors = append(w.errors, err)
}

func (w *writer) VisitStreamNode(n *ast.StreamNode) {
	for _, doc := range n.Documents() {
		w.buf.WriteString("---\n")
		doc.Accept(w)
		w.buf.WriteString("...\n")
	}
}

func (w *writer) VisitTagNode(n *ast.TagNode) {
	w.buf.WriteString(" !!")
	w.buf.WriteString(n.Text())
	w.buf.WriteByte(' ')
}

func (w *writer) VisitAnchorNode(n *ast.AnchorNode) {
	w.buf.WriteString(" &")
	w.buf.WriteString(n.Text())
	w.buf.WriteByte(' ')
}

func (w *writer) VisitAliasNode(n *ast.AliasNode) {
	w.buf.WriteString(" *")
	w.buf.WriteString(n.Text())
}

func (w *writer) VisitTextNode(n *ast.TextNode) {
	switch n.QuotingType() {
	case ast.AbsentQuotingType:
	case ast.SingleQuotingType:
	case ast.DoubleQuotingType:
		txt, err := chars.ConvertToYAMLDoubleQuotedString(n.Text())
		if err != nil {
			w.appendError(err)
		}
		w.buf.WriteByte('"')
		w.buf.WriteString(txt)
	}
}

func (w *writer) VisitSequenceNode(n *ast.SequenceNode) {
	w.increaseIndentation()
	for _, entry := range n.Entries() {
		w.buf.WriteByte('\n')
		w.writeIndentation()
		w.buf.WriteByte('-')
		w.buf.WriteByte(' ')
		entry.Accept(w)
	}
	w.decreaseIndentation()
}

func (w *writer) VisitMappingNode(n *ast.MappingNode) {
	w.increaseIndentation()
	for _, entry := range n.Entries() {
		w.buf.WriteByte('\n')
		w.writeIndentation()
		entry.Accept(w)
	}
	w.decreaseIndentation()
}

func (w *writer) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	key, value := n.Key(), n.Value()

	isComplexKey := key.Type() == ast.SequenceType || key.Type() == ast.MappingType

	if isComplexKey {
		w.buf.WriteString("? ")
		w.increaseIndentation()
	}

	key.Accept(w)

	if isComplexKey {
		w.decreaseIndentation()
	}
	w.buf.WriteString(": ")

	value.Accept(w)
}

func (w *writer) VisitNullNode(n *ast.NullNode) {
	w.buf.WriteString(nullValue)
}

func (w *writer) VisitPropertiesNode(n *ast.PropertiesNode) {
	tag, anchor := n.Tag(), n.Anchor()
	var tagValid bool
	if tagValid = ast.ValidNode(tag); tagValid {
		tag.Accept(w)
	}
	if ast.ValidNode(anchor) {
		if tagValid {
			w.buf.WriteByte(' ')
		}
		anchor.Accept(w)
	}
}

func (w *writer) VisitContentNode(n *ast.ContentNode) {
	properties, content := n.Properties(), n.Content()
	if ast.ValidNode(properties) {
		properties.Accept(w)
		w.buf.WriteByte(' ')
	}
	content.Accept(w)
}

func (w *writer) increaseIndentation() {
	w.indentation += w.indentationDelta
}

func (w *writer) decreaseIndentation() {
	w.indentation -= w.indentationDelta
}

func (w *writer) writeIndentation() {
	w.buf.Grow(w.indentation)
	for i := 0; i < w.indentation; i++ {
		w.buf.WriteByte(' ')
	}
}

package writer

import (
	"bytes"
	"errors"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/chars"
	"io"
	"strings"
)

const (
	defaultBasicIndentation = 0
	defaultIndendationDelta = 2
)

const nullValue = "null"

type Writer struct {
	buf                     *bytes.Buffer
	errors                  []error
	indentation             int
	indentationDelta        int
	preparedForComplexNodes bool
}

func NewWriter() *Writer {
	return &Writer{
		buf:              bytes.NewBuffer(nil),
		errors:           nil,
		indentation:      defaultBasicIndentation,
		indentationDelta: defaultIndendationDelta,
	}
}

func (w *Writer) WriteTo(dst io.Writer, ast ast.Node) error {
	ast.Accept(w)
	if w.hasErrors() {
		return w.error()
	}
	_, err := io.Copy(dst, w.buf)
	if err != nil {
		return err
	}
	w.buf.Reset()
	return nil
}

func (w *Writer) WriteString(ast ast.Node) (string, error) {
	var sb strings.Builder
	if err := w.WriteTo(&sb, ast); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (w *Writer) appendError(err error) {
	w.errors = append(w.errors, err)
}

func (w *Writer) hasErrors() bool {
	return len(w.errors) > 0
}

func (w *Writer) error() error {
	return errors.Join(w.errors...)
}

func (w *Writer) VisitStreamNode(n *ast.StreamNode) {
	for _, doc := range n.Documents() {
		w.buf.WriteString("---\n")
		doc.Accept(w)
		w.buf.WriteString("...\n")
	}
}

func (w *Writer) VisitTagNode(n *ast.TagNode) {
	w.buf.WriteString(" !!")
	w.buf.WriteString(n.Text())
	w.buf.WriteByte(' ')
}

func (w *Writer) VisitAnchorNode(n *ast.AnchorNode) {
	w.buf.WriteString(" &")
	w.buf.WriteString(n.Text())
	w.buf.WriteByte(' ')
}

func (w *Writer) VisitAliasNode(n *ast.AliasNode) {
	w.buf.WriteString(" *")
	w.buf.WriteString(n.Text())
}

func (w *Writer) VisitTextNode(n *ast.TextNode) {
	switch n.QuotingType() {
	case ast.AbsentQuotingType:
		w.buf.WriteString(n.Text())
	case ast.SingleQuotingType:
		txt, err := chars.ConvertToYAMLSingleQuotedString(n.Text())
		if err != nil {
			w.appendError(err)
		}
		w.buf.WriteByte('\'')
		w.buf.WriteString(txt)
		w.buf.WriteByte('\'')
	case ast.DoubleQuotingType:
		txt, err := chars.ConvertToYAMLDoubleQuotedString(n.Text())
		if err != nil {
			w.appendError(err)
		}
		w.buf.WriteByte('"')
		w.buf.WriteString(txt)
		w.buf.WriteByte('"')
	}
}

func (w *Writer) VisitSequenceNode(n *ast.SequenceNode) {
	w.maybeExecutePreparedLogic(n)
	for _, entry := range n.Entries() {
		w.maybeWriteIndentation()
		w.buf.WriteByte('-')
		w.buf.WriteByte(' ')
		w.increaseIndentation()
		entry.Accept(w)
		w.decreaseIndentation()
		w.maybeWriteLineBreak()
	}
}

func (w *Writer) VisitMappingNode(n *ast.MappingNode) {
	w.maybeExecutePreparedLogic(n)
	for _, entry := range n.Entries() {
		w.maybeWriteIndentation()
		entry.Accept(w)
		w.maybeWriteLineBreak()
	}
}

func (w *Writer) VisitMappingEntryNode(n *ast.MappingEntryNode) {
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
	w.buf.WriteByte(':')

	w.increaseIndentation()

	w.prepareForComplexNodes()

	value.Accept(w)
	w.decreaseIndentation()
}

func (w *Writer) VisitNullNode(n *ast.NullNode) {
	w.buf.WriteString(nullValue)
}

func (w *Writer) VisitPropertiesNode(n *ast.PropertiesNode) {
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

func (w *Writer) VisitContentNode(n *ast.ContentNode) {
	w.maybeExecutePreparedLogic(n)
	properties, content := n.Properties(), n.Content()
	if ast.ValidNode(properties) {
		properties.Accept(w)
		w.buf.WriteByte(' ')
	}
	content.Accept(w)
}

func (w *Writer) prepareForComplexNodes() {
	w.preparedForComplexNodes = true
}

func (w *Writer) maybeExecutePreparedLogic(n ast.Node) {
	switch n.Type() {
	case ast.SequenceType, ast.MappingType:
		w.maybeWriteLineBreak()
		w.preparedForComplexNodes = false
	case ast.ContentType:
		if ast.ValidNode(n.(*ast.ContentNode).Properties()) {
			w.buf.WriteByte(' ')
		}
	default:
		w.preparedForComplexNodes = false
	}
}

func (w *Writer) increaseIndentation() {
	w.indentation += w.indentationDelta
}

func (w *Writer) decreaseIndentation() {
	w.indentation -= w.indentationDelta
}

func (w *Writer) writeIndentation() {
	w.buf.Grow(w.indentation)
	for i := 0; i < w.indentation; i++ {
		w.buf.WriteByte(' ')
	}
}

func (w *Writer) hasWriteLineBreak() bool {
	bufData := w.buf.Bytes()
	return len(bufData) > 0 && bufData[len(bufData)-1] == '\n'
}

func (w *Writer) maybeWriteIndentation() {
	if w.hasWriteLineBreak() {
		w.writeIndentation()
	}
}

func (w *Writer) maybeWriteLineBreak() {
	if !w.hasWriteLineBreak() {
		w.buf.WriteByte('\n')
	}
}

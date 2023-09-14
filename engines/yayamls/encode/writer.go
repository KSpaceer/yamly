package encode

import (
	"bytes"
	"errors"
	"github.com/KSpaceer/yamly"
	"github.com/KSpaceer/yamly/engines/yayamls/ast"
	"github.com/KSpaceer/yamly/engines/yayamls/chars"
	"io"
	"strings"
)

const (
	defaultBasicIndentation = 0
	defaultIndendationDelta = 2
)

const nullValue = "null"

var _ yamly.TreeWriter[ast.Node] = (*ASTWriter)(nil)

type ASTWriter struct {
	buf              *bytes.Buffer
	errors           []error
	indentation      int
	indentationDelta int

	beforeComplex string
	beforeSimple  string

	metAnchors map[string]struct{}

	opts writeOptions
}

func NewASTWriter(opts ...WriteOption) *ASTWriter {
	w := ASTWriter{
		buf:              bytes.NewBuffer(nil),
		errors:           nil,
		indentation:      defaultBasicIndentation,
		indentationDelta: defaultIndendationDelta,
		metAnchors:       map[string]struct{}{},
	}

	for _, opt := range opts {
		opt(&w.opts)
	}

	return &w
}

type AnchorsKeeper interface {
	StoreAnchor(anchorName string)
	BindToLatestAnchor(n ast.Node)
	DereferenceAlias(alias string) (ast.Node, error)
}

type writeOptions struct {
	anchorsKeeper AnchorsKeeper
}

type WriteOption func(*writeOptions)

func WithAnchorsKeeper(ak AnchorsKeeper) WriteOption {
	return func(options *writeOptions) {
		options.anchorsKeeper = ak
	}
}

func (w *ASTWriter) WriteTo(dst io.Writer, ast ast.Node) error {
	if err := w.write(ast); err != nil {
		return err
	}
	_, err := io.Copy(dst, w.buf)
	if err != nil {
		return err
	}
	w.buf.Reset()
	return nil
}

func (w *ASTWriter) WriteString(ast ast.Node) (string, error) {
	var sb strings.Builder
	if err := w.WriteTo(&sb, ast); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (w *ASTWriter) WriteBytes(ast ast.Node) ([]byte, error) {
	if err := w.write(ast); err != nil {
		return nil, err
	}
	data := w.buf.Bytes()
	w.buf = bytes.NewBuffer(nil)
	return data, nil
}

func (w *ASTWriter) write(ast ast.Node) error {
	w.reset()

	ast.Accept(w)
	if w.hasErrors() {
		return w.error()
	}
	return nil
}

func (w *ASTWriter) appendError(err error) {
	w.errors = append(w.errors, err)
}

func (w *ASTWriter) hasErrors() bool {
	return len(w.errors) > 0
}

func (w *ASTWriter) error() error {
	return errors.Join(w.errors...)
}

func (w *ASTWriter) VisitStreamNode(n *ast.StreamNode) {
	for _, doc := range n.Documents() {
		w.buf.WriteString("---\n")
		doc.Accept(w)
		w.buf.WriteString("...\n")
	}
}

func (w *ASTWriter) VisitTagNode(n *ast.TagNode) {
	w.buf.WriteString("!!")
	w.buf.WriteString(n.Text())
}

func (w *ASTWriter) VisitAnchorNode(n *ast.AnchorNode) {
	anchor := n.Text()
	w.buf.WriteString("&")
	w.buf.WriteString(anchor)

	if w.opts.anchorsKeeper != nil {
		w.metAnchors[anchor] = struct{}{}
		w.opts.anchorsKeeper.StoreAnchor(anchor)
	}
}

func (w *ASTWriter) VisitAliasNode(n *ast.AliasNode) {
	alias := n.Text()
	_, hasWroteAnchor := w.metAnchors[alias]
	if w.opts.anchorsKeeper != nil && !hasWroteAnchor {
		anchored, err := w.opts.anchorsKeeper.DereferenceAlias(alias)
		if err != nil {
			w.appendError(err)
		} else if ast.ValidNode(anchored) {
			anchored.Accept(w)
		}
		return
	}

	w.writePreparedData(n)
	w.buf.WriteString("*")
	w.buf.WriteString(n.Text())
}

func (w *ASTWriter) VisitTextNode(n *ast.TextNode) {
	w.writePreparedData(n)
	switch txt := n.Text(); n.QuotingType() {
	case ast.AbsentQuotingType:
		if isMultiline(txt) {
			w.writeMultilineLiteralText(txt)
		} else {
			w.buf.WriteString(txt)
		}
	case ast.SingleQuotingType:
		w.writeSingleQuotedText(txt)
	case ast.DoubleQuotingType:
		w.writeDoubleQuotedText(txt)
	default:
		if isMultiline(txt) {
			w.writeDoubleQuotedText(txt)
		} else {
			w.buf.WriteString(txt)
		}
	}
}

func (w *ASTWriter) VisitSequenceNode(n *ast.SequenceNode) {
	w.writePreparedData(n)
	for _, entry := range n.Entries() {
		w.maybeWriteIndentation()
		w.buf.WriteByte('-')
		w.increaseIndentation()
		w.writeBeforeComplexElements(" ")
		w.writeBeforeSimpleElements(" ")
		entry.Accept(w)
		w.decreaseIndentation()
		w.maybeWriteLineBreak()
	}
}

func (w *ASTWriter) VisitMappingNode(n *ast.MappingNode) {
	w.writePreparedData(n)
	for _, entry := range n.Entries() {
		w.maybeWriteIndentation()
		entry.Accept(w)
		w.maybeWriteLineBreak()
	}
}

func (w *ASTWriter) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	key, value := n.Key(), n.Value()

	isComplexKey := isComplex(n.Key())
	if isComplexKey {
		w.buf.WriteString("? ")
		w.increaseIndentation()
	}

	key.Accept(w)

	if isComplexKey {
		w.decreaseIndentation()
	}
	w.buf.WriteByte(':')
	w.writeBeforeComplexElements("\n")
	w.writeBeforeSimpleElements(" ")

	w.increaseIndentation()

	value.Accept(w)
	w.decreaseIndentation()
}

func (w *ASTWriter) VisitNullNode(n *ast.NullNode) {
	w.writePreparedData(n)
	w.buf.WriteString(nullValue)
}

func (w *ASTWriter) VisitPropertiesNode(n *ast.PropertiesNode) {
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

func (w *ASTWriter) VisitContentNode(n *ast.ContentNode) {
	w.writePreparedData(n)
	properties, content := n.Properties(), n.Content()
	if ast.ValidNode(properties) {
		w.buf.WriteByte(' ')
		properties.Accept(w)

		if w.opts.anchorsKeeper != nil {
			w.opts.anchorsKeeper.BindToLatestAnchor(n)
		}
	}
	content.Accept(w)
}

func (w *ASTWriter) writeBeforeComplexElements(s string) {
	w.beforeComplex = s
}

func (w *ASTWriter) writeBeforeSimpleElements(s string) {
	w.beforeSimple = s
}

func (w *ASTWriter) writePreparedData(n ast.Node) {
	switch n.Type() {
	case ast.SequenceType, ast.MappingType:
		w.buf.WriteString(w.beforeComplex)
	case ast.ContentType:
		return
	default:
		w.buf.WriteString(w.beforeSimple)
	}
	w.beforeComplex = ""
	w.beforeSimple = ""
}

func (w *ASTWriter) increaseIndentation() {
	w.indentation += w.indentationDelta
}

func (w *ASTWriter) decreaseIndentation() {
	w.indentation -= w.indentationDelta
}

func (w *ASTWriter) writeIndentation() {
	w.buf.Grow(w.indentation)
	for i := 0; i < w.indentation; i++ {
		w.buf.WriteByte(' ')
	}
}

func (w *ASTWriter) hasWriteLineBreak() bool {
	bufData := w.buf.Bytes()
	return len(bufData) > 0 && bufData[len(bufData)-1] == '\n'
}

func (w *ASTWriter) maybeWriteIndentation() {
	if w.hasWriteLineBreak() {
		w.writeIndentation()
	}
}

func (w *ASTWriter) maybeWriteLineBreak() {
	if !w.hasWriteLineBreak() {
		w.buf.WriteByte('\n')
	}
}

func (w *ASTWriter) writeMultilineLiteralText(txt string) {
	lines := strings.Split(txt, "\n")
	chompingIndicator := chars.StripChompingCharacter
	if lines[len(lines)-1] == "" {
		chompingIndicator = chars.KeepChompingCharacter
	}
	w.buf.WriteByte('|')
	w.buf.WriteRune(chompingIndicator)
	for i := range lines {
		w.buf.WriteByte('\n')
		if lines[i] != "" {
			w.writeIndentation()
			w.buf.WriteString(lines[i])
		}
	}
}

func (w *ASTWriter) writeSingleQuotedText(txt string) {
	txt, err := chars.ConvertToYAMLSingleQuotedString(txt)
	if err != nil {
		w.appendError(err)
	}
	w.buf.WriteByte('\'')
	w.buf.WriteString(txt)
	w.buf.WriteByte('\'')
}

func (w *ASTWriter) writeDoubleQuotedText(txt string) {
	txt, err := chars.ConvertToYAMLDoubleQuotedString(txt)
	if err != nil {
		w.appendError(err)
	}
	w.buf.WriteByte('"')
	w.buf.WriteString(txt)
	w.buf.WriteByte('"')
}

func (w *ASTWriter) reset() {
	w.buf.Reset()
	w.errors = w.errors[:0]
	w.indentation = defaultBasicIndentation
	w.beforeSimple = ""
	w.beforeComplex = ""
	clear(w.metAnchors)
}

func isMultiline(s string) bool {
	return strings.ContainsRune(s, '\n')
}

func isComplex(n ast.Node) bool {
	switch n.Type() {
	case ast.SequenceType, ast.MappingType:
		return true
	case ast.ContentType:
		return isComplex(n.(*ast.ContentNode).Content())
	default:
		return false
	}
}

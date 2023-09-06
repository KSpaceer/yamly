package encode

import (
	"fmt"
	"github.com/KSpaceer/yayamls"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/parser"
	"github.com/KSpaceer/yayamls/schema"
	"time"
)

var _ yayamls.TreeBuilder[ast.Node] = (*ASTBuilder)(nil)

type ASTBuilder struct {
	root  ast.Node
	route []ast.Node

	opts builderOpts

	fatalError error
}

type builderOpts struct {
	unquoteOneliners bool
}

type ASTBuilderOption func(*builderOpts)

func WithUnquotedOneLineStrings() ASTBuilderOption {
	return func(opts *builderOpts) {
		opts.unquoteOneliners = true
	}
}

func NewASTBuilder(opts ...ASTBuilderOption) *ASTBuilder {
	b := ASTBuilder{}

	for _, opt := range opts {
		opt(&b.opts)
	}
	return &b
}

func (t *ASTBuilder) InsertInteger(val int64) {
	insertNonNullValue(t, val, schema.FromInteger, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertUnsigned(val uint64) {
	insertNonNullValue(t, val, schema.FromUnsignedInteger, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertBoolean(val bool) {
	insertNonNullValue(t, val, schema.FromBoolean, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertFloat(val float64) {
	insertNonNullValue(t, val, schema.FromFloat, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertString(val string) {
	quoting := ast.DoubleQuotingType
	if t.opts.unquoteOneliners && !isMultiline(val) {
		quoting = ast.AbsentQuotingType
	}
	insertNonNullValue(t, val, func(t string) string { return t }, quoting)
}

func (t *ASTBuilder) InsertTimestamp(val time.Time) {
	insertNonNullValue(t, val, schema.FromTimestamp, ast.DoubleQuotingType)
}

func (t *ASTBuilder) InsertNull() {
	t.insertNode(ast.NewNullNode(), false)
}

func (t *ASTBuilder) StartSequence() {
	sequence := ast.NewSequenceNode(nil)
	t.insertNode(sequence, true)
}

func (t *ASTBuilder) EndSequence() {
	if t.fatalError != nil {
		return
	}
	currentNode := t.currentNode()
	switch {
	case !ast.ValidNode(currentNode):
		t.fatalError = fmt.Errorf("failed to end sequence: got invalid node")
	case currentNode.Type() != ast.SequenceType:
		t.fatalError = fmt.Errorf("failed to end sequence: expected sequence, but currenlty at %s", currentNode.Type())
	default:
		t.popNode()
	}
}

func (t *ASTBuilder) StartMapping() {
	mapping := ast.NewMappingNode(nil)
	t.insertNode(mapping, true)
}

func (t *ASTBuilder) EndMapping() {
	if t.fatalError != nil {
		return
	}
	currentNode := t.currentNode()
	switch {
	case !ast.ValidNode(currentNode):
		t.fatalError = fmt.Errorf("failed to end mapping: got invalid node")
	case currentNode.Type() != ast.MappingType:
		t.fatalError = fmt.Errorf("failed to end mapping: expected mapping, but currenlty at %s", currentNode.Type())
	default:
		t.popNode()
	}
}

func (t *ASTBuilder) InsertRaw(data []byte, err error) {
	if t.fatalError != nil {
		return
	}
	if err != nil {
		t.fatalError = err
		return
	}
	tree, err := parser.ParseBytes(data, parser.WithOmitStream())
	if err != nil {
		t.fatalError = err
		return
	}
	if tree.Type() == ast.StreamType {
		t.fatalError = fmt.Errorf("failed to insert raw: expected single document, got stream of documents")
		return
	}
	t.insertNode(tree, false)
}

func (t *ASTBuilder) InsertRawText(text []byte, err error) {
	if t.fatalError != nil {
		return
	}
	if err != nil {
		t.fatalError = err
		return
	}
	t.InsertString(string(text))
}

func (t *ASTBuilder) Result() (ast.Node, error) {
	root := t.root
	err := t.fatalError
	t.route = t.route[:0]
	t.root = nil
	t.fatalError = nil
	return root, err
}

func insertNonNullValue[T any](t *ASTBuilder, val T, converter func(T) string, quotingType ast.QuotingType) {
	t.insertNode(
		ast.NewTextNode(
			converter(val),
			ast.WithQuotingType(quotingType),
		),
		false,
	)
}

func (t *ASTBuilder) insertNode(n ast.Node, pushToRoute bool) {
	if t.fatalError != nil {
		return
	}
	currentNode := t.currentNode()
	if !ast.ValidNode(currentNode) {
		t.pushNode(n)
		t.root = n
		return
	}

	switch currentNode.Type() {
	case ast.MappingType:
		mapping := currentNode.(*ast.MappingNode)
		entry := ast.NewMappingEntryNode(n, nil)
		mapping.AppendEntry(entry)
		t.pushNode(entry)
	case ast.MappingEntryType:
		entry := currentNode.(*ast.MappingEntryNode)
		entry.SetValue(n)
		t.popNode()
	case ast.SequenceType:
		sequence := currentNode.(*ast.SequenceNode)
		sequence.AppendEntry(n)
	default:
		t.fatalError = fmt.Errorf(
			"cannot insert new node to tree: currenlty at node with type %s",
			currentNode.Type(),
		)
	}
	if pushToRoute {
		t.pushNode(n)
	}
}

func (t *ASTBuilder) currentNode() ast.Node {
	if len(t.route) == 0 {
		return nil
	}
	return t.route[len(t.route)-1]
}

func (t *ASTBuilder) pushNode(n ast.Node) {
	t.route = append(t.route, n)
}

func (t *ASTBuilder) popNode() {
	if len(t.route) == 0 {
		return
	}
	t.route = t.route[:len(t.route)-1]
}

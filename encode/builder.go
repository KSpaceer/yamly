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

func (t *ASTBuilder) InsertInteger(val int64) error {
	return insertNonNullValue(t, val, schema.FromInteger, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertNullableInteger(val *int64) error {
	return insertNullableValue(t, val, schema.FromInteger, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertUnsigned(val uint64) error {
	return insertNonNullValue(t, val, schema.FromUnsignedInteger, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertNullableUnsigned(val *uint64) error {
	return insertNullableValue(t, val, schema.FromUnsignedInteger, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertBoolean(val bool) error {
	return insertNonNullValue(t, val, schema.FromBoolean, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertNullableBoolean(val *bool) error {
	return insertNullableValue(t, val, schema.FromBoolean, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertFloat(val float64) error {
	return insertNonNullValue(t, val, schema.FromFloat, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertNullableFloat(val *float64) error {
	return insertNullableValue(t, val, schema.FromFloat, ast.AbsentQuotingType)
}

func (t *ASTBuilder) InsertString(val string) error {
	quoting := ast.DoubleQuotingType
	if t.opts.unquoteOneliners && !isMultiline(val) {
		quoting = ast.AbsentQuotingType
	}
	return insertNonNullValue(t, val, func(t string) string { return t }, quoting)
}

func (t *ASTBuilder) InsertNullableString(val *string) error {
	quoting := ast.DoubleQuotingType
	if t.opts.unquoteOneliners && val != nil && !isMultiline(*val) {
		quoting = ast.AbsentQuotingType
	}
	return insertNullableValue(t, val, func(t string) string { return t }, quoting)
}

func (t *ASTBuilder) InsertTimestamp(val time.Time) error {
	return insertNonNullValue(t, val, schema.FromTimestamp, ast.DoubleQuotingType)
}

func (t *ASTBuilder) InsertNullableTimestamp(val *time.Time) error {
	return insertNullableValue(t, val, schema.FromTimestamp, ast.DoubleQuotingType)
}

func (t *ASTBuilder) InsertNull() error {
	return t.insertNode(ast.NewNullNode(), false)
}

func (t *ASTBuilder) StartSequence() error {
	sequence := ast.NewSequenceNode(nil)
	return t.insertNode(sequence, true)
}

func (t *ASTBuilder) EndSequence() error {
	currentNode := t.currentNode()
	switch {
	case !ast.ValidNode(currentNode):
		return fmt.Errorf("failed to end sequence: got invalid node")
	case currentNode.Type() != ast.SequenceType:
		return fmt.Errorf("failed to end sequence: expected sequence, but currenlty at %s", currentNode.Type())
	default:
		t.popNode()
		return nil
	}
}

func (t *ASTBuilder) StartMapping() error {
	mapping := ast.NewMappingNode(nil)
	return t.insertNode(mapping, true)
}

func (t *ASTBuilder) EndMapping() error {
	currentNode := t.currentNode()
	switch {
	case !ast.ValidNode(currentNode):
		return fmt.Errorf("failed to end mapping: got invalid node")
	case currentNode.Type() != ast.MappingType:
		return fmt.Errorf("failed to end mapping: expected mapping, but currenlty at %s", currentNode.Type())
	default:
		t.popNode()
		return nil
	}
}

func (t *ASTBuilder) InsertRaw(data []byte) error {
	tree, err := parser.ParseBytes(data, parser.WithOmitStream())
	if err != nil {
		return err
	}
	if tree.Type() == ast.StreamType {
		return fmt.Errorf("failed to insert raw: expected single document, got stream of documents")
	}
	return t.insertNode(tree, false)
}

func (t *ASTBuilder) Result() (ast.Node, error) {
	root := t.root
	t.route = t.route[:0]
	t.root = nil
	return root, nil
}

func insertNonNullValue[T any](t *ASTBuilder, val T, converter func(T) string, quotingType ast.QuotingType) error {
	return t.insertNode(
		ast.NewTextNode(
			converter(val),
			ast.WithQuotingType(quotingType),
		),
		false,
	)
}

func insertNullableValue[T any](
	t *ASTBuilder,
	val *T,
	converter func(T) string,
	quotingType ast.QuotingType,
) error {
	var newNode ast.Node

	if val == nil {
		newNode = ast.NewNullNode()
	} else {
		newNode = ast.NewTextNode(
			converter(*val),
			ast.WithQuotingType(quotingType),
		)
	}

	return t.insertNode(newNode, false)
}

func (t *ASTBuilder) insertNode(n ast.Node, pushToRoute bool) error {
	currentNode := t.currentNode()
	if !ast.ValidNode(currentNode) {
		t.pushNode(n)
		t.root = n
		return nil
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
		return fmt.Errorf(
			"cannot insert new node to tree: currenlty at node with type %s",
			currentNode.Type(),
		)
	}
	if pushToRoute {
		t.pushNode(n)
	}
	return nil
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

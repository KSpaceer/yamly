package encode

import (
	"fmt"
	"github.com/KSpaceer/yamly"
	"github.com/KSpaceer/yamly/engines/yayamls/ast"
	"github.com/KSpaceer/yamly/engines/yayamls/parser"
	"github.com/KSpaceer/yamly/engines/yayamls/schema"
	"time"
)

var _ yamly.TreeBuilder[ast.Node] = (*ASTBuilder)(nil)

// ASTBuilder implements yamly.TreeBuilder
type ASTBuilder struct {
	root  ast.Node
	route []ast.Node

	opts builderOpts

	fatalError error
}

type builderOpts struct {
	unquoteOneliners bool
}

// ASTBuilderOption allows to modify ASTBuilder behavior
type ASTBuilderOption func(*builderOpts)

// WithUnquotedOneLineStrings make ASTBuilder create nodes for one-line strings
// with absent quoting style (instead of default double quoting style)
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

func (b *ASTBuilder) InsertInteger(val int64) {
	insertNonNullValue(b, val, schema.FromInteger, ast.AbsentQuotingType)
}

func (b *ASTBuilder) InsertUnsigned(val uint64) {
	insertNonNullValue(b, val, schema.FromUnsignedInteger, ast.AbsentQuotingType)
}

func (b *ASTBuilder) InsertBoolean(val bool) {
	insertNonNullValue(b, val, schema.FromBoolean, ast.AbsentQuotingType)
}

func (b *ASTBuilder) InsertFloat(val float64) {
	insertNonNullValue(b, val, schema.FromFloat, ast.AbsentQuotingType)
}

func (b *ASTBuilder) InsertString(val string) {
	quoting := ast.DoubleQuotingType
	if b.opts.unquoteOneliners && !isMultiline(val) {
		quoting = ast.AbsentQuotingType
	}
	insertNonNullValue(b, val, func(t string) string { return t }, quoting)
}

func (b *ASTBuilder) InsertTimestamp(val time.Time) {
	insertNonNullValue(b, val, schema.FromTimestamp, ast.DoubleQuotingType)
}

func (b *ASTBuilder) InsertNull() {
	b.insertNode(ast.NewNullNode(), false)
}

func (b *ASTBuilder) StartSequence() {
	sequence := ast.NewSequenceNode(nil)
	b.insertNode(sequence, true)
}

func (b *ASTBuilder) EndSequence() {
	if b.fatalError != nil {
		return
	}
	currentNode := b.currentNode()
	switch {
	case !ast.ValidNode(currentNode):
		b.fatalError = fmt.Errorf("failed to end sequence: got invalid node")
	case currentNode.Type() != ast.SequenceType:
		b.fatalError = fmt.Errorf("failed to end sequence: expected sequence, but currenlty at %s", currentNode.Type())
	default:
		b.popNode()
	}
}

func (b *ASTBuilder) StartMapping() {
	mapping := ast.NewMappingNode(nil)
	b.insertNode(mapping, true)
}

func (b *ASTBuilder) EndMapping() {
	if b.fatalError != nil {
		return
	}
	currentNode := b.currentNode()
	switch {
	case !ast.ValidNode(currentNode):
		b.fatalError = fmt.Errorf("failed to end mapping: got invalid node")
	case currentNode.Type() != ast.MappingType:
		b.fatalError = fmt.Errorf("failed to end mapping: expected mapping, but currently at %s", currentNode.Type())
	default:
		b.popNode()
	}
}

func (b *ASTBuilder) InsertRaw(data []byte, err error) {
	if b.fatalError != nil {
		return
	}
	if err != nil {
		b.fatalError = err
		return
	}
	tree, err := parser.ParseBytes(data, parser.WithOmitStream())
	if err != nil {
		b.fatalError = err
		return
	}
	if tree.Type() == ast.StreamType {
		b.fatalError = fmt.Errorf("failed to insert raw: expected single document, got stream of documents")
		return
	}
	b.insertNode(tree, false)
}

func (b *ASTBuilder) InsertRawText(text []byte, err error) {
	if b.fatalError != nil {
		return
	}
	if err != nil {
		b.fatalError = err
		return
	}
	b.InsertString(string(text))
}

func (b *ASTBuilder) Result() (ast.Node, error) {
	root := b.root
	err := b.fatalError
	b.route = b.route[:0]
	b.root = nil
	b.fatalError = nil
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

func (b *ASTBuilder) insertNode(n ast.Node, pushToRoute bool) {
	if b.fatalError != nil {
		return
	}
	currentNode := b.currentNode()
	if !ast.ValidNode(currentNode) {
		b.pushNode(n)
		b.root = n
		return
	}

	switch currentNode.Type() {
	case ast.MappingType:
		mapping := currentNode.(*ast.MappingNode)
		entry := ast.NewMappingEntryNode(n, nil)
		mapping.AppendEntry(entry)
		b.pushNode(entry)
	case ast.MappingEntryType:
		entry := currentNode.(*ast.MappingEntryNode)
		entry.SetValue(n)
		b.popNode()
	case ast.SequenceType:
		sequence := currentNode.(*ast.SequenceNode)
		sequence.AppendEntry(n)
	default:
		b.fatalError = fmt.Errorf(
			"cannot insert new node to tree: currenlty at node with type %s",
			currentNode.Type(),
		)
	}
	if pushToRoute {
		b.pushNode(n)
	}
}

func (b *ASTBuilder) currentNode() ast.Node {
	if len(b.route) == 0 {
		return nil
	}
	return b.route[len(b.route)-1]
}

func (b *ASTBuilder) pushNode(n ast.Node) {
	b.route = append(b.route, n)
}

func (b *ASTBuilder) popNode() {
	if len(b.route) == 0 {
		return
	}
	b.route = b.route[:len(b.route)-1]
}

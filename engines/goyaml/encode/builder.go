package encode

import (
	"fmt"
	"github.com/KSpaceer/yamly"
	"github.com/KSpaceer/yamly/engines/pkg/schema"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

var _ yamly.TreeBuilder[*yaml.Node] = (*ASTBuilder)(nil)

type ASTBuilder struct {
	root  *yaml.Node
	route []*yaml.Node

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

func (b *ASTBuilder) InsertInteger(val int64) {
	insertNonNullValue(b, val, schema.FromInteger, 0)
}

func (b *ASTBuilder) InsertUnsigned(val uint64) {
	insertNonNullValue(b, val, schema.FromUnsignedInteger, 0)
}

func (b *ASTBuilder) InsertBoolean(val bool) {
	insertNonNullValue(b, val, schema.FromBoolean, 0)
}

func (b *ASTBuilder) InsertFloat(val float64) {
	insertNonNullValue(b, val, schema.FromFloat, 0)
}

func (b *ASTBuilder) InsertString(val string) {
	style := yaml.DoubleQuotedStyle
	if b.opts.unquoteOneliners && !isMultiline(val) {
		style = 0
	}
	insertNonNullValue(b, val, func(s string) string { return s }, style)
}

func (b *ASTBuilder) InsertTimestamp(val time.Time) {
	insertNonNullValue(b, val, schema.FromTimestamp, yaml.DoubleQuotedStyle)
}

func (b *ASTBuilder) InsertNull() {
	b.insertNode(
		&yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!null",
			Value: "null",
		},
		false,
	)
}

func (b *ASTBuilder) StartSequence() {
	b.insertNode(
		&yaml.Node{
			Kind: yaml.SequenceNode,
		},
		true,
	)
}

func (b *ASTBuilder) EndSequence() {
	if b.fatalError != nil {
		return
	}
	currentNode := b.currentNode()
	switch {
	case currentNode == nil:
		b.fatalError = fmt.Errorf("failed to end sequence: got invalid node")
	case currentNode.Kind != yaml.SequenceNode:
		b.fatalError = fmt.Errorf(
			"failed to end sequence: expected sequence, but currenlty at %s",
			kindToString(currentNode.Kind),
		)
	default:
		b.popNode()
	}
}

func (b *ASTBuilder) StartMapping() {
	b.insertNode(
		&yaml.Node{
			Kind: yaml.MappingNode,
		},
		true,
	)
}

func (b *ASTBuilder) EndMapping() {
	if b.fatalError != nil {
		return
	}
	currentNode := b.currentNode()
	switch {
	case currentNode == nil:
		b.fatalError = fmt.Errorf("failed to end mapping: got invalid node")
	case currentNode.Kind != yaml.MappingNode:
		b.fatalError = fmt.Errorf(
			"failed to end mapping: expected mapping, but currenlty at %s",
			kindToString(currentNode.Kind),
		)
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
	var n yaml.Node
	if err := yaml.Unmarshal(data, &n); err != nil {
		b.fatalError = err
		return
	}
	var result *yaml.Node
	if n.Kind == yaml.DocumentNode {
		result = n.Content[0]
	} else {
		result = &n
	}
	b.insertNode(result, false)
}

func (b *ASTBuilder) InsertRawText(data []byte, err error) {
	if b.fatalError != nil {
		return
	}
	if err != nil {
		b.fatalError = err
		return
	}
	b.InsertString(string(data))
}

func (b *ASTBuilder) Result() (*yaml.Node, error) {
	root := b.root
	err := b.fatalError
	b.route = b.route[:0]
	b.root = nil
	b.fatalError = nil
	return root, err
}

func insertNonNullValue[T any](b *ASTBuilder, val T, converter func(T) string, style yaml.Style) {
	b.insertNode(
		&yaml.Node{
			Kind:  yaml.ScalarNode,
			Style: style,
			Value: converter(val),
		},
		false,
	)
}

func (b *ASTBuilder) insertNode(n *yaml.Node, pushToRoute bool) {
	if b.fatalError != nil {
		return
	}
	currentNode := b.currentNode()
	if currentNode == nil {
		b.pushNode(n)
		b.root = n
		return
	}

	switch currentNode.Kind {
	case yaml.MappingNode, yaml.SequenceNode:
		currentNode.Content = append(currentNode.Content, n)
	default:
		b.fatalError = fmt.Errorf(
			"cannot insert new node to tree: currently at node with kind %s",
			kindToString(currentNode.Kind),
		)
	}
	if pushToRoute {
		b.pushNode(n)
	}
}

func (b *ASTBuilder) currentNode() *yaml.Node {
	if len(b.route) == 0 {
		return nil
	}
	return b.route[len(b.route)-1]
}

func (b *ASTBuilder) pushNode(n *yaml.Node) {
	b.route = append(b.route, n)
}

func (b *ASTBuilder) popNode() {
	if len(b.route) == 0 {
		return
	}
	b.route = b.route[:len(b.route)-1]
}

func isMultiline(s string) bool {
	return strings.ContainsRune(s, '\n')
}

func kindToString(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		return "unknown"
	}
}

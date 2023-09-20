// Package schema is used to derive types from YAML source. For more information see shared schema package.
package schema

import (
	"time"

	"github.com/KSpaceer/yamly/engines/pkg/schema"
	"github.com/KSpaceer/yamly/engines/yayamls/ast"
)

const (
	MergeKey = schema.MergeKey
)

func IsNull(n ast.Node) bool {
	switch n.Type() {
	case ast.NullType:
		return true
	case ast.TextType:
		txtNode := n.(*ast.TextNode) // nolint: forcetypeassert
		txt := txtNode.Text()
		return schema.IsNull(txt) &&
			(txtNode.QuotingType() == ast.UnknownQuotingType ||
				txtNode.QuotingType() == ast.AbsentQuotingType)
	}
	return false
}

func IsBoolean(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode) // nolint: forcetypeassert
	txt := txtNode.Text()
	return schema.IsBoolean(txt)
}

func FromBoolean(val bool) string {
	return schema.FromBoolean(val)
}

func ToBoolean(src string) (bool, error) {
	return schema.ToBoolean(src)
}

func IsInteger(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode) // nolint: forcetypeassert
	txt := txtNode.Text()
	return schema.IsInteger(txt)
}

func IsUnsignedInteger(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode) // nolint: forcetypeassert
	txt := txtNode.Text()
	return schema.IsUnsignedInteger(txt)
}

func FromInteger(val int64) string {
	return schema.FromInteger(val)
}

func ToInteger(src string, bitSize int) (int64, error) {
	return schema.ToInteger(src, bitSize)
}

func FromUnsignedInteger(val uint64) string {
	return schema.FromUnsignedInteger(val)
}

func ToUnsignedInteger(src string, bitSize int) (uint64, error) {
	return schema.ToUnsignedInteger(src, bitSize)
}

func IsFloat(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode) // nolint: forcetypeassert
	txt := txtNode.Text()
	return schema.IsFloat(txt)
}

func FromFloat(val float64) string {
	return schema.FromFloat(val)
}

func ToFloat(src string, bitSize int) (float64, error) {
	return schema.ToFloat(src, bitSize)
}

func IsBinary(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode) // nolint: forcetypeassert
	txt := txtNode.Text()
	return schema.IsBinary(txt)
}

func IsMergeKey(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode) // nolint: forcetypeassert
	txt := txtNode.Text()
	return txt == MergeKey
}

func IsTimestamp(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode) // nolint: forcetypeassert
	txt := txtNode.Text()
	return schema.IsTimestamp(txt)
}

func FromTimestamp(val time.Time) string {
	return schema.FromTimestamp(val)
}

func ToTimestamp(src string) (t time.Time, err error) {
	return schema.ToTimestamp(src)
}

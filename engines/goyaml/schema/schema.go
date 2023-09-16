package schema

import (
	"github.com/KSpaceer/yamly/engines/pkg/schema"
	"gopkg.in/yaml.v3"
	"time"
)

const (
	MergeKey = schema.MergeKey
)

func IsNull(n *yaml.Node) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}
	return n.ShortTag() == "!!null"
}

func IsBoolean(n *yaml.Node) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}
	return schema.IsBoolean(n.Value)
}

func FromBoolean(val bool) string {
	return schema.FromBoolean(val)
}

func ToBoolean(src string) (bool, error) {
	return schema.ToBoolean(src)
}

func IsInteger(n *yaml.Node) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}
	return schema.IsInteger(n.Value)
}

func IsUnsignedInteger(n *yaml.Node) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}
	return schema.IsUnsignedInteger(n.Value)
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

func IsFloat(n *yaml.Node) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}
	return schema.IsFloat(n.Value)
}

func FromFloat(val float64) string {
	return schema.FromFloat(val)
}

func ToFloat(src string, bitSize int) (float64, error) {
	return schema.ToFloat(src, bitSize)
}

func IsBinary(n *yaml.Node) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}
	return n.ShortTag() == "!!binary" || schema.IsBinary(n.Value)
}

func IsMergeKey(n *yaml.Node) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}
	return n.Tag == "!!merge" || n.Value == MergeKey
}

func IsTimestamp(n *yaml.Node) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}
	return schema.IsTimestamp(n.Value)
}

func FromTimestamp(val time.Time) string {
	return schema.FromTimestamp(val)
}

func ToTimestamp(src string) (t time.Time, err error) {
	return schema.ToTimestamp(src)
}

package reader

import (
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/schema"
)

type expectancyResult int8

const (
	expectancyResultUnknown expectancyResult = iota
	expectancyResultMatch
	expectancyResultDeny
	expectancyResultContinue
)

type expectNullable struct {
	underlying expecter
}

func (e expectNullable) process(n ast.Node) expectancyResult {
	if schema.IsNull(n) {
		return expectancyResultMatch
	}
	return e.underlying.process(n)
}

func (e expectNullable) name() string {
	return e.underlying.name() + "Nullable"
}

type expectInteger struct{}

func (e expectInteger) process(n ast.Node) expectancyResult {
	switch n.Type() {
	case ast.TextType:
		result := expectancyResultDeny
		if schema.IsInteger(n) {
			result = expectancyResultMatch
		}
		return result
	case ast.ContentType, ast.PropertiesType, ast.AnchorType, ast.TagType, ast.StreamType:
		return expectancyResultContinue
	default:
		return expectancyResultDeny
	}
}

func (e expectInteger) name() string {
	return "ExpectInteger"
}

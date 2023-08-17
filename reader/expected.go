package reader

import (
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/schema"
)

type expectNullable struct {
	underlying expecter
}

func (e expectNullable) process(n ast.Node, prev visitingResult) visitingResult {
	if schema.IsNull(n) {
		switch prev {
		case visitingResultUnknown, visitingResultContinue, visitingResultDeny:
			return visitingResultMatch
		default:
			return visitingResultContinue
		}
	}
	return e.underlying.process(n, prev)
}

func (e expectNullable) name() string {
	return e.underlying.name() + "Nullable"
}

type expectInteger struct{}

func (e expectInteger) process(n ast.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsInteger)
}

func (e expectInteger) name() string {
	return "ExpectInteger"
}

type expectBoolean struct{}

func (e expectBoolean) name() string {
	return "ExpectBoolean"
}

func (e expectBoolean) process(n ast.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsBoolean)
}

func processTerminalNode(n ast.Node, prev visitingResult, predicate func(ast.Node) bool) visitingResult {
	switch n.Type() {
	case ast.TextType:
		if predicate(n) {
			switch prev {
			case visitingResultUnknown, visitingResultContinue, visitingResultDeny:
				return visitingResultMatch
			default:
				return visitingResultContinue
			}
		}
		return visitingResultDeny
	case ast.MappingType, ast.SequenceType:
		switch prev {
		case visitingResultMatch, visitingResultContinue:
			return visitingResultContinue
		default:
			return visitingResultDeny
		}
	case ast.ContentType, ast.PropertiesType, ast.AnchorType, ast.TagType, ast.StreamType, ast.MappingEntryType:
		return visitingResultContinue
	default:
		return visitingResultDeny
	}
}

type expectMapping struct{}

func (e expectMapping) name() string {
	return "ExpectMapping"
}

func (e expectMapping) process(n ast.Node, prev visitingResult) visitingResult {
	switch n.Type() {
	case ast.MappingType:
		switch prev {
		case visitingResultUnknown, visitingResultDeny:
			return visitingResultMatch
		case visitingResultContinue, visitingResultMatch:
			return visitingResultContinue
		}
	case ast.SequenceType:
		switch prev {
		case visitingResultMatch, visitingResultContinue:
			return visitingResultContinue
		default:
			return visitingResultDeny
		}
	case ast.ContentType, ast.PropertiesType, ast.AnchorType, ast.TagType, ast.StreamType, ast.MappingEntryType:
		return visitingResultContinue
	default:
		return visitingResultDeny
	}
}

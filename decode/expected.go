package decode

import (
	"github.com/KSpaceer/yamly/ast"
	"github.com/KSpaceer/yamly/schema"
)

type expectNull struct{}

func (expectNull) name() string {
	return "ExpectNull"
}

func (expectNull) process(n ast.Node, prev visitingResult) visitingResult {
	if schema.IsNull(n) {
		switch prev.conclusion {
		case visitingConclusionUnknown, visitingConclusionContinue, visitingConclusionDeny:
			return visitingResult{
				conclusion: visitingConclusionConsume,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		}
	}

	switch n.Type() {
	case ast.ContentType, ast.PropertiesType, ast.AnchorType, ast.TagType, ast.StreamType, ast.MappingEntryType:
		return visitingResult{
			conclusion: visitingConclusionContinue,
		}
	default:
		switch prev.conclusion {
		case visitingConclusionMatch, visitingConclusionContinue:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionDeny,
			}
		}
	}
}

type expectInteger struct{}

func (expectInteger) process(n ast.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsInteger)
}

func (expectInteger) name() string {
	return "ExpectInteger"
}

type expectBoolean struct{}

func (expectBoolean) name() string {
	return "ExpectBoolean"
}

func (expectBoolean) process(n ast.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsBoolean)
}

type expectFloat struct{}

func (expectFloat) name() string {
	return "ExpectFloat"
}

func (expectFloat) process(n ast.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsFloat)
}

type expectString struct {
	checkForNull bool
}

func (expectString) name() string {
	return "ExpectString"
}

func (e expectString) process(n ast.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, e.isString)
}

func (e expectString) isString(n ast.Node) bool {
	if e.checkForNull {
		if schema.IsNull(n) {
			return false
		}
	}
	return n.Type() == ast.TextType
}

type expectTimestamp struct{}

func (expectTimestamp) name() string {
	return "ExpectTimestamp"
}

func (expectTimestamp) process(n ast.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsTimestamp)
}

func processTerminalNode(n ast.Node, prev visitingResult, predicate func(ast.Node) bool) visitingResult {
	switch n.Type() {
	case ast.TextType:
		if predicate(n) {
			switch prev.conclusion {
			case visitingConclusionUnknown, visitingConclusionContinue, visitingConclusionDeny:
				return visitingResult{
					conclusion: visitingConclusionConsume,
					action:     visitingActionExtract,
				}
			default:
				return visitingResult{
					conclusion: visitingConclusionContinue,
				}
			}
		}
		return visitingResult{
			conclusion: visitingConclusionDeny,
		}
	case ast.MappingType, ast.SequenceType:
		switch prev.conclusion {
		case visitingConclusionMatch, visitingConclusionConsume, visitingConclusionContinue:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionDeny,
			}
		}
	case ast.ContentType, ast.PropertiesType, ast.AnchorType, ast.TagType, ast.StreamType, ast.MappingEntryType:
		return visitingResult{
			conclusion: visitingConclusionContinue,
		}
	default:
		return visitingResult{
			conclusion: visitingConclusionDeny,
		}
	}
}

type expectSequence struct{}

func (e expectSequence) name() string {
	return "ExpectSequence"
}

func (e expectSequence) process(n ast.Node, prev visitingResult) visitingResult {
	switch n.Type() {
	case ast.SequenceType:
		switch prev.conclusion {
		case visitingConclusionUnknown, visitingConclusionDeny:
			return visitingResult{
				conclusion: visitingConclusionMatch,
				action:     visitingActionExtract,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		}
	case ast.MappingType, ast.TextType:
		switch prev.conclusion {
		case visitingConclusionMatch, visitingConclusionConsume, visitingConclusionContinue:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionDeny,
			}
		}
	case ast.ContentType, ast.PropertiesType, ast.AnchorType,
		ast.TagType, ast.StreamType, ast.MappingEntryType:
		return visitingResult{
			conclusion: visitingConclusionContinue,
		}
	}
	return visitingResult{
		conclusion: visitingConclusionDeny,
	}
}

type expectMapping struct{}

func (e expectMapping) name() string {
	return "ExpectMapping"
}

func (e expectMapping) process(n ast.Node, prev visitingResult) visitingResult {
	switch n.Type() {
	case ast.MappingType:
		switch prev.conclusion {
		case visitingConclusionUnknown, visitingConclusionDeny:
			return visitingResult{
				conclusion: visitingConclusionMatch,
				action:     visitingActionExtract,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		}
	case ast.SequenceType, ast.TextType:
		switch prev.conclusion {
		case visitingConclusionMatch, visitingConclusionConsume, visitingConclusionContinue:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionDeny,
			}
		}
	case ast.ContentType, ast.PropertiesType, ast.AnchorType,
		ast.TagType, ast.StreamType, ast.MappingEntryType:
		return visitingResult{
			conclusion: visitingConclusionContinue,
		}
	}
	return visitingResult{
		conclusion: visitingConclusionDeny,
	}
}

type expectAny struct{}

func (expectAny) name() string {
	return "ExpectAny"
}

func (expectAny) process(n ast.Node, prev visitingResult) visitingResult {
	switch n.Type() {
	case ast.MappingType, ast.SequenceType, ast.TextType, ast.NullType:
		switch prev.conclusion {
		case visitingConclusionMatch, visitingConclusionConsume, visitingConclusionContinue:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionMatch,
			}
		}
	default:
		return visitingResult{
			conclusion: visitingConclusionContinue,
		}
	}
}

type expectRaw struct{}

func (expectRaw) name() string {
	return "ExpectRaw"
}

func (expectRaw) process(n ast.Node, prev visitingResult) visitingResult {
	switch n.Type() {
	case ast.ContentType, ast.PropertiesType, ast.TagType, ast.AnchorType:
		return visitingResult{
			conclusion: visitingConclusionContinue,
		}
	default:
		switch prev.conclusion {
		case visitingConclusionMatch, visitingConclusionConsume, visitingConclusionContinue:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionMatch,
			}
		}
	}

}

type expectSkip struct{}

func (expectSkip) name() string {
	return "ExpectSkip"
}

func (expectSkip) process(n ast.Node, prev visitingResult) visitingResult {
	switch n.Type() {
	case ast.ContentType, ast.PropertiesType, ast.TagType, ast.AnchorType:
		return visitingResult{
			conclusion: visitingConclusionContinue,
		}
	default:
		switch prev.conclusion {
		case visitingConclusionMatch, visitingConclusionConsume, visitingConclusionContinue:
			return visitingResult{
				conclusion: visitingConclusionContinue,
			}
		default:
			return visitingResult{
				conclusion: visitingConclusionConsume,
			}
		}
	}
}

package decode

import (
	"github.com/KSpaceer/yamly/engines/goyaml/schema"
	"gopkg.in/yaml.v3"
)

type expectNull struct{}

func (expectNull) name() string {
	return "ExpectNull"
}

func (expectNull) process(n *yaml.Node, prev visitingResult) visitingResult {
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

type expectInteger struct{}

func (expectInteger) name() string {
	return "ExpectInteger"
}

func (expectInteger) process(n *yaml.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsInteger)
}

type expectBoolean struct{}

func (expectBoolean) name() string {
	return "ExpectBoolean"
}

func (expectBoolean) process(n *yaml.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsBoolean)
}

type expectFloat struct{}

func (expectFloat) name() string {
	return "ExpectFloat"
}

func (expectFloat) process(n *yaml.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsFloat)
}

type expectString struct {
	checkForNull bool
}

func (expectString) name() string {
	return "ExpectString"
}

func (e expectString) process(n *yaml.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, e.isString)
}

func (e expectString) isString(n *yaml.Node) bool {
	if e.checkForNull {
		if schema.IsNull(n) {
			return false
		}
	}
	return n.Kind == yaml.ScalarNode
}

type expectTimestamp struct{}

func (expectTimestamp) name() string {
	return "ExpectTimestamp"
}

func (expectTimestamp) process(n *yaml.Node, prev visitingResult) visitingResult {
	return processTerminalNode(n, prev, schema.IsTimestamp)
}

func processTerminalNode(n *yaml.Node, prev visitingResult, predicate func(*yaml.Node) bool) visitingResult {
	switch n.Kind {
	case yaml.ScalarNode:
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
	case yaml.SequenceNode, yaml.MappingNode:
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
	default:
		return visitingResult{
			conclusion: visitingConclusionDeny,
		}
	}
}

type expectSequence struct{}

func (expectSequence) name() string {
	return "ExpectSequence"
}

func (expectSequence) process(n *yaml.Node, prev visitingResult) visitingResult {
	switch n.Kind {
	case yaml.SequenceNode:
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
	default:
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
	}
}

type expectMapping struct{}

func (expectMapping) name() string {
	return "ExpectMapping"
}

func (expectMapping) process(n *yaml.Node, prev visitingResult) visitingResult {
	switch n.Kind {
	case yaml.MappingNode:
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
	default:
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
	}
}

type expectAny struct{}

func (expectAny) name() string {
	return "ExpectAny"
}

func (expectAny) process(_ *yaml.Node, prev visitingResult) visitingResult {
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

type expectRaw struct{}

func (expectRaw) name() string {
	return "ExpectRaw"
}

func (expectRaw) process(_ *yaml.Node, prev visitingResult) visitingResult {
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

type expectNode = expectRaw

type expectSkip struct{}

func (expectSkip) name() string {
	return "ExpectSkip"
}

func (expectSkip) process(_ *yaml.Node, prev visitingResult) visitingResult {
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

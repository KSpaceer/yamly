package decode

import (
	"errors"
	"fmt"

	"github.com/KSpaceer/yamly/engines/goyaml/schema"
	"gopkg.in/yaml.v3"
)

type anyBuilder struct {
	value any

	extractString bool
	stringValue   string

	extractMergeMap bool
	mergeMap        map[string]any

	errors []error
}

func (a *anyBuilder) extractAnyValue(n *yaml.Node) (any, error) {
	a.visitNode(n)
	if a.hasErrors() {
		return nil, errors.Join(a.errors...)
	}
	return a.value, nil
}

func (a *anyBuilder) visitNode(n *yaml.Node) {
	if n == nil {
		return
	}

	switch n.Kind {
	case yaml.DocumentNode:
		a.visitNode(n.Content[0])
	case yaml.SequenceNode:
		a.visitSequenceNode(n)
	case yaml.MappingNode:
		a.visitMappingNode(n)
	case yaml.ScalarNode:
		a.visitScalarNode(n)
	case yaml.AliasNode:
		a.visitNode(n.Alias)
	}
}

func (a *anyBuilder) visitSequenceNode(n *yaml.Node) {
	s := make([]any, 0, len(n.Content))

	savedExtractString := a.extractString
	a.extractString = false

	for _, entry := range n.Content {
		a.visitNode(entry)
		if !a.hasErrors() {
			s = append(s, a.value)
		}
	}

	a.extractString = savedExtractString

	if a.extractString {
		a.stringValue = fmt.Sprint(s)
	} else {
		a.value = s
	}
}

func (a *anyBuilder) visitMappingNode(n *yaml.Node) {
	m := make(map[string]any, len(n.Content)/2)
	entriesAmount := len(n.Content)
	for i := 0; i < entriesAmount; i += 2 {
		key := n.Content[i]
		var value *yaml.Node
		if i+1 >= entriesAmount {
			value = &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!null",
				Value: "null",
			}
		} else {
			value = n.Content[i+1]
		}

		a.processMappingPair(key, value)

		if !a.hasErrors() {
			if a.mergeMap == nil {
				m[a.stringValue] = a.value
			} else {
				mergeMaps(m, a.mergeMap)
				a.mergeMap = nil
			}
		}
	}

	switch {
	case a.extractMergeMap:
		a.mergeMap = m
	case a.extractString:
		a.stringValue = fmt.Sprint(m)
	default:
		a.value = m
	}
}

func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
}

func (a *anyBuilder) processMappingPair(keyNode, valueNode *yaml.Node) {
	savedExtractString, savedExtractMergeMap := a.extractString, a.extractMergeMap

	a.extractString = true
	a.visitNode(keyNode)
	if a.stringValue == schema.MergeKey {
		a.extractMergeMap = true
	}

	key := a.stringValue
	a.extractString = false

	a.visitNode(valueNode)

	a.stringValue = key
	a.extractString = savedExtractString
	a.extractMergeMap = savedExtractMergeMap
}

func (a *anyBuilder) visitScalarNode(n *yaml.Node) {
	if a.extractString {
		a.stringValue = n.Value
	} else {
		var err error
		switch {
		case schema.IsNull(n):
			a.value = nil
		case schema.IsTimestamp(n):
			a.value, err = schema.ToTimestamp(n.Value)
		case schema.IsUnsignedInteger(n):
			a.value, err = schema.ToUnsignedInteger(n.Value, 64)
		case schema.IsInteger(n):
			a.value, err = schema.ToInteger(n.Value, 64)
		case schema.IsFloat(n):
			a.value, err = schema.ToFloat(n.Value, 64)
		case schema.IsBoolean(n):
			a.value, err = schema.ToBoolean(n.Value)
		default:
			a.value = n.Value
		}
		if err != nil {
			a.appendError(err)
		}
	}
}

func (a *anyBuilder) appendError(err error) {
	a.errors = append(a.errors, err)
}

func (a *anyBuilder) hasErrors() bool {
	return len(a.errors) > 0
}

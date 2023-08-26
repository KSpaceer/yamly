package reader

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/schema"
)

type anyBuilder struct {
	value any

	extractString bool
	stringValue   string

	extractMergeMap bool
	mergeMap        map[string]any

	anchors *anchorsKeeper
	errors  []error
}

func newAnyBuilder(anchors *anchorsKeeper) anyBuilder {
	return anyBuilder{
		anchors: anchors,
	}
}

func (a *anyBuilder) extractAnyValue(n ast.Node) (any, error) {
	a.visitNode(n)
	if a.hasErrors() {
		return nil, errors.Join(a.errors...)
	}
	return a.value, nil
}

func (a *anyBuilder) VisitStreamNode(*ast.StreamNode) {
	a.appendError(fmt.Errorf("didn't expect stream node while building value of any type"))
}

func (*anyBuilder) VisitTagNode(*ast.TagNode) {}

func (a *anyBuilder) VisitAnchorNode(n *ast.AnchorNode) {
	a.anchors.markAsLatestVisited(n.Text())
}

func (a *anyBuilder) VisitAliasNode(n *ast.AliasNode) {
	anchored, err := a.anchors.dereferenceAlias(n.Text())
	if err != nil {
		a.appendError(err)
	} else {
		a.visitNode(anchored)
	}
}

func (a *anyBuilder) VisitTextNode(n *ast.TextNode) {
	if a.extractString {
		a.stringValue = n.Text()
	} else {
		a.extractAnyValueFromText(n)
	}
}

func (a *anyBuilder) VisitSequenceNode(n *ast.SequenceNode) {
	entries := n.Entries()
	s := make([]any, 0, len(entries))

	savedExtractString := a.extractString
	a.extractString = false

	for _, entry := range entries {
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

func (a *anyBuilder) VisitMappingNode(n *ast.MappingNode) {
	entries := n.Entries()
	m := make(map[string]any, len(entries))
	for _, entry := range entries {
		a.visitNode(entry)
		if !a.hasErrors() {
			if a.mergeMap == nil {
				m[a.stringValue] = a.value
			} else {
				for k, v := range a.mergeMap {
					if _, ok := m[k]; !ok {
						m[k] = v
					}
				}
				a.mergeMap = nil
			}
		}
	}
	if a.extractMergeMap {
		a.mergeMap = m
	} else if a.extractString {
		a.stringValue = fmt.Sprint(m)
	} else {
		a.value = m
	}
}

func (a *anyBuilder) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	savedExtractString, savedExtractMergeMap := a.extractString, a.extractMergeMap

	a.extractString = true
	a.visitNode(n.Key())
	if a.stringValue == schema.MergeKey {
		a.extractMergeMap = true
	}

	key := a.stringValue
	a.extractString = false

	a.visitNode(n.Value())

	a.stringValue = key
	a.extractString = savedExtractString
	a.extractMergeMap = savedExtractMergeMap
}

func (a *anyBuilder) VisitNullNode(*ast.NullNode) {
	if a.extractString {
		a.stringValue = "null"
	} else {
		a.value = nil
	}
}

func (a *anyBuilder) VisitPropertiesNode(n *ast.PropertiesNode) {
	a.visitNode(n.Anchor())
}

func (a *anyBuilder) VisitContentNode(n *ast.ContentNode) {
	a.visitNode(n.Properties())
	content := n.Content()
	a.anchors.maybeBindToLatestVisited(content)
	a.visitNode(content)
}

func (a *anyBuilder) extractAnyValueFromText(n *ast.TextNode) {
	var err error
	switch {
	case schema.IsNull(n):
		a.value = nil
		return
	case schema.IsTimestamp(n):
		a.value, err = schema.ToTimestamp(n.Text())
	case schema.IsUnsignedInteger(n):
		a.value, err = schema.ToUnsignedInteger(n.Text())
	case schema.IsInteger(n):
		a.value, err = schema.ToInteger(n.Text())
	case schema.IsFloat(n):
		a.value, err = schema.ToFloat(n.Text())
	case schema.IsBoolean(n):
		a.value, err = schema.ToBoolean(n.Text())
	default:
		a.value = n.Text()
	}
	if err != nil {
		a.appendError(err)
	}
}

func (a *anyBuilder) visitNode(n ast.Node) {
	if ast.ValidNode(n) {
		n.Accept(a)
	}
}

func (a *anyBuilder) appendError(err error) {
	a.errors = append(a.errors, err)
}

func (a *anyBuilder) hasErrors() bool {
	return len(a.errors) > 0
}

package decode

import "gopkg.in/yaml.v3"

// rawTransformer is used to transform AST before serializing to bytes.
// The goal of transformation is to provide full context to the subtree,
// like dereferencing aliases etc.
// For example we have next YAML:
//
// root:
//
//	outer: &anchor value
//	inner:
//	  first: *anchor
//	  second: null
//
// transformer is used on "inner" element
// transformer will replace *anchor with value, because there is no anchor "&anchor"
// in "inner" element.
type rawTransformer struct {
	topLevelAnchor string
	metAnchors     map[string]struct{}

	dereferencedAliases []dereferencedAliasInfo
}

func newRawTransformer() rawTransformer {
	return rawTransformer{
		metAnchors: map[string]struct{}{},
	}
}

type dereferencedAliasInfo struct {
	parent *yaml.Node
	alias  *yaml.Node
	idx    int
}

func (rt *rawTransformer) transform(n *yaml.Node) *yaml.Node {
	rt.topLevelAnchor = n.Anchor
	n.Anchor = ""
	rt.visitNode(n)
	if n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

func (rt *rawTransformer) visitNode(n *yaml.Node) {
	if n.Anchor != "" {
		rt.metAnchors[n.Anchor] = struct{}{}
	}
	switch n.Kind {
	case yaml.AliasNode:
		rt.visitNode(n.Alias)
	case yaml.DocumentNode, yaml.MappingNode, yaml.SequenceNode:
		for i, child := range n.Content {
			if child.Kind == yaml.AliasNode {
				if _, metAnchor := rt.metAnchors[child.Value]; !metAnchor {
					n.Content[i] = child.Alias
					rt.dereferencedAliases = append(rt.dereferencedAliases, dereferencedAliasInfo{
						parent: n,
						alias:  child,
						idx:    i,
					})
				}
			}
			rt.visitNode(child)
		}
	}
}

func (rt *rawTransformer) restore(n *yaml.Node) {
	n.Anchor = rt.topLevelAnchor
	for _, deref := range rt.dereferencedAliases {
		deref.parent.Content[deref.idx] = deref.alias
	}
}

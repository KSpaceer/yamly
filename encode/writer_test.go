package encode_test

import (
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/encode"
	"testing"
)

func TestWriteString(t *testing.T) {
	type tcase struct {
		name      string
		ast       ast.Node
		expected  string
		expectErr bool

		anchors     mockAnchorsKeeper
		withAnchors bool
	}

	tcases := []tcase{
		{
			name:     "empty YAML",
			ast:      ast.NewStreamNode(nil),
			expected: "",
		},
		{
			name: "simple mapping entry",
			ast: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("key"),
							ast.NewTextNode("value"),
						),
					}),
				),
			}),
			expected: "---\nkey: value\n...\n",
		},
		{
			name: "simple sequence",
			ast: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewTextNode("value1"),
						ast.NewTextNode("value2"),
					},
				)),
			}),
			expected: "---\n- value1\n- value2\n...\n",
		},
		{
			name: "simple mapping with sequence and simple value",
			ast: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("sequence"),
							ast.NewContentNode(
								nil,
								ast.NewSequenceNode(
									[]ast.Node{
										ast.NewTextNode("sequencevalue1"),
										ast.NewTextNode("sequencevalue2"),
									},
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("simple"),
							ast.NewTextNode("value"),
						),
					},
				)),
			}),
			expected: "---\nsequence:\n  - sequencevalue1\n  - sequencevalue2\nsimple: value\n...\n",
		},
		{
			name: "simple sequence with mapping and simple single quoted value",
			ast: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewMappingNode(
							[]ast.Node{
								ast.NewMappingEntryNode(
									ast.NewTextNode("key1"),
									ast.NewTextNode("value1"),
								),
								ast.NewMappingEntryNode(
									ast.NewTextNode("key2"),
									ast.NewTextNode("value2"),
								),
							},
						),
						ast.NewTextNode("quotedvalue", ast.WithQuotingType(ast.SingleQuotingType)),
					},
				)),
			}),
			expected: "---\n- key1: value1\n  key2: value2\n- 'quotedvalue'\n...\n",
		},
		{
			name: "nested mapping with properties",
			ast: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("mapping"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("map"),
									ast.NewAnchorNode("ref"),
								),
								ast.NewMappingNode(
									[]ast.Node{
										ast.NewMappingEntryNode(
											ast.NewTextNode("innerkey"),
											ast.NewTextNode("innervalue"),
										),
									},
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("aliased"),
							ast.NewAliasNode("ref"),
						),
					},
				)),
			}),
			expected: "---\nmapping: !!map &ref\n  innerkey: innervalue\naliased: *ref\n...\n",
		},
		{
			name: "multiline quote absent string",
			ast: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewContentNode(
							ast.NewPropertiesNode(
								nil,
								ast.NewAnchorNode("lit"),
							),
							ast.NewTextNode("firstrow\nsecondrow\n\n", ast.WithQuotingType(ast.AbsentQuotingType)),
						),
						ast.NewContentNode(
							ast.NewPropertiesNode(
								ast.NewTagNode("dq"),
								nil,
							),
							ast.NewTextNode("firstrow\nsecondrow\n\n", ast.WithQuotingType(ast.DoubleQuotingType)),
						),
					},
				)),
			}),
			expected: "---\n- &lit |+\n  firstrow\n  secondrow\n\n- !!dq \"firstrow\\nsecondrow\\n\\n\"\n...\n",
		},
		{
			name: "complex key",
			ast: ast.NewStreamNode([]ast.Node{
				ast.NewMappingNode([]ast.Node{
					ast.NewMappingEntryNode(
						ast.NewSequenceNode([]ast.Node{
							ast.NewTextNode("a"),
							ast.NewTextNode("b"),
						}),
						ast.NewSequenceNode([]ast.Node{
							ast.NewTextNode("c"),
							ast.NewTextNode("d"),
						}),
					),
				}),
			}),
			expected: "---\n? - a\n  - b\n:\n  - c\n  - d\n...\n",
		},
		{
			name: "null mapping",
			ast: ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewNullNode(),
					ast.NewNullNode(),
				),
			}),
			expected: "null: null\n",
		},
		{
			name: "preloaded anchor",
			ast: ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewTextNode("key"),
					ast.NewAliasNode("anc"),
				),
			}),
			expected: "key: value\n",
			anchors: mockAnchorsKeeper{
				m: map[string]ast.Node{
					"anc": ast.NewTextNode("value"),
				},
			},
			withAnchors: true,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			var opts []encode.WriteOption
			if tc.withAnchors {
				opts = append(opts, encode.WithAnchorsKeeper(&tc.anchors))
			}

			w := encode.NewASTWriter(opts...)

			result, err := w.WriteString(tc.ast)
			if err != nil {
				if tc.expectErr {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected %q, but got %q", tc.expected, result)
			}
		})
	}
}

type mockAnchorsKeeper struct {
	m      map[string]ast.Node
	latest string
}

func (m *mockAnchorsKeeper) StoreAnchor(anchorName string) {
	m.latest = anchorName
}

func (m *mockAnchorsKeeper) BindToLatestAnchor(n ast.Node) {
	if m.latest != "" {
		m.m[m.latest] = n
		m.latest = ""
	}
}

func (m *mockAnchorsKeeper) DereferenceAlias(alias string) (ast.Node, error) {
	anchored, ok := m.m[alias]
	if !ok {
		return nil, fmt.Errorf("alias %q not found", alias)
	}
	return anchored, nil
}

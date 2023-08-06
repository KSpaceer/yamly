package writer_test

import (
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/writer"
	"testing"
)

func TestWriteString(t *testing.T) {
	type tcase struct {
		name      string
		ast       ast.Node
		expected  string
		expectErr bool
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
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			w := writer.NewWriter()
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

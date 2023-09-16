package encode_test

import (
	"encoding/json"
	"github.com/KSpaceer/yamly"
	"github.com/KSpaceer/yamly/engines/yayamls/ast"
	"github.com/KSpaceer/yamly/engines/yayamls/ast/astutils"
	"github.com/KSpaceer/yamly/engines/yayamls/encode"
	"math"
	"strings"
	"testing"
	"time"
)

func TestBuilder_Simple(t *testing.T) {
	type tcase struct {
		name      string
		calls     func(b yamly.TreeBuilder[ast.Node])
		expected  ast.Node
		expectErr bool
	}

	tcases := []tcase{
		{
			name: "simple integer",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.InsertInteger(15)
			},
			expected: ast.NewTextNode("15"),
		},
		{
			name: "simple unsigned",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.InsertUnsigned(0xFF)
			},
			expected: ast.NewTextNode("255"),
		},
		{
			name: "simple boolean",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.InsertBoolean(true)
			},
			expected: ast.NewTextNode("true"),
		},
		{
			name: "simple float",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.InsertFloat(33e6)
			},
			expected: ast.NewTextNode("3.3e+07"),
		},
		{
			name: "simple string",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.InsertString("Null")
			},
			expected: ast.NewTextNode("Null", ast.WithQuotingType(ast.DoubleQuotingType)),
		},
		{
			name: "simple timestamp",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.InsertTimestamp(
					time.Date(2023, 8, 27, 21, 42, 0, 0, time.UTC),
				)
			},
			expected: ast.NewTextNode(
				time.Date(2023, 8, 27, 21, 42, 0, 0, time.UTC).Format(time.RFC3339),
				ast.WithQuotingType(ast.DoubleQuotingType),
			),
		},
		{
			name: "null",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.InsertNull()
			},
			expected: ast.NewNullNode(),
		},
		{
			name: "simple sequence",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.StartSequence()
				b.EndSequence()
			},
			expected: ast.NewSequenceNode(nil),
		},
		{
			name: "simple mapping",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.StartMapping()
				b.EndMapping()
			},
			expected: ast.NewMappingNode(nil),
		},
		{
			name: "ending complex node without starting",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.EndSequence()
			},
			expectErr: true,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := encode.NewASTBuilder()
			tc.calls(b)
			result, err := b.Result()
			if err != nil {
				if !tc.expectErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			} else if tc.expectErr {
				t.Errorf("expected error, but got nil")
			}
			compareAST(t, tc.expected, result)
		})
	}
}

func TestBuilder_Complex(t *testing.T) {
	type tcase struct {
		name      string
		calls     func(b yamly.TreeBuilder[ast.Node])
		expected  ast.Node
		expectErr bool
	}

	tcases := []tcase{
		{
			name: "mapping with one pair",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.StartMapping()
				b.InsertString("key")
				b.InsertString("value")
				b.EndMapping()
			},
			expected: ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewTextNode("key", ast.WithQuotingType(ast.DoubleQuotingType)),
					ast.NewTextNode("value", ast.WithQuotingType(ast.DoubleQuotingType)),
				),
			}),
		},
		{
			name: "sequence with two entries",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.StartSequence()
				b.InsertString("val1")
				b.InsertString("val2")
				b.EndSequence()
			},
			expected: ast.NewSequenceNode([]ast.Node{
				ast.NewTextNode("val1", ast.WithQuotingType(ast.DoubleQuotingType)),
				ast.NewTextNode("val2", ast.WithQuotingType(ast.DoubleQuotingType)),
			}),
		},
		{
			name: "struct-like mapping",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.StartMapping()
				{
					b.InsertString("name")
					b.InsertString("name")
					b.InsertString("score")
					b.InsertUnsigned(250)
					b.InsertString("subscription")
					b.InsertBoolean(true)
					b.InsertString("nested")
					b.StartMapping()
					{
						b.InsertString("inner")
						b.InsertFloat(math.Inf(1))
						b.InsertString("seq")
						b.InsertNull()
					}
					b.EndMapping()
				}
				b.EndMapping()
			},
			expected: ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewTextNode("name", ast.WithQuotingType(ast.DoubleQuotingType)),
					ast.NewTextNode("name", ast.WithQuotingType(ast.DoubleQuotingType)),
				),
				ast.NewMappingEntryNode(
					ast.NewTextNode("score", ast.WithQuotingType(ast.DoubleQuotingType)),
					ast.NewTextNode("250", ast.WithQuotingType(ast.AbsentQuotingType)),
				),
				ast.NewMappingEntryNode(
					ast.NewTextNode("subscription", ast.WithQuotingType(ast.DoubleQuotingType)),
					ast.NewTextNode("true", ast.WithQuotingType(ast.AbsentQuotingType)),
				),
				ast.NewMappingEntryNode(
					ast.NewTextNode("nested", ast.WithQuotingType(ast.DoubleQuotingType)),
					ast.NewMappingNode([]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("inner", ast.WithQuotingType(ast.DoubleQuotingType)),
							ast.NewTextNode(".inf", ast.WithQuotingType(ast.AbsentQuotingType)),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("seq", ast.WithQuotingType(ast.DoubleQuotingType)),
							ast.NewNullNode(),
						),
					}),
				),
			}),
		},
		{
			name: "raw insertion",
			calls: func(b yamly.TreeBuilder[ast.Node]) {
				b.StartMapping()
				b.InsertString("key")
				b.InsertString("value")
				b.InsertString("raw")
				b.InsertRaw(json.Marshal([]int64{1, 2, 3}))
				b.EndMapping()
			},
			expected: ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewTextNode("key", ast.WithQuotingType(ast.DoubleQuotingType)),
					ast.NewTextNode("value", ast.WithQuotingType(ast.DoubleQuotingType)),
				),
				ast.NewMappingEntryNode(
					ast.NewTextNode("raw", ast.WithQuotingType(ast.DoubleQuotingType)),
					ast.NewSequenceNode([]ast.Node{
						ast.NewTextNode("1", ast.WithQuotingType(ast.AbsentQuotingType)),
						ast.NewTextNode("2", ast.WithQuotingType(ast.AbsentQuotingType)),
						ast.NewTextNode("3", ast.WithQuotingType(ast.AbsentQuotingType)),
					}),
				),
			}),
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := encode.NewASTBuilder()
			tc.calls(b)
			result, err := b.Result()
			if err != nil {
				if !tc.expectErr {
					t.Errorf("unexpected error: %v", err)
				}
			} else if tc.expectErr {
				t.Errorf("expected error, but got nil")
			}
			compareAST(t, tc.expected, result)
		})
	}
}

func TestBuilder_InsertRaw(t *testing.T) {
	type tcase struct {
		name     string
		src      []byte
		expected ast.Node
	}

	tcases := []tcase{
		{
			name:     "simple value",
			src:      []byte("'22'"),
			expected: ast.NewTextNode("22", ast.WithQuotingType(ast.SingleQuotingType)),
		},
		{
			name: "simple mapping",
			src:  []byte("key: value"),
			expected: ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewTextNode("key"),
					ast.NewTextNode("value"),
				),
			}),
		},
		{
			name: "simple sequence",
			src:  []byte("[1, 2, 3]"),
			expected: ast.NewSequenceNode([]ast.Node{
				ast.NewTextNode("1"),
				ast.NewTextNode("2"),
				ast.NewTextNode("3"),
			}),
		},
		{
			name: "sequence of mappings",
			src:  []byte("[{1: 2}, {3: 4}, {5: 6}]"),
			expected: ast.NewSequenceNode([]ast.Node{
				ast.NewMappingNode([]ast.Node{
					ast.NewMappingEntryNode(
						ast.NewTextNode("1"),
						ast.NewTextNode("2"),
					),
				}),
				ast.NewMappingNode([]ast.Node{
					ast.NewMappingEntryNode(
						ast.NewTextNode("3"),
						ast.NewTextNode("4"),
					),
				}),
				ast.NewMappingNode([]ast.Node{
					ast.NewMappingEntryNode(
						ast.NewTextNode("5"),
						ast.NewTextNode("6"),
					),
				}),
			}),
		},
		{
			name: "mapping of sequences",
			src:  []byte("[1, 2, 3]: [4, 5, 6]"),
			expected: ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewSequenceNode([]ast.Node{
						ast.NewTextNode("1"),
						ast.NewTextNode("2"),
						ast.NewTextNode("3"),
					}),
					ast.NewSequenceNode([]ast.Node{
						ast.NewTextNode("4"),
						ast.NewTextNode("5"),
						ast.NewTextNode("6"),
					}),
				),
			}),
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := encode.NewASTBuilder()
			b.StartMapping()
			b.InsertString("raw")
			b.InsertRaw(tc.src, nil)
			b.EndMapping()
			result, err := b.Result()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			compareAST(t, ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewTextNode("raw", ast.WithQuotingType(ast.DoubleQuotingType)),
					tc.expected,
				),
			}), result)
		})
	}
}

func TestBuilder_K8SManifest(t *testing.T) {
	/*
			apiVersion: v1
		    kind: PersistentVolumeClaim
		    metadata:
		      name: pvc-claim
		    spec:
		      storageClassName: manual
		      accessModes:
		        - ReadWriteOnce
		      resources:
		        requests:
		          storage: 3Gi
	*/

	expected := ast.NewMappingNode([]ast.Node{
		ast.NewMappingEntryNode(
			ast.NewTextNode("apiVersion"),
			ast.NewTextNode("v1"),
		),
		ast.NewMappingEntryNode(
			ast.NewTextNode("kind"),
			ast.NewTextNode("PersistentVolumeClaim"),
		),
		ast.NewMappingEntryNode(
			ast.NewTextNode("metadata"),
			ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewTextNode("name"),
					ast.NewTextNode("pvc-claim"),
				),
			}),
		),
		ast.NewMappingEntryNode(
			ast.NewTextNode("spec"),
			ast.NewMappingNode([]ast.Node{
				ast.NewMappingEntryNode(
					ast.NewTextNode("storageClassName"),
					ast.NewTextNode("manual"),
				),
				ast.NewMappingEntryNode(
					ast.NewTextNode("accessModes"),
					ast.NewSequenceNode([]ast.Node{
						ast.NewTextNode("ReadWriteOnce"),
					}),
				),
				ast.NewMappingEntryNode(
					ast.NewTextNode("resources"),
					ast.NewMappingNode([]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("requests"),
							ast.NewMappingNode([]ast.Node{
								ast.NewMappingEntryNode(
									ast.NewTextNode("storage"),
									ast.NewTextNode("3Gi"),
								),
							}),
						),
					}),
				),
			}),
		),
	})

	b := encode.NewASTBuilder(encode.WithUnquotedOneLineStrings())
	var build func(b yamly.TreeBuilder[ast.Node])
	build = func(b yamly.TreeBuilder[ast.Node]) {
		b.StartMapping()
		{
			b.InsertString("apiVersion")
			b.InsertString("v1")
			b.InsertString("kind")
			b.InsertString("PersistentVolumeClaim")
			b.InsertString("metadata")
			b.StartMapping()
			{
				b.InsertString("name")
				b.InsertString("pvc-claim")
			}
			b.EndMapping()
			b.InsertString("spec")
			b.StartMapping()
			{
				b.InsertString("storageClassName")
				b.InsertString("manual")
				b.InsertString("accessModes")
				b.StartSequence()
				{
					b.InsertString("ReadWriteOnce")
				}
				b.EndSequence()
				b.InsertString("resources")
				b.StartMapping()
				{
					b.InsertString("requests")
					b.StartMapping()
					{
						b.InsertString("storage")
						b.InsertString("3Gi")
					}
					b.EndMapping()
				}
				b.EndMapping()
			}
			b.EndMapping()
		}
		b.EndMapping()
	}
	build(b)

	result, err := b.Result()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	compareAST(t, expected, result)
}

func compareAST(t *testing.T, expectedAST, gotAST ast.Node) {
	t.Helper()

	cmp := astutils.NewComparator()
	printer := astutils.NewPrinter()

	if !cmp.Equal(expectedAST, gotAST) {
		var s strings.Builder
		if err := printer.Print(expectedAST, &s); err != nil {
			t.Fatalf("failed to print AST: %v", err)
		}
		expected := s.String()
		s.Reset()
		if err := printer.Print(gotAST, &s); err != nil {
			t.Fatalf("failed to print AST: %v", err)
		}
		got := s.String()
		s.Reset()
		t.Errorf("AST are not equal:\n\nExpected:\n%s\n\nGot:\n%s\n", expected, got)
		t.Fail()
	}
}

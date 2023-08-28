package encode_test

import (
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/ast/astutils"
	"github.com/KSpaceer/yayamls/encode"
	"math"
	"strings"
	"testing"
	"time"
)

func TestBuilder_Simple(t *testing.T) {
	type tcase struct {
		name      string
		calls     func(b encode.TreeBuilder[ast.Node]) error
		expected  ast.Node
		expectErr bool
	}

	tcases := []tcase{
		{
			name: "simple integer",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertInteger(15)
			},
			expected: ast.NewTextNode("15"),
		},
		{
			name: "simple nullable integer",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertNullableInteger(nil)
			},
			expected: ast.NewNullNode(),
		},
		{
			name: "simple unsigned",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertUnsigned(0xFF)
			},
			expected: ast.NewTextNode("255"),
		},
		{
			name: "simple nullable unsigned",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertNullableUnsigned(nil)
			},
			expected: ast.NewNullNode(),
		},
		{
			name: "simple boolean",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertBoolean(true)
			},
			expected: ast.NewTextNode("true"),
		},
		{
			name: "simple nullable boolean",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertNullableBoolean(nil)
			},
			expected: ast.NewNullNode(),
		},
		{
			name: "simple float",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertFloat(33e6)
			},
			expected: ast.NewTextNode("3.3e+07"),
		},
		{
			name: "simple nullable float",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertNullableFloat(nil)
			},
			expected: ast.NewNullNode(),
		},
		{
			name: "simple string",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertString("Null")
			},
			expected: ast.NewTextNode("Null", ast.WithQuotingType(ast.DoubleQuotingType)),
		},
		{
			name: "simple timestamp",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.InsertTimestamp(
					time.Date(2023, 8, 27, 21, 42, 0, 0, time.UTC),
				)
			},
			expected: ast.NewTextNode(
				time.Date(2023, 8, 27, 21, 42, 0, 0, time.UTC).Format(time.RFC3339),
				ast.WithQuotingType(ast.DoubleQuotingType),
			),
		},
		{
			name: "simple sequence",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				if err := b.StartSequence(); err != nil {
					return err
				}
				return b.EndSequence()
			},
			expected: ast.NewSequenceNode(nil),
		},
		{
			name: "simple mapping",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				if err := b.StartMapping(); err != nil {
					return err
				}
				return b.EndMapping()
			},
			expected: ast.NewMappingNode(nil),
		},
		{
			name: "ending complex node without starting",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				return b.EndSequence()
			},
			expectErr: true,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := encode.NewASTBuilder()
			if err := tc.calls(b); err != nil {
				if !tc.expectErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			} else if tc.expectErr {

				t.Errorf("expected error, but got nil")
			}
			result, err := b.Result()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			compareAST(t, tc.expected, result)
		})
	}
}

func TestBuilder_Complex(t *testing.T) {
	type tcase struct {
		name      string
		calls     func(b encode.TreeBuilder[ast.Node]) error
		expected  ast.Node
		expectErr bool
	}

	tcases := []tcase{
		{
			name: "mapping with one pair",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				if err := b.StartMapping(); err != nil {
					return err
				}
				if err := b.InsertString("key"); err != nil {
					return err
				}
				if err := b.InsertString("value"); err != nil {
					return err
				}
				return b.EndMapping()
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
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				if err := b.StartSequence(); err != nil {
					return err
				}
				if err := b.InsertString("val1"); err != nil {
					return err
				}
				if err := b.InsertString("val2"); err != nil {
					return err
				}
				return b.EndSequence()
			},
			expected: ast.NewSequenceNode([]ast.Node{
				ast.NewTextNode("val1", ast.WithQuotingType(ast.DoubleQuotingType)),
				ast.NewTextNode("val2", ast.WithQuotingType(ast.DoubleQuotingType)),
			}),
		},
		{
			name: "struct-like mapping",
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				if err := b.StartMapping(); err != nil {
					return err
				}
				{
					if err := b.InsertString("name"); err != nil {
						return err
					}
					if err := b.InsertString("name"); err != nil {
						return err
					}
					if err := b.InsertString("score"); err != nil {
						return err
					}
					score := uint64(250)
					if err := b.InsertNullableUnsigned(&score); err != nil {
						return err
					}
					if err := b.InsertString("subscription"); err != nil {
						return err
					}
					if err := b.InsertBoolean(true); err != nil {
						return err
					}
					if err := b.InsertString("nested"); err != nil {
						return err
					}
					if err := b.StartMapping(); err != nil {
						return err
					}
					{
						if err := b.InsertString("inner"); err != nil {
							return err
						}
						if err := b.InsertFloat(math.Inf(1)); err != nil {
							return err
						}
						if err := b.InsertString("seq"); err != nil {
							return err
						}
						if err := b.InsertNull(); err != nil {
							return err
						}
					}
					if err := b.EndMapping(); err != nil {
						return err
					}
				}
				return b.EndMapping()
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
			calls: func(b encode.TreeBuilder[ast.Node]) error {
				if err := b.StartMapping(); err != nil {
					return err
				}
				if err := b.InsertString("key"); err != nil {
					return err
				}
				value := "value"
				if err := b.InsertNullableString(&value); err != nil {
					return err
				}
				if err := b.InsertString("raw"); err != nil {
					return err
				}
				if err := b.InsertRaw([]byte("[1, 2, 3]")); err != nil {
					return err
				}
				return b.EndMapping()
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
			if err := tc.calls(b); err != nil {
				if !tc.expectErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			} else if tc.expectErr {
				t.Errorf("expected error, but got nil")
			}
			result, err := b.Result()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
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
			if err := b.StartMapping(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err := b.InsertString("raw"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err := b.InsertRaw(tc.src); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err := b.EndMapping(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
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
	var build func(b encode.TreeBuilder[ast.Node]) error
	build = func(b encode.TreeBuilder[ast.Node]) error {
		if err := b.StartMapping(); err != nil {
			return err
		}
		{
			if err := b.InsertString("apiVersion"); err != nil {
				return err
			}
			if err := b.InsertString("v1"); err != nil {
				return err
			}
			if err := b.InsertString("kind"); err != nil {
				return err
			}
			if err := b.InsertString("PersistentVolumeClaim"); err != nil {
				return err
			}
			if err := b.InsertString("metadata"); err != nil {
				return err
			}
			if err := b.StartMapping(); err != nil {
				return err
			}
			{
				if err := b.InsertString("name"); err != nil {
					return err
				}
				if err := b.InsertString("pvc-claim"); err != nil {
					return err
				}
			}
			if err := b.EndMapping(); err != nil {
				return err
			}
			if err := b.InsertString("spec"); err != nil {
				return err
			}
			if err := b.StartMapping(); err != nil {
				return err
			}
			{
				if err := b.InsertString("storageClassName"); err != nil {
					return err
				}
				if err := b.InsertString("manual"); err != nil {
					return err
				}
				if err := b.InsertString("accessModes"); err != nil {
					return err
				}
				if err := b.StartSequence(); err != nil {
					return err
				}
				{
					if err := b.InsertString("ReadWriteOnce"); err != nil {
						return err
					}
				}
				if err := b.EndSequence(); err != nil {
					return err
				}
				if err := b.InsertString("resources"); err != nil {
					return err
				}
				if err := b.StartMapping(); err != nil {
					return err
				}
				{
					if err := b.InsertString("requests"); err != nil {
						return err
					}
					if err := b.StartMapping(); err != nil {
						return err
					}
					{
						if err := b.InsertString("storage"); err != nil {
							return err
						}
						if err := b.InsertString("3Gi"); err != nil {
							return err
						}
					}
					if err := b.EndMapping(); err != nil {
						return err
					}
				}
				if err := b.EndMapping(); err != nil {
					return err
				}
			}
			if err := b.EndMapping(); err != nil {
				return err
			}
		}
		if err := b.EndMapping(); err != nil {
			return err
		}
		return nil
	}
	if err := build(b); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

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

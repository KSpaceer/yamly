package decode_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/KSpaceer/yayamls"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/decode"
	"github.com/KSpaceer/yayamls/parser"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestReader_Simple(t *testing.T) {
	type tcase struct {
		name       string
		ast        ast.Node
		calls      func(r yayamls.Decoder, vs *valueStore) error
		expected   []any
		expectDeny bool
	}

	tcases := []tcase{
		{
			name: "simple integer",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("15"),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Integer(64)
				vs.Add(v)
				return r.Error()
			},
			expected: []any{int64(15)},
		},
		{
			name: "simple integer denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("true"),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Integer(64)
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable integer",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("null"),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.Integer(64))
				}
				return r.Error()
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable integer denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewMappingNode(nil),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.Integer(64))
				}
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple hex unsigned",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("0xFF"),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Unsigned(64)
				vs.Add(v)
				return r.Error()
			},
			expected: []any{uint64(0xFF)},
		},
		{
			name: "simple unsigned denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("3.3"),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Unsigned(64)
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable unsigned",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewContentNode(
						nil,
						ast.NewTextNode(""),
					),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.Unsigned(64))
				}
				return r.Error()
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable unsigned denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("lll"),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.Unsigned(64))
				}
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple boolean",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("true"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Boolean()
				vs.Add(v)
				return r.Error()
			},
			expected: []any{true},
		},
		{
			name: "simple boolean denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("YES"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Boolean()
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable boolean",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("NULL"),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.Boolean())
				}
				return r.Error()
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable boolean denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("NULL", ast.WithQuotingType(ast.SingleQuotingType)),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.Boolean())
				}
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple float",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("33e6"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Float(64)
				vs.Add(v)
				return r.Error()
			},
			expected: []any{33e6},
		},
		{
			name: "simple float denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("33ee6"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Float(64)
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable float",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("Null"),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.Float(64))
				}
				return r.Error()
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable float denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewMappingNode(nil),
				}),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.Float(64))
				}
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple string",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("Null", ast.WithQuotingType(ast.DoubleQuotingType)),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.String()
				vs.Add(v)
				return r.Error()
			},
			expected: []any{"Null"},
		},
		{
			name: "simple string denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewSequenceNode([]ast.Node{
						ast.NewTextNode("Null", ast.WithQuotingType(ast.DoubleQuotingType)),
					}),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.String()
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable string denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewSequenceNode(nil),
					ast.NewTextNode("~"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					vs.Add(r.String())
				}
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple timestamp",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("2023-08-23"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Timestamp()
				vs.Add(v)
				return r.Error()
			},
			expected: []any{
				time.Date(2023, 8, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "simple timestamp denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("sss"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				v := r.Timestamp()
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple sequence",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewSequenceNode(nil),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				state := r.Sequence()
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return r.Error()
			},
			expected: []any{false, 0},
		},
		{
			name: "simple sequence denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewMappingNode(nil),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				state := r.Sequence()
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable sequence",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewNullNode(),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					state := r.Sequence()
					vs.Add(state.HasUnprocessedItems())
					vs.Add(state.Size())
				}
				return r.Error()
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable sequence denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("a"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					state := r.Sequence()
					vs.Add(state.HasUnprocessedItems())
					vs.Add(state.Size())
				}
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple mapping",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewMappingNode(nil),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				state := r.Mapping()
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return r.Error()
			},
			expected: []any{false, 0},
		},
		{
			name: "simple mapping denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewSequenceNode(nil),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				state := r.Mapping()
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable mapping",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewNullNode(),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
					vs.Add(nil)
				} else {
					state := r.Mapping()
					vs.Add(state.HasUnprocessedItems())
					vs.Add(state.Size())
				}
				return r.Error()
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable mapping denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("text"),
				},
			),
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				if r.TryNull() {
				} else {
					state := r.Mapping()
					vs.Add(state.HasUnprocessedItems())
					vs.Add(state.Size())
					vs.Add(nil)
				}
				return r.Error()
			},
			expectDeny: true,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			r := decode.NewASTReader(tc.ast)
			vs := valueStore{}
			err := tc.calls(r, &vs)
			if err != nil {
				switch {
				case tc.expectDeny && errors.Is(err, yayamls.Denied):
					return
				default:
					t.Fatalf("unexpected error: %v", err)
				}
			}
			got := vs.Values()
			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("values are not equal:\n\nexpected: %v\n\ngot: %v", tc.expected, got)
			}
		})

	}
}

func TestReader_Complex(t *testing.T) {
	type tcase struct {
		name       string
		src        string
		calls      func(r yayamls.Decoder, vs *valueStore) error
		expected   []any
		expectDeny bool
		expectEOS  bool
	}

	tcases := []tcase{
		{
			name: "mapping with one pair",
			src:  "key: value",
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				mapState := r.Mapping()
				for mapState.HasUnprocessedItems() {
					key := r.String()
					value := r.String()
					vs.Add(key)
					vs.Add(value)
				}
				return r.Error()
			},
			expected: []any{"key", "value"},
		},
		{
			name: "sequence with two entries",
			src:  "['val1', \"val2\"]",
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				seqState := r.Sequence()
				for seqState.HasUnprocessedItems() {
					val := r.String()
					vs.Add(val)
				}
				return r.Error()
			},
			expected: []any{"val1", "val2"},
		},
		{
			name: "struct-like mapping",
			src:  "name: \"name\"\nscore: 250\nsubscription: true\nnested: {\"inner\": .inf, \"seq\": null}",
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				mapState := r.Mapping()
				for mapState.HasUnprocessedItems() {
					key := r.String()
					switch key {
					case "name":
						value := r.String()
						vs.Add(value)
					case "score":
						if r.TryNull() {
							vs.Add(nil)
						} else {
							vs.Add(r.Unsigned(64))
						}
					case "subscription":
						value := r.Boolean()
						vs.Add(value)
					case "nested":
						nestedMapState := r.Mapping()
						for nestedMapState.HasUnprocessedItems() {
							key := r.String()
							switch key {
							case "inner":
								value := r.Float(64)
								vs.Add(value)
							case "seq":
								if r.TryNull() {
									vs.Add(nil)
								} else {
									nestedSeqState := r.Sequence()
									for nestedSeqState.HasUnprocessedItems() {
										value := r.String()
										vs.Add(value)
									}
								}
							default:
								return fmt.Errorf("unknown nested field %s", key)
							}
						}
					default:
						return fmt.Errorf("unknown field %s", key)
					}
				}
				return r.Error()
			},
			expected: []any{"name", uint64(250), true, math.Inf(1), nil},
		},
		{
			name: "anchor and alias",
			src:  "a: &anc value\nb: *anc",
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				mapState := r.Mapping()
				for mapState.HasUnprocessedItems() {
					_ = r.String()
					value := r.String()
					vs.Add(value)
				}
				return r.Error()
			},
			expected: []any{"value", "value"},
		},
		{
			name: "anchor and alias with any",
			src:  "a: &anc value\nb: *anc",
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				mapState := r.Mapping()
				for mapState.HasUnprocessedItems() {
					_ = r.String()
					value := r.Any()
					vs.Add(value)
				}
				return r.Error()
			},
			expected: []any{"value", "value"},
		},
		{
			name: "struct-like mapping with any",
			src:  "name: 'name'\nscore: 100\nunique: {key: value}\nenable: true",
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				mapState := r.Mapping()
				for mapState.HasUnprocessedItems() {
					key := r.String()
					switch key {
					case "name":
						value := r.String()
						vs.Add(value)
					case "score":
						if r.TryNull() {
							vs.Add(nil)
						} else {
							vs.Add(r.Unsigned(64))
						}
					case "unique":
						value := r.Any()
						vs.Add(value)
					case "enable":
						value := r.Boolean()
						vs.Add(value)
					default:
						return fmt.Errorf("unknown field %s", key)
					}
				}
				return r.Error()
			},
			expected: []any{"name", uint64(100), map[string]any{"key": "value"}, true},
		},
		{
			name: "anchor and alias with raw",
			src:  "a: &anc value\nb: *anc",
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				mapState := r.Mapping()
				_ = r.String()
				value := r.String()
				vs.Add(value)
				_ = r.String()
				value2 := r.Raw()
				vs.Add(value2)
				if mapState.HasUnprocessedItems() {
					return fmt.Errorf("map still has unprocessed items")
				}
				return r.Error()
			},
			expected: []any{"value", []byte("value")},
		},
		{
			name: "anchor and alias with raw (reversed)",
			src:  "a: &anc value\nb: *anc",
			calls: func(r yayamls.Decoder, vs *valueStore) error {
				mapState := r.Mapping()
				_ = r.String()
				value := r.Raw()
				vs.Add(value)
				_ = r.String()
				value2 := r.String()
				vs.Add(value2)
				if mapState.HasUnprocessedItems() {
					return fmt.Errorf("map still has unprocessed items")
				}
				return r.Error()
			},
			expected: []any{[]byte("value"), "value"},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseString(tc.src)
			if err != nil {
				t.Fatalf("parsing failed: %v", err)
			}
			r := decode.NewASTReader(tree)
			vs := valueStore{}
			if err = tc.calls(r, &vs); err != nil {
				switch {
				case tc.expectDeny && errors.Is(err, yayamls.Denied):
					return
				default:
					t.Fatalf("unexpected error: %v", err)
				}
			}
			got := vs.Values()
			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("values are not equal:\n\nexpected: %v\n\ngot: %v", tc.expected, got)
			}
		})
	}
}

func TestReader_SequencesWithNulls(t *testing.T) {
	type tcase struct {
		name       string
		src        string
		methodName string
		args       []any
		expected   []any
	}

	tcases := []tcase{
		{
			name:       "integer nullables",
			src:        "[1, -2, null, 3]",
			methodName: "Integer",
			args:       []any{64},
			expected:   []any{int64(1), int64(-2), nil, int64(3)},
		},
		{
			name:       "unsigned nullables",
			src:        "[0o777, 0xEEEE, 2, null]",
			methodName: "Unsigned",
			args:       []any{64},
			expected:   []any{uint64(0o777), uint64(0xEEEE), uint64(2), nil},
		},
		{
			name:       "boolean nullables",
			src:        "[true, ~, null, false]",
			methodName: "Boolean",
			expected:   []any{true, nil, nil, false},
		},
		{
			name:       "string nullables",
			src:        "[plain, 'single', null, \"double\", NULL]",
			methodName: "String",
			expected:   []any{"plain", "single", nil, "double", nil},
		},
		{
			name:       "float nullables",
			src:        "[-.INF, 3e18, .223, 25, NULL, Null]",
			methodName: "Float",
			args:       []any{64},
			expected:   []any{math.Inf(-1), 3e18, .223, float64(25), nil, nil},
		},
		{
			name:       "timestamp nullables",
			src:        `["2023-08-20T08:24:02Z", "2008-01-02", null]`,
			methodName: "Timestamp",
			expected: []any{
				time.Date(2023, 8, 20, 8, 24, 2, 0, time.UTC),
				time.Date(2008, 1, 2, 0, 0, 0, 0, time.UTC),
				nil,
			},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseString(tc.src)
			if err != nil {
				t.Fatalf("parser failed: %v", err)
			}
			r := decode.NewASTReader(tree)
			var values []any
			methodVal := reflect.ValueOf(r).MethodByName(tc.methodName)
			args := make([]reflect.Value, len(tc.args))
			for i := range args {
				args[i] = reflect.ValueOf(tc.args[i])
			}
			err = func() error {
				seqState := r.Sequence()
				for seqState.HasUnprocessedItems() {
					if r.TryNull() {
						values = append(values, nil)
					} else {
						results := methodVal.Call(args)
						values = append(values, results[0].Interface())
					}
				}
				return r.Error()
			}()
			if err != nil {
				t.Fatalf("failed to read: %v", err)
			}
			if len(tc.expected) != len(values) {
				t.Fatalf("values are not equal:\n\nexpected: %v\n\ngot: %v", tc.expected, values)
			}
			for i := range tc.expected {
				equal := true

				if tc.expected[i] != nil && values[i] != nil {
					v1, v2 := reflect.ValueOf(tc.expected[i]), reflect.ValueOf(values[i])
					if !v1.CanConvert(v2.Type()) {
						t.Errorf("can't case value %v to type of value %v", v1, v2)
					} else {
						v1 = v1.Convert(v2.Type())
						equal = v1.Equal(v2)
					}
				} else if tc.expected[i] != nil || values[i] != nil {
					equal = false
				}

				if !equal {
					t.Errorf("values at index %d are not equal:\nexpected: %v\ngot: %v",
						i, tc.expected[i], values[i])
				}
			}
		})
	}
}

func TestReader_ExpectAny(t *testing.T) {
	type tcase struct {
		name     string
		src      string
		expected any
	}

	tcases := []tcase{
		{
			name:     "string",
			src:      "'null'",
			expected: "null",
		},
		{
			name:     "null",
			src:      "null",
			expected: nil,
		},
		{
			name:     "unsigned",
			src:      "255",
			expected: uint64(255),
		},
		{
			name:     "integer",
			src:      "-255",
			expected: int64(-255),
		},
		{
			name:     "timestamp",
			src:      "2023-08-26",
			expected: time.Date(2023, 8, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "float",
			src:      "2e-6",
			expected: 2e-6,
		},
		{
			name:     "boolean",
			src:      "TRUE",
			expected: true,
		},
		{
			name: "map",
			src: `
                 string: 'null'
                 null: null
                 unsigned: 255
                 integer: -255
                 timestamp: 2023-08-26
                 float: 2e-6
                 boolean: true`,
			expected: map[string]any{
				"string":    "null",
				"null":      nil,
				"unsigned":  uint64(255),
				"integer":   int64(-255),
				"timestamp": time.Date(2023, 8, 26, 0, 0, 0, 0, time.UTC),
				"float":     2e-6,
				"boolean":   true,
			},
		},
		{
			name: "sequence",
			src:  `['null', null, 255, -255, 2023-08-26, 2e-6, true]`,
			expected: []any{
				"null",
				nil,
				uint64(255),
				int64(-255),
				time.Date(2023, 8, 26, 0, 0, 0, 0, time.UTC),
				2e-6,
				true,
			},
		},
		{
			name: "anchor and alias",
			src: `
              anchored: &ref value
              alias: *ref
              *ref: "another value"`,
			expected: map[string]any{
				"anchored": "value",
				"alias":    "value",
				"value":    "another value",
			},
		},
		{
			name: "merge key",
			src: `
              default: &default
                first: 15
                second: false
              first:
                second: true
                <<: *default
              second:
                first: 22
                <<: *default`,
			expected: map[string]any{
				"default": map[string]any{
					"first":  uint64(15),
					"second": false,
				},
				"first": map[string]any{
					"first":  uint64(15),
					"second": true,
				},
				"second": map[string]any{
					"first":  uint64(22),
					"second": false,
				},
			},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseString(tc.src)
			if err != nil {
				t.Fatalf("parser failed: %v", err)
			}
			r := decode.NewASTReader(tree)
			result := r.Any()
			if err = r.Error(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(tc.expected, result) {
				t.Errorf("values are not equal:\nexpected: %v\n\ngot: %v", tc.expected, result)
			}
		})
	}
}

func TestReader_ExpectRaw(t *testing.T) {
	type tcase struct {
		name     string
		src      string
		expected []byte
	}

	tcases := []tcase{
		{
			name:     "simple value",
			src:      "'22'",
			expected: []byte("'22'"),
		},
		{
			name:     "simple mapping",
			src:      "key: value",
			expected: []byte("key: value\n"),
		},
		{
			name:     "simple sequence",
			src:      "[1, 2, 3]",
			expected: []byte("- 1\n- 2\n- 3\n"),
		},
		{
			name:     "sequence of mappings",
			src:      "[{1: 2}, {3: 4}, {5: 6}]",
			expected: []byte("- 1: 2\n- 3: 4\n- 5: 6\n"),
		},
		{
			name:     "mapping of sequences",
			src:      "[1, 2, 3]: [4, 5, 6]",
			expected: []byte("? - 1\n  - 2\n  - 3\n:\n  - 4\n  - 5\n  - 6\n"),
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseString(tc.src)
			if err != nil {
				t.Fatalf("parser failed: %v", err)
			}
			if stream, ok := tree.(*ast.StreamNode); ok && len(stream.Documents()) == 1 {
				tree = stream.Documents()[0]
			}
			r := decode.NewASTReader(tree)
			result := r.Raw()
			if err = r.Error(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !bytes.Equal(tc.expected, result) {
				t.Errorf("values are not equal:\nexpected: %s\n\ngot: %s", tc.expected, result)
			}
		})
	}
}

func TestReader_K8SManifest(t *testing.T) {
	const pvcManifest = `
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
`

	type (
		requests struct {
			storage string
		}

		resources struct {
			requests requests
		}

		spec struct {
			storageClassName string
			accessModes      []string
			resources        resources
		}

		metadata struct {
			name string
		}

		manifest struct {
			apiVersion string
			kind       string
			metadata   metadata
			spec       spec
		}
	)

	expected := manifest{
		apiVersion: "v1",
		kind:       "PersistentVolumeClaim",
		metadata: metadata{
			name: "pvc-claim",
		},
		spec: spec{
			storageClassName: "manual",
			accessModes:      []string{"ReadWriteOnce"},
			resources: resources{
				requests: requests{
					storage: "3Gi",
				},
			},
		},
	}

	var got manifest

	tree, err := parser.ParseString(pvcManifest)
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	r := decode.NewASTReader(tree)
	var read func(r yayamls.Decoder) error
	read = func(r yayamls.Decoder) error {
		manifestState := r.Mapping()
		for manifestState.HasUnprocessedItems() {
			key := r.String()
			switch key {
			case "apiVersion":
				got.apiVersion = r.String()
			case "kind":
				got.kind = r.String()
			case "metadata":
				metadataState := r.Mapping()
				for metadataState.HasUnprocessedItems() {
					key = r.String()
					switch key {
					case "name":
						got.metadata.name = r.String()
					default:
						return fmt.Errorf("unknown key %s", key)

					}
				}
			case "spec":
				specState := r.Mapping()
				for specState.HasUnprocessedItems() {
					key = r.String()
					switch key {
					case "storageClassName":
						got.spec.storageClassName = r.String()
					case "accessModes":
						accessModesState := r.Sequence()
						for accessModesState.HasUnprocessedItems() {
							value := r.String()
							got.spec.accessModes = append(got.spec.accessModes, value)
						}
					case "resources":
						resourcesState := r.Mapping()
						for resourcesState.HasUnprocessedItems() {
							key = r.String()
							switch key {
							case "requests":
								requestsState := r.Mapping()
								for requestsState.HasUnprocessedItems() {
									key = r.String()
									switch key {
									case "storage":
										got.spec.resources.requests.storage = r.String()
									default:
										return fmt.Errorf("unknown key %s", key)
									}
								}
							default:
								return fmt.Errorf("unknown key %s", key)
							}
						}
					default:
						return fmt.Errorf("unknown key %s", key)
					}
				}
			default:
				return fmt.Errorf("unknown key %s", key)
			}
		}
		return r.Error()
	}

	if err = read(r); err != nil {
		t.Fatalf("failed to read from YAML: %v", err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("expected: %v\n\n\ngot: %v", expected, got)
	}
}

type valueStore []any

func (vs *valueStore) Add(v any) {
	*vs = append(*vs, v)
}

func (vs *valueStore) Values() []any {
	return *vs
}

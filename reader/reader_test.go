package reader_test

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/parser"
	"github.com/KSpaceer/yayamls/reader"
	"math"
	"reflect"
	"testing"
)

func TestReader_Simple(t *testing.T) {
	type tcase struct {
		name       string
		ast        ast.Node
		calls      func(r reader.Reader, vs *valueStore) error
		expected   []any
		expectDeny bool
	}

	tcases := []tcase{
		{
			name: "single integer",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("15"),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectInteger()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
			},
			expected: []any{int64(15)},
		},
		{
			name: "simple integer denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("true"),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectInteger()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
			},
			expectDeny: true,
		},
		{
			name: "simple nullable integer",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("null"),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableInteger()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable integer denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewMappingNode(nil),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableInteger()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
			},
			expectDeny: true,
		},
		{
			name: "simple hex unsigned",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("0xFF"),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectUnsigned()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
			},
			expected: []any{uint64(0xFF)},
		},
		{
			name: "simple unsigned denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("3.3"),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectUnsigned()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableUnsigned()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable unsigned denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("lll"),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableUnsigned()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectBoolean()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectBoolean()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
			},
			expectDeny: true,
		},
		{
			name: "simple nullable boolean",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("NULL"),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableBoolean()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable boolean denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("NULL", ast.WithQuotingType(ast.SingleQuotingType)),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableBoolean()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectFloat()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectFloat()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
			},
			expectDeny: true,
		},
		{
			name: "simple nullable float",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewTextNode("Null"),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableFloat()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
			},
			expected: []any{nil},
		},
		{
			name: "simple nullable float denied",
			ast: ast.NewStreamNode(
				[]ast.Node{
					ast.NewMappingNode(nil),
				}),
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableFloat()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectString()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				v, err := r.ExpectString()
				if err != nil {
					return err
				}
				vs.Add(v)
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				v, notNull, err := r.ExpectNullableString()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(v)
				} else {
					vs.Add(nil)
				}
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				state, err := r.ExpectSequence()
				if err != nil {
					return err
				}
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				state, err := r.ExpectSequence()
				if err != nil {
					return err
				}
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				state, notNull, err := r.ExpectNullableSequence()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(state.HasUnprocessedItems())
					vs.Add(state.Size())
				} else {
					vs.Add(nil)
				}
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				state, notNull, err := r.ExpectNullableSequence()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(state.HasUnprocessedItems())
					vs.Add(state.Size())
				} else {
					vs.Add(nil)
				}
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				state, err := r.ExpectMapping()
				if err != nil {
					return err
				}
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				state, err := r.ExpectMapping()
				if err != nil {
					return err
				}
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				state, notNull, err := r.ExpectNullableMapping()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(state.HasUnprocessedItems())
					vs.Add(state.Size())
				} else {
					vs.Add(nil)
				}
				return nil
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
			calls: func(r reader.Reader, vs *valueStore) error {
				state, notNull, err := r.ExpectNullableMapping()
				if err != nil {
					return err
				}
				if notNull {
					vs.Add(state.HasUnprocessedItems())
					vs.Add(state.Size())
				} else {
					vs.Add(nil)
				}
				return nil
			},
			expectDeny: true,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			r := reader.NewReader(tc.ast)
			vs := valueStore{}
			err := tc.calls(r, &vs)
			if err != nil {
				switch {
				case tc.expectDeny && errors.Is(err, &reader.DenyError{}):
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
		calls      func(r reader.Reader, vs *valueStore) error
		expected   []any
		expectDeny bool
		expectEOS  bool
	}

	tcases := []tcase{
		{
			name: "mapping with one pair",
			src:  "key: value",
			calls: func(r reader.Reader, vs *valueStore) error {
				mapState, err := r.ExpectMapping()
				if err != nil {
					return err
				}
				for mapState.HasUnprocessedItems() {
					key, err := r.ExpectString()
					if err != nil {
						return err
					}
					value, err := r.ExpectString()
					if err != nil {
						return err
					}
					vs.Add(key)
					vs.Add(value)
				}
				return nil
			},
			expected: []any{"key", "value"},
		},
		{
			name: "sequence with two entries",
			src:  "['val1', \"val2\"]",
			calls: func(r reader.Reader, vs *valueStore) error {
				seqState, err := r.ExpectSequence()
				if err != nil {
					return err
				}
				for seqState.HasUnprocessedItems() {
					val, err := r.ExpectString()
					if err != nil {
						return err
					}
					vs.Add(val)
				}
				return nil
			},
			expected: []any{"val1", "val2"},
		},
		{
			name: "struct-like mapping",
			src:  "name: \"name\"\nscore: 250\nsubscription: true\nnested: {\"inner\": .inf, \"seq\": null}",
			calls: func(r reader.Reader, vs *valueStore) error {
				mapState, err := r.ExpectMapping()
				if err != nil {
					return err
				}
				for mapState.HasUnprocessedItems() {
					key, err := r.ExpectString()
					if err != nil {
						return err
					}
					switch key {
					case "name":
						value, err := r.ExpectString()
						if err != nil {
							return err
						}
						vs.Add(value)
					case "score":
						value, notNull, err := r.ExpectNullableUnsigned()
						if err != nil {
							return err
						}
						if notNull {
							vs.Add(value)
						} else {
							vs.Add(nil)
						}
					case "subscription":
						value, err := r.ExpectBoolean()
						if err != nil {
							return err
						}
						vs.Add(value)
					case "nested":
						nestedMapState, err := r.ExpectMapping()
						if err != nil {
							return err
						}
						for nestedMapState.HasUnprocessedItems() {
							key, err := r.ExpectString()
							if err != nil {
								return err
							}
							switch key {
							case "inner":
								value, err := r.ExpectFloat()
								if err != nil {
									return err
								}
								vs.Add(value)
							case "seq":
								nestedSeqState, notNull, err := r.ExpectNullableSequence()
								if err != nil {
									return err
								}
								for notNull && nestedSeqState.HasUnprocessedItems() {
									value, err := r.ExpectString()
									if err != nil {
										return err
									}
									vs.Add(value)
								}
								if !notNull {
									vs.Add(nil)
								}
							default:
								return fmt.Errorf("unknown nested field %s", key)
							}
						}
					default:
						return fmt.Errorf("unknown field %s", key)
					}
				}
				return nil
			},
			expected: []any{"name", uint64(250), true, math.Inf(1), nil},
		},
		{
			name: "anchor and alias",
			src:  "a: &anc value\nb: *anc",
			calls: func(r reader.Reader, vs *valueStore) error {
				mapState, err := r.ExpectMapping()
				if err != nil {
					return err
				}
				for mapState.HasUnprocessedItems() {
					_, err := r.ExpectString()
					if err != nil {
						return err
					}
					value, err := r.ExpectString()
					if err != nil {
						return err
					}
					vs.Add(value)
				}
				return nil
			},
			expected: []any{"value", "value"},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseString(tc.src)
			if err != nil {
				t.Fatalf("parsing failed: %v", err)
			}
			r := reader.NewReader(tree)
			vs := valueStore{}
			if err = tc.calls(r, &vs); err != nil {
				switch {
				case tc.expectDeny && errors.Is(err, &reader.DenyError{}):
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

func TestReader_SequencesOfNullables(t *testing.T) {
	type tcase struct {
		name       string
		src        string
		methodName string
		expected   []any
	}

	tcases := []tcase{
		{
			name:       "integer nullables",
			src:        "[1, -2, null, 3]",
			methodName: "ExpectNullableInteger",
			expected:   []any{int64(1), int64(-2), nil, int64(3)},
		},
		{
			name:       "unsigned nullables",
			src:        "[0o777, 0xEEEE, 2, null]",
			methodName: "ExpectNullableUnsigned",
			expected:   []any{uint64(0o777), uint64(0xEEEE), uint64(2), nil},
		},
		{
			name:       "boolean nullables",
			src:        "[true, ~, null, false]",
			methodName: "ExpectNullableBoolean",
			expected:   []any{true, nil, nil, false},
		},
		{
			name:       "string nullables",
			src:        "[plain, 'single', null, \"double\", NULL]",
			methodName: "ExpectNullableString",
			expected:   []any{"plain", "single", nil, "double", nil},
		},
		{
			name:       "float nullables",
			src:        "[-.INF, 3e18, .223, 25, NULL, Null]",
			methodName: "ExpectNullableFloat",
			expected:   []any{math.Inf(-1), 3e18, .223, float64(25), nil, nil},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseString(tc.src)
			if err != nil {
				t.Fatalf("parser failed: %v", err)
			}
			r := reader.NewReader(tree)
			var values []any
			methodVal := reflect.ValueOf(r).MethodByName(tc.methodName)
			err = func() error {
				seqState, err := r.ExpectSequence()
				if err != nil {
					return err
				}
				for seqState.HasUnprocessedItems() {
					results := methodVal.Call(nil)
					err := results[2].Interface()
					if err != nil {
						return err.(error)
					}
					if results[1].Interface().(bool) {
						values = append(values, results[0].Interface())
					} else {
						values = append(values, nil)
					}
				}
				return nil
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

func TestReader_K8SManifest(t *testing.T) {
	const pvcManifest = `
    apiVersion: v1
    kind: Persistent
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
		kind:       "Persistent",
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
	r := reader.NewReader(tree)
	var read func(r reader.Reader) error
	read = func(r reader.Reader) error {
		manifestState, err := r.ExpectMapping()
		if err != nil {
			return err
		}
		for manifestState.HasUnprocessedItems() {
			key, err := r.ExpectString()
			if err != nil {
				return err
			}
			switch key {
			case "apiVersion":
				got.apiVersion, err = r.ExpectString()
				if err != nil {
					return err
				}
			case "kind":
				got.kind, err = r.ExpectString()
				if err != nil {
					return err
				}
			case "metadata":
				metadataState, err := r.ExpectMapping()
				if err != nil {
					return err
				}
				for metadataState.HasUnprocessedItems() {
					key, err = r.ExpectString()
					if err != nil {
						return err
					}
					switch key {
					case "name":
						got.metadata.name, err = r.ExpectString()
						if err != nil {
							return err
						}
					default:
						return fmt.Errorf("unknown key %s", key)

					}
				}
			case "spec":
				specState, err := r.ExpectMapping()
				if err != nil {
					return err
				}
				for specState.HasUnprocessedItems() {
					key, err = r.ExpectString()
					if err != nil {
						return err
					}
					switch key {
					case "storageClassName":
						got.spec.storageClassName, err = r.ExpectString()
						if err != nil {
							return err
						}
					case "accessModes":
						accessModesState, err := r.ExpectSequence()
						if err != nil {
							return err
						}
						for accessModesState.HasUnprocessedItems() {
							value, err := r.ExpectString()
							if err != nil {
								return err
							}
							got.spec.accessModes = append(got.spec.accessModes, value)
						}
					case "resources":
						resourcesState, err := r.ExpectMapping()
						if err != nil {
							return err
						}
						for resourcesState.HasUnprocessedItems() {
							key, err = r.ExpectString()
							if err != nil {
								return err
							}
							switch key {
							case "requests":
								requestsState, err := r.ExpectMapping()
								if err != nil {
									return err
								}
								for requestsState.HasUnprocessedItems() {
									key, err = r.ExpectString()
									if err != nil {
										return err
									}
									switch key {
									case "storage":
										got.spec.resources.requests.storage, err = r.ExpectString()
										if err != nil {
											return err
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
			default:
				return fmt.Errorf("unknown key %s", key)
			}
		}
		return nil
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

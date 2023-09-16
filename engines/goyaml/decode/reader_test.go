package decode_test

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yamly"
	"github.com/KSpaceer/yamly/engines/goyaml/decode"
	"gopkg.in/yaml.v3"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestReader_Simple(t *testing.T) {
	type tcase struct {
		name       string
		src        []byte
		calls      func(r yamly.Decoder, vs *valueStore) error
		expected   []any
		expectDeny bool
	}

	tcases := []tcase{
		{
			name: "simple integer",
			src:  []byte("15"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Integer(64)
				vs.Add(v)
				return r.Error()
			},
			expected: []any{int64(15)},
		},
		{
			name: "simple integer denied",
			src:  []byte("true"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Integer(64)
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable integer",
			src:  []byte("null"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("{}"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("0xFF"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Unsigned(64)
				vs.Add(v)
				return r.Error()
			},
			expected: []any{uint64(0xFF)},
		},
		{
			name: "simple unsigned denied",
			src:  []byte("3.3"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Unsigned(64)
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable unsigned",
			src:  []byte("~"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("lll"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("true"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Boolean()
				vs.Add(v)
				return r.Error()
			},
			expected: []any{true},
		},
		{
			name: "simple boolean denied",
			src:  []byte("YES"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Boolean()
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable boolean",
			src:  []byte("NULL"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("'NULL'"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("33e6"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Float(64)
				vs.Add(v)
				return r.Error()
			},
			expected: []any{33e6},
		},
		{
			name: "simple float denied",
			src:  []byte("33ee6"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Float(64)
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable float",
			src:  []byte("Null"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("[]"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte(`"Null"`),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.String()
				vs.Add(v)
				return r.Error()
			},
			expected: []any{"Null"},
		},
		{
			name: "simple string denied",
			src:  []byte(`- "Null"`),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.String()
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple nullable string denied",
			src:  []byte("[~]"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("2023-08-23"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  []byte("sss"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				v := r.Timestamp()
				vs.Add(v)
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple sequence",
			src:  []byte("[]"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				state := r.Sequence()
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return r.Error()
			},
			expected: []any{false, 0},
		},
		{
			name: "simple sequence denied",
			src:  []byte("{}"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				state := r.Sequence()
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return r.Error()
			},
			expectDeny: true,
		},
		{
			name: "simple mapping",
			src:  []byte("{}"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				state := r.Mapping()
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return r.Error()
			},
			expected: []any{false, 0},
		},
		{
			name: "simple mapping denied",
			src:  []byte("[]"),
			calls: func(r yamly.Decoder, vs *valueStore) error {
				state := r.Mapping()
				vs.Add(state.HasUnprocessedItems())
				vs.Add(state.Size())
				return r.Error()
			},
			expectDeny: true,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			var tree yaml.Node
			if err := yaml.Unmarshal(tc.src, &tree); err != nil {
				t.Fatalf("parsing failed: %v", err)
			}
			r := decode.NewASTReader(&tree)
			vs := valueStore{}
			err := tc.calls(r, &vs)
			if err != nil {
				switch {
				case tc.expectDeny && errors.Is(err, yamly.Denied):
					return
				default:
					t.Errorf("unexpected error: %v", err)
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
		calls      func(r yamly.Decoder, vs *valueStore) error
		expected   []any
		expectDeny bool
		expectEOS  bool
	}

	tcases := []tcase{
		{
			name: "mapping with one pair",
			src:  "key: value",
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
			src:  "name: 'name'\nscore: 100\nskip_me: null\nunique: {key: value}\nenable: true\n",
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
					default: // skip_me
						r.Skip()
					}
				}
				return r.Error()
			},
			expected: []any{"name", uint64(100), map[string]any{"key": "value"}, true},
		},
		{
			name: "anchor and alias with raw",
			src:  "a: &anc value\nb: *anc",
			calls: func(r yamly.Decoder, vs *valueStore) error {
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
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			var tree yaml.Node
			if err := yaml.Unmarshal([]byte(tc.src), &tree); err != nil {
				t.Fatalf("parsing failed: %v", err)
			}
			r := decode.NewASTReader(&tree)
			vs := valueStore{}
			if err := tc.calls(r, &vs); err != nil {
				switch {
				case tc.expectDeny && errors.Is(err, yamly.Denied):
					return
				default:
					t.Errorf("unexpected error: %v", err)
				}
			}
			got := vs.Values()
			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("values are not equal:\n\nexpected: %v\n\ngot: %v", tc.expected, got)
			}
		})
	}
}

type valueStore []any

func (vs *valueStore) Add(v any) {
	*vs = append(*vs, v)
}

func (vs *valueStore) Values() []any {
	return *vs
}

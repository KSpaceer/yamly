package encode_test

import (
	"encoding/json"
	"github.com/KSpaceer/yamly"
	"github.com/KSpaceer/yamly/engines/goyaml/encode"
	"gopkg.in/yaml.v3"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestBuilder_Simple(t *testing.T) {
	type tcase struct {
		name      string
		calls     func(b yamly.TreeBuilder[*yaml.Node])
		expected  any // comparing with marshal
		expectErr bool
	}

	tcases := []tcase{
		{
			name: "simple integer",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.InsertInteger(15)
			},
			expected: int64(15),
		},
		{
			name: "simple unsigned",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.InsertUnsigned(0xFF)
			},
			expected: uint(0xFF),
		},
		{
			name: "simple boolean",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.InsertBoolean(false)
			},
			expected: false,
		},
		{
			name: "simple float",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.InsertFloat(-3.3e3)
			},
			expected: -3.3e3,
		},
		{
			name: "simple string",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.InsertString("null")
			},
			expected: "null",
		},
		{
			name: "simple timestamp",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.InsertTimestamp(
					time.Date(2023, 8, 27, 21, 42, 0, 0, time.UTC),
				)
			},
			expected: time.Date(2023, 8, 27, 21, 42, 0, 0, time.UTC),
		},
		{
			name: "null",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.InsertNull()
			},
			expected: (*bool)(nil), // typed nil for reflect
		},
		{
			name: "simple sequence",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.StartSequence()
				b.InsertInteger(22)
				b.EndSequence()
			},
			expected: []int{22},
		},
		{
			name: "simple mapping",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.StartMapping()
				b.InsertString("key")
				b.InsertString("value")
				b.EndMapping()
			},
			expected: map[string]string{"key": "value"},
		},
		{
			name: "ending complex node without starting",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.EndMapping()
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
			gotValuePtr := reflect.New(reflect.TypeOf(tc.expected))
			if err = result.Decode(gotValuePtr.Interface()); err != nil {
				t.Errorf("failed to decode tree: %v", err)
			}
			if !reflect.DeepEqual(tc.expected, gotValuePtr.Elem().Interface()) {
				t.Errorf(
					"values are not equal:\nexpected: %[1]v with type %[1]T\ngot: %[2]v with type %[2]T",
					tc.expected,
					gotValuePtr.Elem().Interface(),
				)
			}
		})
	}
}

func TestBuilder_Complex(t *testing.T) {
	type tcase struct {
		name      string
		calls     func(b yamly.TreeBuilder[*yaml.Node])
		expected  any // comparing with marshal
		expectErr bool
	}

	tcases := []tcase{
		{
			name: "struct-like mapping",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
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
			expected: struct {
				Name         string `yaml:"name"`
				Score        uint   `yaml:"score"`
				Subscription bool   `yaml:"subscription"`
				Nested       struct {
					Inner float64  `yaml:"inner"`
					Seq   []string `yaml:"seq"`
				} `yaml:"nested"`
			}{
				Name:         "name",
				Score:        250,
				Subscription: true,
				Nested: struct {
					Inner float64  `yaml:"inner"`
					Seq   []string `yaml:"seq"`
				}{
					Inner: math.Inf(1),
					Seq:   nil,
				},
			},
		},
		{
			name: "raw insertion",
			calls: func(b yamly.TreeBuilder[*yaml.Node]) {
				b.StartMapping()
				b.InsertString("key")

				b.StartSequence()
				b.InsertInteger(4)
				b.InsertInteger(5)
				b.EndSequence()

				b.InsertString("raw")
				b.InsertRaw(json.Marshal([]int64{1, 2, 3}))
				b.EndMapping()
			},
			expected: map[string][]int{
				"key": {4, 5},
				"raw": {1, 2, 3},
			},
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
			gotValuePtr := reflect.New(reflect.TypeOf(tc.expected))
			if err = result.Decode(gotValuePtr.Interface()); err != nil {
				t.Errorf("failed to decode tree: %v", err)
			}
			if !reflect.DeepEqual(tc.expected, gotValuePtr.Elem().Interface()) {
				t.Errorf(
					"values are not equal:\nexpected: %[1]v with type %[1]T\ngot: %[2]v with type %[2]T",
					tc.expected,
					gotValuePtr.Elem().Interface(),
				)
			}
		})
	}
}

func TestBuilder_InsertRaw(t *testing.T) {
	type tcase struct {
		name     string
		src      []byte
		expected any
	}

	tcases := []tcase{
		{
			name:     "simple value",
			src:      []byte("'22'"),
			expected: "22",
		},
		{
			name:     "simple mapping",
			src:      []byte("key: value"),
			expected: map[string]string{"key": "value"},
		},
		{
			name:     "simple sequence",
			src:      []byte("[1, 2, 3]"),
			expected: []uint32{1, 2, 3},
		},
		{
			name: "sequence of mappings",
			src: []byte(`
              - 1: 2
              - 3: 4
              - 5: 6
            `),
			expected: []map[int]int{
				{1: 2},
				{3: 4},
				{5: 6},
			},
		},
		{
			name: "mapping of sequences",
			src: []byte(`
            ? - 1
              - 2
              - 3
            : - 4
              - 5
              - 6
            `),
			expected: map[[3]int][3]int{
				{1, 2, 3}: {4, 5, 6},
			},
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
			mapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(tc.expected)) // map[string]type
			mapValuePtr := reflect.New(mapType)                                       // var m *map[string]type
			if err = result.Decode(mapValuePtr.Interface()); err != nil {             // result.Decode(m)
				t.Errorf("failed to decode tree: %v", err)
			}
			mapValue := mapValuePtr.Elem()                                                              // mp := *m
			if !reflect.DeepEqual(tc.expected, mapValue.MapIndex(reflect.ValueOf("raw")).Interface()) { // mp["raw"] == tc.expected
				t.Errorf(
					"values are not equal:\nexpected: %[1]v with type %[1]T\ngot: %[2]v with type %[2]T",
					tc.expected,
					mapValue.MapIndex(reflect.ValueOf("raw")).Interface(),
				)
			}
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
	type (
		Metadata struct {
			Name string `yaml:"name"`
		}
		Requests struct {
			Storage string `yaml:"storage"`
		}
		Resources struct {
			Requests Requests `yaml:"requests"`
		}
		Spec struct {
			StorageClassName string    `yaml:"storageClassName"`
			AccessModes      []string  `yaml:"accessModes"`
			Resources        Resources `yaml:"resources"`
		}
		PVCManfiest struct {
			APIVersion string   `yaml:"apiVersion"`
			Kind       string   `yaml:"kind"`
			Metadata   Metadata `yaml:"metadata"`
			Spec       Spec     `yaml:"spec"`
		}
	)

	expected := PVCManfiest{
		APIVersion: "v1",
		Kind:       "PersistentVolumeClaim",
		Metadata: Metadata{
			Name: "pvc-claim",
		},
		Spec: Spec{
			StorageClassName: "manual",
			AccessModes:      []string{"ReadWriteOnce"},
			Resources: Resources{
				Requests: Requests{
					Storage: "3Gi",
				},
			},
		},
	}

	b := encode.NewASTBuilder(encode.WithUnquotedOneLineStrings())

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

	result, err := b.Result()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var got PVCManfiest
	if err = result.Decode(&got); err != nil {
		t.Errorf("failed to decode result: %v", err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("values are not equal:\nexpected: %v\ngot: %v", expected, got)
	}
}

package test_test

import (
	"bytes"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"text/template"

	_ "gopkg.in/yaml.v3"

	_ "github.com/KSpaceer/yamly"
	_ "github.com/KSpaceer/yamly/engines/goyaml"
	_ "github.com/KSpaceer/yamly/engines/yayamls"
)

func TestGenerator_EngineGoYAML(t *testing.T) {
	t.Parallel()
	mainCode := `
package main

import (
  "fmt"
  "reflect"
  "os"
  {{ range $import := .Imports }}
  "{{ $import }}"
  {{ end }}

  "gopkg.in/yaml.v3"

  "github.com/KSpaceer/yamly/test/{{ .TmpRoot }}/{{ .PkgName }}"
)

func main() {
	var v {{ if .UsePointer -}}*{{- end -}}{{ .PkgName }}.TestType
	v = {{ .Value }}
    data, err := yaml.Marshal(v) 
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    var v2 {{ if .UsePointer -}}*{{- end -}}{{ .PkgName }}.TestType
    err = yaml.Unmarshal(data, &v2) 
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
	if reflect.DeepEqual(v, v2) {
		fmt.Print("SUCCESS")
    } else {
		fmt.Printf("start: %v\n\n\nfinish: %v", v, v2)
	}
}
`

	mainCodeTemplate := template.Must(template.New("maincode").Parse(mainCode))

	typeDefinitionCode := `
package {{ .PkgName }}

{{ if .Imports }}
import (
  {{ range $import := .Imports }}
  "{{ $import }}"
  {{ end }}
)
{{ end }}

type TestType {{ .TypeDef }}

{{ range $i, $typedef := .ExtraTypeDefs }}
type ExtraType{{ $i }} {{ $typedef }}
{{ end }}
`

	typeDefinitionCodeTemplate := template.Must(template.New("typedef").Parse(typeDefinitionCode))

	runEngineTest(t, mainCodeTemplate, typeDefinitionCodeTemplate, "goyaml")
}

func TestGenerator_EngineYAYAMLS(t *testing.T) {
	t.Parallel()
	mainCode := `
package main

import (
  "fmt"
  "reflect"
  "os"
  {{ range $import := .Imports }}
  "{{ $import }}"
  {{ end }}

  "github.com/KSpaceer/yamly/test/{{ .TmpRoot }}/{{ .PkgName }}"
)

func main() {
	var v {{ if .UsePointer -}}*{{- end -}}{{ .PkgName }}.TestType
	v = {{ .Value }}
    data, err := v.MarshalYAML()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    var v2 {{ if .UsePointer -}}*{{- end -}}{{ .PkgName }}.TestType
	{{ if .UsePointer }}
    v2 = new({{ .PkgName }}.TestType)
    {{ end }}
    err = v2.UnmarshalYAML(data)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
	if reflect.DeepEqual(v, v2) {
		fmt.Print("SUCCESS")
    } else {
		fmt.Printf("start: %v\n\n\nfinish: %v", v, v2)
	}
}
`

	mainCodeTemplate := template.Must(template.New("maincode").Parse(mainCode))

	typeDefinitionCode := `
package {{ .PkgName }}

{{ if .Imports }}
import (
  {{ range $import := .Imports }}
  "{{ $import }}"
  {{ end }}
)
{{ end }}

type TestType {{ .TypeDef }}

{{ range $i, $typedef := .ExtraTypeDefs }}
type ExtraType{{ $i }} {{ $typedef }}
{{ end }}
`

	typeDefinitionCodeTemplate := template.Must(template.New("typedef").Parse(typeDefinitionCode))

	runEngineTest(t, mainCodeTemplate, typeDefinitionCodeTemplate, "yayamls")
}

func runEngineTest(
	t *testing.T,
	mainCodeTemplate, typeDefinitionTemplate *template.Template,
	engine string,
) {
	t.Helper()
	type tcase struct {
		name string

		flags []string

		TmpRoot string

		Imports    []string
		PkgName    string
		TypeDef    string
		Value      string
		UsePointer bool

		ExtraTypeDefs []string
	}

	tcases := []tcase{
		{
			name:    "basic",
			PkgName: "test",
			TypeDef: "int",
			Value:   "15",
		},
		{
			name:    "string",
			PkgName: "test",
			TypeDef: "string",
			Value:   `"hello world"`,
		},
		{
			name:    "[]byte",
			PkgName: "testos",
			TypeDef: "json.RawMessage",
			Value:   `testos.TestType(json.RawMessage("rawmessage"))`,
			Imports: []string{"encoding/json"},
		},
		{
			name:    "slice",
			PkgName: "slicetest",
			TypeDef: "[]uint32",
			Value:   "slicetest.TestType{1, 3, 5, 6}",
		},
		{
			name:    "array",
			PkgName: "arraytest",
			TypeDef: "[5]int",
			Value:   "arraytest.TestType{2: 1, 4: -22}",
		},
		{
			name:    "array of slices",
			PkgName: "slicearray",
			TypeDef: "[3][]bool",
			Value:   "slicearray.TestType{[]bool{true, false}, []bool{false, true, false}}",
		},
		{
			name:    "map",
			PkgName: "maptest",
			TypeDef: "map[string]*int",
			Value:   `func() maptest.TestType { a, b := 1, 2; return maptest.TestType{"a": &a, "b": &b, "c": nil} }()`,
		},
		{
			name:    "map of maps",
			PkgName: "mapmap",
			TypeDef: "map[int]map[int]string",
			Value:   `mapmap.TestType{1: {2: "a", 3: "b"}, 2: {1: "c", 3: "d"}, 3: {1: "e", 2: "f"} }`,
		},
		{
			name:    "simple struct",
			PkgName: "structtest",
			TypeDef: "struct{ Integer int; Unsigned uint; String string; Boolean bool; Timestamp time.Time; }",
			Value:   `structtest.TestType{Integer: 250, Unsigned: 100000, String: "string", Timestamp: time.Now().UTC()}`,
			Imports: []string{"time"},
		},
		{
			name:    "linked list",
			PkgName: "linkedlist",
			TypeDef: "struct{ Value int; Next *TestType; }",
			Value: "linkedlist.TestType{Value: 10, Next: &linkedlist.TestType{Value: 100, Next: " +
				"&linkedlist.TestType{Value: 1000}}}",
		},
		{
			name:    "linked list (with flags)",
			flags:   []string{"--encode-pointer-receiver", "--disallow-unknown-fields", "--omitempty"},
			PkgName: "linkedlist",
			TypeDef: "struct{ Value int; Next *TestType; }",
			Value: "&linkedlist.TestType{Value: 10, Next: &linkedlist.TestType{Value: 100, Next: " +
				"&linkedlist.TestType{Value: 1000}}}",
			UsePointer: true,
		},
		{
			name:    "anonymous struct",
			PkgName: "withanon",
			TypeDef: "struct{ ID uint32; Banned bool; Info struct{ Name string; Age uint8; } }",
			Value:   "withanon.TestType{ID: 22, Info: struct{ Name string; Age uint8}{Name: \"yamly\", Age: 0} }",
		},
		{
			name:    "with tags",
			PkgName: "tagged",
			TypeDef: "struct{ Name string `yaml:\"my_name\"` ; Age int8 `yaml:\"age,omitempty\"` ; Ignored int `yaml:\"-\"` ; }",
			Value:   "tagged.TestType{Name: \"yamly\"}",
		},
		{
			name:    "inlining",
			flags:   []string{"--inline-embedded"},
			PkgName: "inlined",
			TypeDef: "struct{ Name string `yaml:\"my_name\"` ; ExtraType0 ; }",
			Value:   "inlined.TestType{Name: \"inlined-yamly\", ExtraType0: inlined.ExtraType0{Nested: \"yamly-nested\"}}",
			ExtraTypeDefs: []string{
				"struct{ Nested string `yaml:\"nested\"`; }",
			},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			root, err := os.MkdirTemp(".", "tmptest*")
			if err != nil {
				t.Fatalf("failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(root)

			root = path.Clean(root)

			if err := os.Mkdir(root+"/cmd", 0o755); err != nil {
				t.Fatalf("failed to create directory for main package: %v", err)
			}

			if err := os.Mkdir(root+"/"+tc.PkgName, 0o755); err != nil {
				t.Fatalf("failed to create directory for type package: %v", err)
			}

			tc.TmpRoot = root

			mainFile, err := os.OpenFile(root+"/cmd/main.go", os.O_CREATE|os.O_WRONLY, 0o755)
			if err != nil {
				t.Fatalf("failed to create main.go: %v", err)
			}
			defer mainFile.Close()

			if err := mainCodeTemplate.Execute(mainFile, tc); err != nil {
				t.Fatalf("failed to execute main code template: %v", err)
			}

			typeFile, err := os.OpenFile(root+"/"+tc.PkgName+"/type.go", os.O_CREATE|os.O_WRONLY, 0o755)
			if err != nil {
				t.Fatalf("failed to create type.go: %v", err)
			}
			defer typeFile.Close()

			if err := typeDefinitionTemplate.Execute(typeFile, tc); err != nil {
				t.Fatalf("failed to execute typedef code template: %v", err)
			}

			execArgs := []string{"run", "../cmd/yamlygen/main.go"}
			execArgs = append(execArgs, tc.flags...)

			execArgs = append(execArgs, "-type", "TestType", "-engine", engine)
			execArgs = append(execArgs, root+"/"+tc.PkgName)

			var stdout, stderr bytes.Buffer
			cmd := exec.Command("go", execArgs...)
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Errorf("failed to run yamlygen binary: %v\n\nStderr content: %v", err, stderr.String())
			}
			stderr.Reset()

			if data, err := os.ReadFile(root + "/" + tc.PkgName + "/test_type_yamly.go"); err != nil {
				t.Errorf("failed to read generated file: %v", err)
			} else {
				t.Logf("GENERATED DATA:\n\n\n%s\n\n\n===========", string(data))
			}

			cmd = exec.Command("go", "run", root+"/cmd/main.go") // nolint: gosec
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Errorf("failed to run test main binary: %v\n\nStderr content: %v", err, stderr.String())
			}

			result := stdout.String()

			if strings.TrimSpace(result) != "SUCCESS" {
				t.Errorf("starting and finished values are seem to be not equal.\n\nStdout: %v", result)
			}
		})
	}
}

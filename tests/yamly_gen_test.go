package tests_test

import (
	"bytes"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"text/template"
)

func TestGenerator(t *testing.T) {
	type tcase struct {
		name string

		flags []string

		TmpRoot string

		Imports []string
		PkgName string
		TypeDef string
		Value   string
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
			Value:   "linkedlist.TestType{Value: 10, Next: &linkedlist.TestType{Value: 100, Next: &linkedlist.TestType{Value: 1000}}}",
		},
		{
			name:    "linked list (with flags)",
			flags:   []string{"--encode-pointer-receiver", "--disallow-unknown-fields", "--omitempty"},
			PkgName: "linkedlist",
			TypeDef: "struct{ Value int; Next *TestType; }",
			Value:   "linkedlist.TestType{Value: 10, Next: &linkedlist.TestType{Value: 100, Next: &linkedlist.TestType{Value: 1000}}}",
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
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			root, err := os.MkdirTemp(".", "tmptest*")
			if err != nil {
				t.Fatalf("failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(root)

			root = path.Clean(root)

			if err := os.Mkdir(root+"/cmd", 0755); err != nil {
				t.Fatalf("failed to create directory for main package: %v", err)
			}

			if err := os.Mkdir(root+"/"+tc.PkgName, 0755); err != nil {
				t.Fatalf("failed to create directory for type package: %v", err)
			}

			tc.TmpRoot = root

			mainFile, err := os.OpenFile(root+"/cmd/main.go", os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				t.Fatalf("failed to create main.go: %v", err)
			}
			defer mainFile.Close()

			if err := mainCodeTemplate.Execute(mainFile, tc); err != nil {
				t.Fatalf("failed to execute main code template: %v", err)
			}

			typeFile, err := os.OpenFile(root+"/"+tc.PkgName+"/type.go", os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				t.Fatalf("failed to create type.go: %v", err)
			}
			defer typeFile.Close()

			if err := typeDefinitionCodeTemplate.Execute(typeFile, tc); err != nil {
				t.Fatalf("failed to execute typedef code template: %v", err)
			}

			execArgs := []string{"run", "../cmd/yamlygen/main.go"}
			execArgs = append(execArgs, tc.flags...)

			execArgs = append(execArgs, "-type", "TestType")
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

			cmd = exec.Command("go", "run", root+"/cmd/main.go")
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

var mainCode = `
package main

import (
  "fmt"
  "reflect"
  "os"
  {{ range $import := .Imports }}
  "{{ $import }}"
  {{ end }}

  "github.com/KSpaceer/yamly/tests/{{ .TmpRoot }}/{{ .PkgName }}"
)

func main() {
	var v {{ .PkgName }}.TestType
	v = {{ .Value }}
    data, err := v.MarshalYAML()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    var v2 {{ .PkgName }}.TestType
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

var mainCodeTemplate = template.Must(template.New("maincode").Parse(mainCode))

var typeDefinitionCode = `
package {{ .PkgName }}

{{ if .Imports }}
import (
  {{ range $import := .Imports }}
  "{{ $import }}"
  {{ end }}
)
{{ end }}

type TestType {{ .TypeDef }}
`

var typeDefinitionCodeTemplate = template.Must(template.New("typedef").Parse(typeDefinitionCode))

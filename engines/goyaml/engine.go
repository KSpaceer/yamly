// Package goyaml represents an engine for yamly based on
// gopkg.in/yaml.v3 package.
package goyaml

import (
	"fmt"
	"github.com/KSpaceer/yamly/generator"
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
	"strings"
)

const (
	pkgGoYaml = "gopkg.in/yaml.v3"
	pkgDecode = "github.com/KSpaceer/yamly/engines/goyaml/decode"
	pkgEncode = "github.com/KSpaceer/yamly/engines/goyaml/encode"
)

var Generator generator.EngineGenerator = engineGenerator{}

type engineGenerator struct{}

func (engineGenerator) Packages() map[string]string {
	return map[string]string{
		pkgGoYaml: "yaml",
		pkgDecode: "decode",
		pkgEncode: "encode",
	}
}

func (engineGenerator) WarningSuppressors() []string {
	return []string{"*encode.ASTWriter", "*decode.ASTReader", "yaml.Marshaler"}
}

func (engineGenerator) GenerateUnmarshalers(dst io.Writer, decodeFuncName string, typeName string) error {
	fmt.Fprintln(dst, "// UnmarshalYAML supports yaml.Unmarshaler interface")
	fmt.Fprintln(dst, "func (v *"+typeName+") UnmarshalYAML(value *yaml.Node) error {")
	fmt.Fprintln(dst, "  in := decode.NewASTReader(value)")
	fmt.Fprintln(dst, "  "+decodeFuncName+"(in, v)")
	fmt.Fprintln(dst, "  return in.Error()")
	fmt.Fprintln(dst, "}")
	return nil
}

func (engineGenerator) GenerateMarshalers(dst io.Writer, encodeFuncName string, typeName string) error {
	fmt.Fprintln(dst, "// MarshalYAML support yaml.Marshaler interface")
	fmt.Fprintln(dst, "func (v "+typeName+") MarshalYAML() (any, error) {")
	fmt.Fprintln(dst, "out := encode.NewASTBuilder()")
	fmt.Fprintln(dst, "  "+encodeFuncName+"(out, v)")
	fmt.Fprintln(dst, "  return out.Result()")
	fmt.Fprintln(dst, "}")
	return nil
}

func (engineGenerator) UnmarshalersImplementationCheck(
	dst io.Writer,
	t reflect.Type,
	outArg string,
	indent int,
) (generator.ImplementationResult, error) {
	unmarshalIface := reflect.TypeOf((*yaml.Unmarshaler)(nil)).Elem()
	if reflect.PtrTo(t).Implements(unmarshalIface) {
		whitespace := strings.Repeat(" ", indent)
		fmt.Fprintln(dst, whitespace+"if extIn, ok := in.(yamly.ExtendedDecoder[*yaml.Node]); ok {")
		fmt.Fprintln(dst, whitespace+"  in.AddError(("+outArg+").UnmarshalYAML(extIn.Node())")
		fmt.Fprintln(dst, whitespace, "} else {")
		return generator.ImplementationResultConditional, nil
	}
	return generator.ImplementationResultFalse, nil
}

func (engineGenerator) MarshalersImplementationCheck(
	dst io.Writer,
	t reflect.Type,
	inArg string,
	indent int,
) (generator.ImplementationResult, error) {
	marshalIface := reflect.TypeOf((*yaml.Marshaler)(nil)).Elem()
	if reflect.PtrTo(t).Implements(marshalIface) {
		fmt.Fprintln(dst, strings.Repeat(" ", indent)+"out.InsertRaw(yaml.Marshal("+inArg+"))")
		return generator.ImplementationResultTrue, nil
	}
	return generator.ImplementationResultFalse, nil
}

func (engineGenerator) GenerateUnmarshalEmptyInterfaceAssertions(dst io.Writer, outArg string, indent int) error {
	whitespace := strings.Repeat(" ", indent)
	fmt.Fprintln(dst, whitespace+"if m, ok := "+outArg+".(yaml.Unmarshaler); ok {")
	fmt.Fprintln(dst, whitespace+"  if extIn, ok := in.(yamly.ExtendedDecoder[*yaml.Node]); ok {")
	fmt.Fprintln(dst, whitespace+"    in.AddError(m.UnmarshalYAML(extIn.Node())")
	fmt.Fprintln(dst, whitespace+"  } else {")
	fmt.Fprintln(dst, whitespace+"    "+outArg+" = in.Any()")
	fmt.Fprintln(dst, whitespace+"} else {")
	fmt.Fprintln(dst, whitespace+"  "+outArg+" = in.Any()")
	fmt.Fprintln(dst, whitespace+"}")
	return nil
}

func (engineGenerator) GenerateMarshalEmptyInterfaceAssertions(dst io.Writer, inArg string, indent int) error {
	whitespace := strings.Repeat(" ", indent)
	fmt.Fprintln(dst, whitespace+"out.InsertRaw(yaml.Marshal("+inArg+"))")
	return nil
}

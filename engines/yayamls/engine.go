// Package yayamls represents an engine for yamly with
// unique YAML marshalling system with full specification coverage.
package yayamls

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/KSpaceer/yamly/generator"
)

const (
	pkgYayamls = "github.com/KSpaceer/yamly/engines/yayamls"
	pkgDecode  = "github.com/KSpaceer/yamly/engines/yayamls/decode"
	pkgEncode  = "github.com/KSpaceer/yamly/engines/yayamls/encode"
)

// Generator is used in generated code.
var Generator generator.EngineGenerator = engineGenerator{}

type engineGenerator struct{}

func (engineGenerator) Packages() map[string]string {
	return map[string]string{
		pkgYayamls: "yayamls",
		pkgDecode:  "decode",
		pkgEncode:  "encode",
	}
}

func (engineGenerator) WarningSuppressors() []string {
	return []string{"*encode.ASTWriter", "*decode.ASTReader", "yayamls.Marshaler"}
}

func (engineGenerator) GenerateUnmarshalers(dst io.Writer, decodeFuncName, typeName string) error {
	fmt.Fprintln(dst, "// UnmarshalYAML supports yayamls.Unmarshaler interface")
	fmt.Fprintln(dst, "func (v *"+typeName+") UnmarshalYAML(data []byte) error {")
	fmt.Fprintln(dst, "  in, err := decode.NewASTReaderFromBytes(data)")
	fmt.Fprintln(dst, "  if err != nil {")
	fmt.Fprintln(dst, "    return err")
	fmt.Fprintln(dst, "  }")
	fmt.Fprintln(dst, "  "+decodeFuncName+"(in, v)")
	fmt.Fprintln(dst, "  return in.Error()")
	fmt.Fprintln(dst, "}")
	return nil
}

func (engineGenerator) GenerateMarshalers(dst io.Writer, encodeFuncName, typeName string) error {
	fmt.Fprintln(dst, "// MarshalYAML supports yayamls.Marshaler")
	fmt.Fprintln(dst, "func (v "+typeName+") MarshalYAML() ([]byte, error) {")
	fmt.Fprintln(dst, "  out := yamly.NewEncoder(encode.NewASTBuilder(), encode.NewASTWriter())")
	fmt.Fprintln(dst, "  "+encodeFuncName+"(out, v)")
	fmt.Fprintln(dst, "  return out.EncodeToBytes()")
	fmt.Fprintln(dst, "}")
	return nil
}

func (engineGenerator) UnmarshalersImplementationCheck(
	dst io.Writer,
	t reflect.Type,
	outArg string,
	indent int,
) (generator.ImplementationResult, error) {
	unmarshalIface := reflect.TypeOf((*Unmarshaler)(nil)).Elem()
	if reflect.PtrTo(t).Implements(unmarshalIface) {
		fmt.Fprintln(dst, strings.Repeat(" ", indent)+"in.AddError(("+outArg+").UnmarshalYAML(in.Raw()))")
		return generator.ImplementationResultTrue, nil
	}
	return generator.ImplementationResultFalse, nil
}

func (engineGenerator) MarshalersImplementationCheck(
	dst io.Writer,
	t reflect.Type,
	inArg string,
	indent int,
) (generator.ImplementationResult, error) {
	marshalIface := reflect.TypeOf((*Marshaler)(nil)).Elem()
	if reflect.PtrTo(t).Implements(marshalIface) {
		fmt.Fprintln(dst, strings.Repeat(" ", indent)+"out.InsertRaw("+inArg+".MarshalYAML())")
		return generator.ImplementationResultTrue, nil
	}
	return generator.ImplementationResultFalse, nil
}

func (engineGenerator) GenerateUnmarshalEmptyInterfaceAssertions(dst io.Writer, outArg string, indent int) error {
	whitespace := strings.Repeat(" ", indent)
	fmt.Fprintln(dst, whitespace+"if m, ok := "+outArg+".(yayamls.Unmarshaler); ok {")
	fmt.Fprintln(dst, whitespace+"  in.AddError(m.UnmarshalYAML(in.Raw()))")
	fmt.Fprintln(dst, whitespace+"} else {")
	fmt.Fprintln(dst, whitespace+"  "+outArg+" = in.Any()")
	fmt.Fprintln(dst, whitespace+"}")
	return nil
}

func (engineGenerator) GenerateMarshalEmptyInterfaceAssertions(dst io.Writer, inArg string, indent int) error {
	whitespace := strings.Repeat(" ", indent)
	fmt.Fprintln(dst, whitespace+"if m, ok := "+inArg+".(yayamls.Marshaler); ok {")
	fmt.Fprintln(dst, whitespace+"  out.InsertRaw(m.MarshalYAML())")
	fmt.Fprintln(dst, whitespace+"} else {")
	// TODO: add reflect-based marshaler for yayamls
	fmt.Fprintln(dst, whitespace+"  out.InsertRaw(nil, nil)")
	fmt.Fprintln(dst, whitespace+"}")
	return nil
}

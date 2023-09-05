package generator

import (
	"fmt"
	"github.com/KSpaceer/yayamls"
	"reflect"
	"strings"
)

func (g *Generator) encoderFunctionName(t reflect.Type) string {
	return g.generateFunctionName("encode", t)
}

var basicEncoderFormatStrings = map[reflect.Kind]string{
	reflect.String:  "out.InsertString(string(%v))",
	reflect.Bool:    "out.InsertBoolean(bool(%v))",
	reflect.Int:     "out.InsertInteger(int64(%v))",
	reflect.Int8:    "out.InsertInteger(int64(%v))",
	reflect.Int16:   "out.InsertInteger(int64(%v))",
	reflect.Int32:   "out.InsertInteger(int64(%v))",
	reflect.Int64:   "out.InsertInteger(int64(%v))",
	reflect.Uint:    "out.InsertUnsigned(uint64(%v))",
	reflect.Uint8:   "out.InsertUnsigned(uint64(%v))",
	reflect.Uint16:  "out.InsertUnsigned(uint64(%v))",
	reflect.Uint32:  "out.InsertUnsigned(uint64(%v))",
	reflect.Uint64:  "out.InsertUnsigned(uint64(%v))",
	reflect.Float32: "out.InsertFloat(float64(%v))",
	reflect.Float64: "out.InsertFloat(float64(%v))",
}

var customEncoderFormatStrings = map[string]string{
	"time.Time": "out.InsertTimestamp(time.Time(%v))",
}

func (g *Generator) generateEncoder(t reflect.Type) error {
	if t.Kind() == reflect.Struct {
		return g.generateStructEncoder(t)
	}

	fname := g.encoderFunctionName(t)
	tname := g.extractTypeName(t)

	if g.encodePointerReceiver {
		tname = "*" + tname
		t = reflect.PtrTo(t)
	}

	fmt.Fprintln(g.out, "func "+fname+"(out yayamls.Inserter, in "+tname+") {")
	if err := g.generateEncoderBodyWithoutCheck(t, "in", fieldTags{}, 2); err != nil {
		return err
	}
	fmt.Fprintln(g.out, "}")
	return nil
}

func (g *Generator) generateEncoderBodyWithoutCheck(
	t reflect.Type,
	inArg string,
	tags fieldTags,
	indent int,
	canBeNull bool,
) error {
	whitespace := strings.Repeat(" ", indent)

	if enc := basicEncoderFormatStrings[t.Kind()]; enc != "" {
		fmt.Fprintf(g.out, whitespace+enc+"\n", inArg)
		return nil
	} else if enc := customEncoderFormatStrings[t.Name()]; enc != "" {
		fmt.Fprintf(g.out, whitespace+enc+"\n", inArg)
		return nil
	}

	switch t.Kind() {
	case reflect.Slice:
		elem := t.Elem()

		if canBeNull {
			fmt.Fprintln(g.out, whitespace+"if "+inArg+" == nil {")
			fmt.Fprintln(g.out, whitespace+"  out.InsertNull()")
			fmt.Fprintln(g.out, whitespace+"} else {")
		} else {
			fmt.Fprintln(g.out, whitespace+"{")
		}

		if elem.Kind() == reflect.Uint8 && elem.Name() == "uint8" {
			fmt.Fprintln(g.out, whitespace+"  out.InsertString(string("+inArg+"))")
		} else {
			vVar := g.generateVarName()
			fmt.Fprintln(g.out, whitespace+"  out.StartSequence()")
			fmt.Fprintln(g.out, whitespace+"  for _, "+vVar+" := range "+inArg+" {")

			if err := g.generateEncoderBody(elem, vVar, tags, indent+2, true); err != nil {
				return err
			}

			fmt.Fprintln(g.out, whitespace+"  }")
			fmt.Fprintln(g.out, whitespace+"  out.EndSequence()")
			fmt.Fprintln(g.out, whitespace+"}")
		}
	case reflect.Array:
		elem := t.Elem()

		if elem.Kind() == reflect.Uint8 && elem.Name() == "uint8" {
			fmt.Fprintln(g.out, whitespace+"out.InsertString(string("+inArg+"[:]))")
		} else {
			iVar := g.generateVarName()
			fmt.Fprintln(g.out, whitespace+"out.StartSequence()")
			fmt.Fprintln(g.out, whitespace+"for "+iVar+" := range "+inArg+" {")

			if err := g.generateEncoderBody(elem, "("+inArg+")["+iVar+"]", tags, indent+2, true); err != nil {
				return err
			}

			fmt.Fprintln(g.out, whitespace+"}")
			fmt.Fprintln(g.out, whitespace+"out.EndSequence()")
		}
	case reflect.Struct:
		enc := g.encoderFunctionName(t)
		g.addType(t)
		if g.encodePointerReceiver {
			if len(inArg) > 0 && inArg[0] == '*' {
				fmt.Fprintln(g.out, whitespace+enc+"(out, "+inArg[1:]+")")
			} else {
				fmt.Fprintln(g.out, whitespace+enc+"(out, &"+inArg+")")
			}
		} else {
			fmt.Fprintln(g.out, whitespace+enc+"(out, "+inArg+")")
		}
	case reflect.Pointer:
		extraIndent := 2
		if canBeNull {
			fmt.Fprintln(g.out, whitespace+"if "+inArg+" == nil {")
			fmt.Fprintln(g.out, whitespace+"  out.InsertNull()")
			fmt.Fprintln(g.out, whitespace+"} else {")
			extraIndent += 2
		}

		if err := g.generateEncoderBody(t.Elem(), "*"+inArg, tags, indent+extraIndent, true); err != nil {
			return err
		}

		if canBeNull {
			fmt.Fprintln(g.out, whitespace+"}")
		}
	case reflect.Map:
		key, elem := t.Key(), t.Elem()
		keyVar, valueVar := g.generateVarName("Key"), g.generateVarName("Value")

		if canBeNull {
			fmt.Fprintln(g.out, whitespace+"if "+inArg+" == nil {")
			fmt.Fprintln(g.out, whitespace+"  out.InsertNull()")
			fmt.Fprintln(g.out, whitespace+"} else {")
		} else {
			fmt.Fprintln(g.out, whitespace+"}")
		}

		fmt.Fprintln(g.out, whitespace+"  out.StartMapping()")
		fmt.Fprintln(g.out, whitespace+"  for "+keyVar+", "+valueVar+" := range "+inArg+" {")

		if err := g.generateEncoderBody(key, keyVar, tags, indent+4, true); err != nil {
			return err
		}

		if err := g.generateEncoderBody(elem, valueVar, tags, indent+4, true); err != nil {
			return err
		}

		fmt.Fprintln(g.out, whitespace+"  }")
		fmt.Fprintln(g.out, whitespace+"  out.EndMapping()")
		fmt.Fprintln(g.out, whitespace+"}")
	case reflect.Interface:
		if t.NumMethod() > 0 {
			if implementMarshalerYAYAMLS(t) {
				fmt.Fprintln(g.out, whitespace+"_ = "+inArg+".MarshalYAYAMLS(out)")
			} else if implementMarshalerYAYAMLS(t) {
				fmt.Fprintln(g.out, whitespace+"_ = out.InsertRaw("+inArg+".MarshalYAML(")
			} else {
				return fmt.Errorf("interface type %v is not supported: expect only interface{} "+
					"(any) or yayamls unmarshalers", t)
			}
		} else {

		}
	}
}

func implementsMarshaler(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*yayamls.Marshaler)(nil)).Elem())
}

func implementMarshalerYAYAMLS(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*yayamls.MarshalerYAYAMLS)(nil)).Elem())
}

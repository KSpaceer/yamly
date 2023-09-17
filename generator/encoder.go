package generator

import (
	"encoding"
	"fmt"
	"github.com/KSpaceer/yamly"
	"reflect"
	"strings"
)

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

func (g *Generator) generateMarshaler(t reflect.Type) error {
	fname := g.encoderFunctionName(t)
	tname := g.extractTypeName(t)

	if g.encodePointerReceiver {
		tname = "*" + tname
	}

	if err := g.engineGen.GenerateMarshalers(g.out, fname, tname); err != nil {
		return err
	}

	fmt.Fprintln(g.out)

	fmt.Fprintln(g.out, "// MarshalYamly supports yamly.MarshalerYamly interface")
	fmt.Fprintln(g.out, "func (v "+tname+") MarshalYamly(out yamly.Inserter) {")
	fmt.Fprintln(g.out, "  "+fname+"(out, v)")
	fmt.Fprintln(g.out, "}")
	return nil
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

	fmt.Fprintln(g.out, "func "+fname+"(out yamly.Inserter, in "+tname+") {")
	if err := g.generateEncoderBodyWithoutCheck(t, "in", fieldTags{}, 2, true); err != nil {
		return err
	}
	fmt.Fprintln(g.out, "}")
	return nil
}

func (g *Generator) generateStructEncoder(t reflect.Type) error {
	fname := g.encoderFunctionName(t)
	tname := g.extractTypeName(t)

	if g.encodePointerReceiver {
		tname = "*" + tname
	}

	fmt.Fprintln(g.out, "func "+fname+"(out yamly.Inserter, in "+tname+") {")
	if g.encodePointerReceiver {
		fmt.Fprintln(g.out, "  if in == nil {")
		fmt.Fprintln(g.out, "    out.InsertNull()")
		fmt.Fprintln(g.out, "    return")
		fmt.Fprintln(g.out, "  }")
	}

	fs, err := getStructFields(t)
	if err != nil {
		return fmt.Errorf("cannot generate encoder for %s: %w", t, err)
	}

	fmt.Fprintln(g.out, "  out.StartMapping()")
	for _, f := range fs {
		if err := g.generateStructFieldEncoder(f); err != nil {
			return err
		}
	}
	fmt.Fprintln(g.out, "  out.EndMapping()")
	fmt.Fprintln(g.out, "}")
	return nil
}

func (g *Generator) generateStructFieldEncoder(f reflect.StructField) error {
	tags := parseTags(f.Tag)

	if tags.omitField {
		return nil
	}
	name := f.Name
	if tags.name != "" {
		name = tags.name
	}
	canBeNull := !(g.omitempty || tags.omitempty)

	if !canBeNull {
		fmt.Fprintln(g.out, "  if "+g.generateNotEmptyCheck(f.Type, "in."+f.Name)+" {")
	} else {
		fmt.Fprintln(g.out, "  {")
	}

	fmt.Fprintln(g.out, "    out.InsertString(\""+name+"\")")
	if err := g.generateEncoderBody(f.Type, "in."+f.Name, tags, 6, canBeNull); err != nil {
		return err
	}
	fmt.Fprintln(g.out, "  }")

	return nil
}

func (g *Generator) generateNotEmptyCheck(t reflect.Type, arg string) string {
	switch t.Kind() {
	case reflect.Slice, reflect.Map:
		return "len(" + arg + ") > 0"
	case reflect.Interface, reflect.Pointer:
		return arg + " != nil"
	case reflect.Bool:
		return arg
	case reflect.String:
		return arg + ` != ""`
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return arg + " != 0"
	default:
		// array does not have "empty" value
		return "true"
	}
}

func (g *Generator) generateEncoderBody(
	t reflect.Type,
	inArg string,
	tags fieldTags,
	indent int,
	canBeNull bool,
) error {
	var finishingText string
	if t != g.currentType {
		whitespace := strings.Repeat(" ", indent)

		marshalIface := reflect.TypeOf((*yamly.MarshalerYamly)(nil)).Elem()
		if reflect.PtrTo(t).Implements(marshalIface) {
			fmt.Fprintln(g.out, whitespace+inArg+".MarshalYamly(out)")
			return nil
		}

		implResult, err := g.engineGen.MarshalersImplementationCheck(g.out, t, inArg, indent)
		if err != nil {
			return err
		}
		switch implResult {
		case ImplementationResultTrue:
			return nil
		case ImplementationResultConditional:
			indent += 2
			finishingText = whitespace + "}"
			whitespace = strings.Repeat(" ", indent)
		}

		marshalIface = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
		if reflect.PtrTo(t).Implements(marshalIface) {
			fmt.Fprintln(g.out, whitespace+"out.InsertRawText("+inArg+".MarshalText())")
			return nil
		}
	}

	err := g.generateEncoderBodyWithoutCheck(t, inArg, tags, indent, canBeNull)
	if err != nil {
		return err
	}
	fmt.Fprintln(g.out, finishingText)
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
	} else if enc := customEncoderFormatStrings[t.String()]; enc != "" {
		g.imports[t.PkgPath()] = t.Name() // assume it is a standard library
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
		}
		fmt.Fprintln(g.out, whitespace+"}")
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
			fmt.Fprintln(g.out, whitespace+"{")
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
			if implementMarshalerYamly(t) {
				fmt.Fprintln(g.out, whitespace+"_ = "+inArg+".MarshalYamly(out)")
			} else if implResult, err := g.engineGen.MarshalersImplementationCheck(g.out, t, inArg, indent); err != nil {
				return err
			} else {
				switch implResult {
				case ImplementationResultFalse:
					return fmt.Errorf("interface type %v is not supported: expect only interface{} "+
						"(any), yamly.Marshaler or engine-specific marshalling interfaces", t)
				case ImplementationResultConditional:
					fmt.Fprintln(g.out, whitespace+"  out.InsertRawText(nil, yamly.MarshalerImplementationError)")
					fmt.Fprintln(g.out, whitespace+"}")
				}
			}
		} else {
			fmt.Fprintln(g.out, whitespace+"if m, ok := "+inArg+".(yamly.MarshalerYamly) {")
			fmt.Fprintln(g.out, whitespace+"  m.MarshalYamly(out)")
			fmt.Fprintln(g.out, whitespace+"} else {")
			if err := g.engineGen.GenerateMarshalEmptyInterfaceAssertions(g.out, inArg, indent+2); err != nil {
				return err
			}
			fmt.Fprintln(g.out, whitespace+"}")
		}
	default:
		return fmt.Errorf("can't encode type %s", t)
	}
	return nil
}

func implementMarshalerYamly(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*yamly.MarshalerYamly)(nil)).Elem())
}

func (g *Generator) encoderFunctionName(t reflect.Type) string {
	return g.generateFunctionName("encode", t)
}

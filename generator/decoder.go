package generator

import (
	"encoding"
	"fmt"
	"github.com/KSpaceer/yamly"
	"reflect"
	"strconv"
	"strings"
)

var basicDecoders = map[reflect.Kind]string{
	reflect.String:  "in.String()",
	reflect.Bool:    "in.Boolean()",
	reflect.Int:     "in.Integer(0)",
	reflect.Int8:    "in.Integer(8)",
	reflect.Int16:   "in.Integer(16)",
	reflect.Int32:   "in.Integer(32)",
	reflect.Int64:   "in.Integer(64)",
	reflect.Uint:    "in.Unsigned(0)",
	reflect.Uint8:   "in.Unsigned(8)",
	reflect.Uint16:  "in.Unsigned(16)",
	reflect.Uint32:  "in.Unsigned(32)",
	reflect.Uint64:  "in.Unsigned(64)",
	reflect.Float32: "in.Float(32)",
	reflect.Float64: "in.Float(64)",
}

var customDecoders = map[string]string{
	"time.Time": "in.Timestamp()",
}

func (g *Generator) generateUnmarshaler(t reflect.Type) error {
	fname := g.decoderFunctionName(t)
	tname := g.extractTypeName(t)

	fmt.Fprintln(g.out, "// UnmarshalYAML supports yamly.Unmarshaler interface")
	fmt.Fprintln(g.out, "func (v *"+tname+") UnmarshalYAML(data []byte) error {")
	fmt.Fprintln(g.out, "  in, err := decode.NewASTReaderFromBytes(data)")
	fmt.Fprintln(g.out, "  if err != nil {")
	fmt.Fprintln(g.out, "    return err")
	fmt.Fprintln(g.out, "  }")
	fmt.Fprintln(g.out, "  "+fname+"(in, v)")
	fmt.Fprintln(g.out, "  return in.Error()")
	fmt.Fprintln(g.out, "}")

	fmt.Fprintln(g.out)

	fmt.Fprintln(g.out, "// UnmarshalYamly supports yamly.UnmarshalerYamly interface")
	fmt.Fprintln(g.out, "func (v *"+tname+") UnmarshalYamly(in yamly.Decoder) {")
	fmt.Fprintln(g.out, "  "+fname+"(in, v)")
	fmt.Fprintln(g.out, "}")

	return nil
}

func (g *Generator) generateDecoder(t reflect.Type) error {
	if t.Kind() == reflect.Struct {
		return g.generateStructDecoder(t)
	}

	fname := g.decoderFunctionName(t)
	tname := g.extractTypeName(t)

	fmt.Fprintln(g.out, "func "+fname+"(in yamly.Decoder, out *"+tname+") {")
	if err := g.generateDecoderBodyWithoutCheck(reflect.PtrTo(t), "out", fieldTags{}, 2, false); err != nil {
		return err
	}
	fmt.Fprintln(g.out, "}")
	return nil
}

func (g *Generator) generateStructDecoder(t reflect.Type) error {
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, but got %s", t)
	}
	fname := g.decoderFunctionName(t)
	tname := g.extractTypeName(t)

	fmt.Fprintln(g.out, "func "+fname+"(in yamly.Decoder, out *"+tname+") {")
	fmt.Fprintln(g.out, "  if in.TryNull() {")
	fmt.Fprintln(g.out, "    var zeroValue "+tname)
	fmt.Fprintln(g.out, "    *out = zeroValue")
	fmt.Fprintln(g.out, "  }")
	fmt.Fprintln(g.out)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.Anonymous || f.Type.Kind() != reflect.Pointer {
			continue
		}
		fmt.Fprintln(g.out, "  out."+f.Name+" = new("+g.extractTypeName(f.Type.Elem())+")")
	}

	fields, err := getStructFields(t)
	if err != nil {
		return fmt.Errorf("cannot generate decoder for %s: %w", t, err)
	}

	fmt.Fprintln(g.out, "  structMappingState := in.Mapping()")
	fmt.Fprintln(g.out, "  for structMappingState.HasUnprocessedItems() {")
	fmt.Fprintln(g.out, "    key := in.String()")
	fmt.Fprintln(g.out, "    if in.TryNull() {")
	fmt.Fprintln(g.out, "      continue")
	fmt.Fprintln(g.out, "    }")
	fmt.Fprintln(g.out, "    switch key {")
	for _, f := range fields {
		if err = g.generateStructFieldDecoder(f); err != nil {
			return err
		}
	}
	fmt.Fprintln(g.out, "    default:")
	if g.disallowUnknownFields {
		fmt.Fprintln(g.out, "      in.AddError(&yamly.UnknownFieldError{Field: key})")
	} else {
		fmt.Fprintln(g.out, "      in.Skip()")
	}
	fmt.Fprintln(g.out, "    }")
	fmt.Fprintln(g.out, "  }")
	fmt.Fprintln(g.out, "}")
	return nil
}

func (g *Generator) generateStructFieldDecoder(f reflect.StructField) error {
	tags := parseTags(f.Tag)

	if tags.omitField {
		return nil
	}
	name := f.Name
	if tags.name != "" {
		name = tags.name
	}

	fmt.Fprintln(g.out, "    case \""+name+"\":")
	if err := g.generateDecoderBody(f.Type, "out."+f.Name, tags, 6, true); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateDecoderBody(
	t reflect.Type,
	outArg string,
	tags fieldTags,
	indent int,
	complexTypeElem bool,
) error {
	if t != g.currentType {
		whitespace := strings.Repeat(" ", indent)

		unmarshalIface := reflect.TypeOf((*yamly.UnmarshalerYamly)(nil)).Elem()
		if reflect.PtrTo(t).Implements(unmarshalIface) {
			fmt.Fprintln(g.out, whitespace+"("+outArg+").UnmarshalYamly(in)")
			return nil
		}

		unmarshalIface = reflect.TypeOf((*yamly.Unmarshaler)(nil)).Elem()
		if reflect.PtrTo(t).Implements(unmarshalIface) {
			fmt.Fprintln(g.out, whitespace+"in.AddError(("+outArg+").UnmarshalYAML(in.Raw()))")
			return nil
		}

		unmarshalIface = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
		if reflect.PtrTo(t).Implements(unmarshalIface) {
			fmt.Fprintln(g.out, whitespace+"in.AddError(("+outArg+").UnmarshalText([]byte(in.String())))")
			return nil
		}
	}

	return g.generateDecoderBodyWithoutCheck(t, outArg, tags, indent, complexTypeElem)
}

func (g *Generator) generateDecoderBodyWithoutCheck(
	t reflect.Type,
	outArg string,
	tags fieldTags,
	indent int,
	complexTypeElem bool,
) error {
	whitespace := strings.Repeat(" ", indent)
	if dec := customDecoders[t.String()]; dec != "" {
		fmt.Fprintln(g.out, whitespace+outArg+" = "+dec)
		return nil
	} else if dec := basicDecoders[t.Kind()]; dec != "" {
		fmt.Fprintln(g.out, whitespace+outArg+" = "+g.extractTypeName(t)+"("+dec+")")
		return nil
	}

	switch t.Kind() {
	case reflect.Slice:
		elem := t.Elem()
		if elem.Kind() == reflect.Uint8 && elem.Name() == "uint8" {
			fmt.Fprintln(g.out, whitespace+"if in.TryNull() {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = nil")
			fmt.Fprintln(g.out, whitespace+"} else {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = []byte(in.String())")
			fmt.Fprintln(g.out, whitespace+"}")
		} else {
			sliceStateVar := g.generateVarName("SeqState")
			sliceElemVar := g.generateVarName()
			fmt.Fprintln(g.out, whitespace+"if in.TryNull() {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = nil")
			fmt.Fprintln(g.out, whitespace+"} else {")
			fmt.Fprintln(g.out, whitespace+"  "+sliceStateVar+" := in.Sequence()")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = make("+g.extractTypeName(t)+", 0, "+sliceStateVar+".Size())")
			fmt.Fprintln(g.out, whitespace+"  for "+sliceStateVar+".HasUnprocessedItems() {")
			fmt.Fprintln(g.out, whitespace+"    var "+sliceElemVar+" "+g.extractTypeName(elem))

			if err := g.generateDecoderBody(elem, sliceElemVar, tags, indent+4, true); err != nil {
				return err
			}

			fmt.Fprintln(g.out, whitespace+"    "+outArg+" = append("+outArg+", "+sliceElemVar+")")
			fmt.Fprintln(g.out, whitespace+"  }")
			fmt.Fprintln(g.out, whitespace+"}")
		}
	case reflect.Array:
		elem := t.Elem()

		if elem.Kind() == reflect.Uint8 && elem.Name() == "uint8" {
			fmt.Fprintln(g.out, whitespace+"if !in.TryNull() {")
			fmt.Fprintln(g.out, whitespace+"  copy("+outArg+", []byte(in.String()))")
			fmt.Fprintln(g.out, whitespace+"}")
		} else {
			arrayStateVar := g.generateVarName("SeqState")
			iterVar := g.generateVarName()
			fmt.Fprintln(g.out, whitespace+"if !in.TryNull() {")
			fmt.Fprintln(g.out, whitespace+"  "+iterVar+" := 0")
			fmt.Fprintln(g.out, whitespace+"  "+arrayStateVar+" := in.Sequence()")
			fmt.Fprintln(g.out, whitespace+"  for "+arrayStateVar+".HasUnprocessedItems() && "+iterVar+" < "+strconv.Itoa(t.Len())+"{")

			if err := g.generateDecoderBody(elem, "("+outArg+")["+iterVar+"]", tags, indent+4, true); err != nil {
				return err
			}
			fmt.Fprintln(g.out, whitespace+"    "+iterVar+"++")
			fmt.Fprintln(g.out, whitespace+"  }")
			fmt.Fprintln(g.out, whitespace+"}")
		}
	case reflect.Struct:
		dec := g.decoderFunctionName(t)
		g.addType(t)

		if len(outArg) > 0 && outArg[0] == '*' {
			fmt.Fprintln(g.out, whitespace+dec+"(in, "+outArg[1:]+")")
		} else {
			fmt.Fprintln(g.out, whitespace+dec+"(in, &"+outArg+")")
		}
	case reflect.Pointer:
		fmt.Fprintln(g.out, whitespace+"if in.TryNull() {")
		fmt.Fprintln(g.out, whitespace+"  "+outArg+" = nil")
		fmt.Fprintln(g.out, whitespace+"} else {")
		if complexTypeElem {
			fmt.Fprintln(g.out, whitespace+"  if "+outArg+" == nil {")
			fmt.Fprintln(g.out, whitespace+"    "+outArg+" = new("+g.extractTypeName(t.Elem())+")")
			fmt.Fprintln(g.out, whitespace+"  }")
		}

		if err := g.generateDecoderBody(t.Elem(), "*"+outArg, tags, indent+2, complexTypeElem); err != nil {
			return err
		}

		fmt.Fprintln(g.out, whitespace+"}")
	case reflect.Map:
		mapStateVar := g.generateVarName("MapState")
		key, elem := t.Key(), t.Elem()
		keyVar, valueVar := g.generateVarName("Key"), g.generateVarName("Value")
		fmt.Fprintln(g.out, whitespace+"if in.TryNull() {")
		fmt.Fprintln(g.out, whitespace+"  "+outArg+" = nil")
		fmt.Fprintln(g.out, whitespace+"} else {")
		fmt.Fprintln(g.out, whitespace+"  "+mapStateVar+" := in.Mapping()")
		if g.omitempty || tags.omitempty {
			fmt.Fprintln(g.out, whitespace+"  if "+mapStateVar+".Size() == 0 {")
			fmt.Fprintln(g.out, "    "+outArg+" = nil")
			fmt.Fprintln(g.out, "  } else {")
			fmt.Fprintln(g.out, whitespace+"    "+outArg+" = make(map["+g.extractTypeName(key)+"]"+g.extractTypeName(elem)+", "+
				mapStateVar+".Size())")
			fmt.Fprintln(g.out, whitespace+"  }")
		} else {
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = make(map["+g.extractTypeName(key)+"]"+g.extractTypeName(elem)+", "+
				mapStateVar+".Size())")
		}
		fmt.Fprintln(g.out)
		fmt.Fprintln(g.out, whitespace+"  for "+mapStateVar+".HasUnprocessedItems() {")
		fmt.Fprintln(g.out, whitespace+"    var (")
		fmt.Fprintln(g.out, whitespace+"      "+keyVar+" "+g.extractTypeName(key))
		fmt.Fprintln(g.out, whitespace+"      "+valueVar+" "+g.extractTypeName(elem))
		fmt.Fprintln(g.out, whitespace+"    )")

		if err := g.generateDecoderBody(key, keyVar, tags, indent+4, true); err != nil {
			return err
		}

		if err := g.generateDecoderBody(elem, valueVar, tags, indent+4, true); err != nil {
			return err
		}

		fmt.Fprintln(g.out, whitespace+"    ("+outArg+")["+keyVar+"] = "+valueVar)
		fmt.Fprintln(g.out, whitespace+"  }")
		fmt.Fprintln(g.out, whitespace+"}")
	case reflect.Interface:
		if t.NumMethod() > 0 {
			if implementsUnmarshalerYamly(t) {
				fmt.Fprintln(g.out, whitespace+"_ = "+outArg+".UnmarshalYamly(in)")
			} else if implementsUnmarshaler(t) {
				fmt.Fprintln(g.out, whitespace+"_ = "+outArg+".UnmarshalYAML(in.Raw())")
			} else {
				return fmt.Errorf("interface type %v is not supported: expect only interface{} "+
					"(any) or yamly unmarshalers", t)
			}
		} else {
			fmt.Fprintln(g.out, whitespace+"if m, ok := "+outArg+".(yamly.UnmarshalerYamly); ok {")
			fmt.Fprintln(g.out, whitespace+"  in.AddError(m.UnmarshalYamly(in))")
			fmt.Fprintln(g.out, whitespace+"} else if m, ok := "+outArg+".(yamly.Unmarshaler); ok {")
			fmt.Fprintln(g.out, whitespace+"  in.AddError(m.Unmarshal(in.Raw()))")
			fmt.Fprintln(g.out, whitespace+"} else {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = in.Any()")
			fmt.Fprintln(g.out, whitespace+"}")
		}
	default:
		return fmt.Errorf("can't decode type %s", t)
	}
	return nil
}

func implementsUnmarshaler(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*yamly.Unmarshaler)(nil)).Elem())
}

func implementsUnmarshalerYamly(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*yamly.UnmarshalerYamly)(nil)).Elem())
}

func (g *Generator) decoderFunctionName(t reflect.Type) string {
	return g.generateFunctionName("decode", t)
}

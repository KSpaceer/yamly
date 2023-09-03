package generator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var basicDecoders = map[reflect.Kind]string{
	reflect.String:  "in.ExpectString()",
	reflect.Bool:    "in.ExpectBoolean()",
	reflect.Int:     "in.ExpectInteger(0)",
	reflect.Int8:    "in.ExpectInteger(8)",
	reflect.Int16:   "in.ExpectInteger(16)",
	reflect.Int32:   "in.ExpectInteger(32)",
	reflect.Int64:   "in.ExpectInteger(64)",
	reflect.Uint:    "in.ExpectUnsigned(0)",
	reflect.Uint8:   "in.ExpectUnsigned(8)",
	reflect.Uint16:  "in.ExpectUnsigned(16)",
	reflect.Uint32:  "in.ExpectUnsigned(32)",
	reflect.Uint64:  "in.ExpectUnsigned(64)",
	reflect.Float32: "in.ExpectFloat(32)",
	reflect.Float64: "in.ExpectFloat(64)",
}

type nullableDecodeInfo struct {
	method string
	castTo string
}

var basicNullableDecoders = map[reflect.Kind]nullableDecodeInfo{
	reflect.String:  {method: "in.ExpectNullableString()"},
	reflect.Bool:    {method: "in.ExpectNullableBoolean()"},
	reflect.Int:     {method: "in.ExpectNullableInteger(0)", castTo: "int"},
	reflect.Int8:    {method: "in.ExpectNullableInteger(8)", castTo: "int8"},
	reflect.Int16:   {method: "in.ExpectNullableInteger(16)", castTo: "int16"},
	reflect.Int32:   {method: "in.ExpectNullableInteger(32)", castTo: "int32"},
	reflect.Int64:   {method: "in.ExpectNullableInteger(64)"},
	reflect.Uint:    {method: "in.ExpectNullableUnsigned(0)", castTo: "uint"},
	reflect.Uint8:   {method: "in.ExpectNullableUnsigned(8)", castTo: "uint8"},
	reflect.Uint16:  {method: "in.ExpectNullableUnsigned(16)", castTo: "uint16"},
	reflect.Uint32:  {method: "in.ExpectNullableUnsigned(32)", castTo: "uint32"},
	reflect.Uint64:  {method: "in.ExpectNullableUnsigned(64)"},
	reflect.Float32: {method: "in.ExpectNullableFloat(32)", castTo: "float32"},
	reflect.Float64: {method: "in.ExpectNullableFloat(64)"},
}

var customDecoders = map[string]string{
	"time.Time": "in.ExpectTimestamp()",
}

func (g *Generator) generateDecoder(t reflect.Type) error {
	fname := g.decoderFunctionName(t)
	tname := g.extractTypeName(t)

	fmt.Fprintln(g.out, "func", fname, "in yayamls.Decoder, out *", tname, ") {")
}

func (g *Generator) generateDecoderBody(
	t reflect.Type,
	outArg string,
	tags fieldTags,
	indent int,
) error {
	return nil
}

func (g *Generator) generateDecoderBodyWithoutCheck(
	t reflect.Type,
	outArg string,
	tags fieldTags,
	indent int,
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
			resultVar, isNotNullVar := g.generateVarName(), g.generateVarName("IsNotNull")
			fmt.Fprintln(g.out, whitespace+resultVar+", "+isNotNullVar+" := in.ExpectNullableString()")
			fmt.Fprintln(g.out, whitespace+"if "+isNotNullVar+" {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = []byte("+resultVar+")")
			fmt.Fprintln(g.out, whitespace+"} else {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = nil")
			fmt.Fprintln(g.out, whitespace+"}")
		} else {
			sliceStateVar, isNotNullVar := g.generateVarName("SeqState"), g.generateVarName("IsNotNull")
			sliceElemVar := g.generateVarName()
			fmt.Fprintln(g.out, whitespace+sliceStateVar+", "+isNotNullVar+" := in.ExpectNullableSequence()")
			fmt.Fprintln(g.out, whitespace+"if "+isNotNullVar+" {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = make("+g.extractTypeName(t)+", 0, "+sliceStateVar+".Size())")
			fmt.Fprintln(g.out, whitespace+"  for "+sliceStateVar+".HasUnprocessedItems() {")
			fmt.Fprintln(g.out, whitespace+"    var "+sliceElemVar+" "+g.extractTypeName(elem))
			fmt.Fprintln(g.out, whitespace+"    "+outArg+" = append("+outArg+", "+sliceElemVar+")")

			if err := g.generateDecoderBody(elem, sliceElemVar, tags, indent+4); err != nil {
				return err
			}

			fmt.Fprintln(g.out, whitespace+"  }")
			fmt.Fprintln(g.out, whitespace+"} else {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = nil")
			fmt.Fprintln(g.out, whitespace+"}")
		}
	case reflect.Array:
		elem := t.Elem()

		if elem.Kind() == reflect.Uint8 && elem.Name() == "uint8" {
			resultVar, isNotNullVar := g.generateVarName(), g.generateVarName("IsNotNull")
			fmt.Fprintln(g.out, whitespace+resultVar+", "+isNotNullVar+" := in.ExpectNullableString()")
			fmt.Fprintln(g.out, whitespace+"if "+isNotNullVar+" {")
			fmt.Fprintln(g.out, whitespace+"  copy("+outArg+", []byte("+resultVar+")[:])")
			fmt.Fprintln(g.out, whitespace+"} else {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = nil")
			fmt.Fprintln(g.out, whitespace+"}")
		} else {
			arrayStateVar, isNotNullVar := g.generateVarName("SeqState"), g.generateVarName("IsNotNull")
			iterVar := g.generateVarName()
			fmt.Fprintln(g.out, whitespace+arrayStateVar+", "+isNotNullVar+" := in.ExpectNullableSequence()")
			fmt.Fprintln(g.out, whitespace+"if "+isNotNullVar+" {")
			fmt.Fprintln(g.out, whitespace+"  "+iterVar+" := 0")
			fmt.Fprintln(g.out, whitespace+"  for "+arrayStateVar+".HasUnprocessedItems() && "+iterVar+" < "+strconv.Itoa(t.Len())+"{")

			if err := g.generateDecoderBody(elem, "("+outArg+")["+iterVar+"]", tags, indent+4); err != nil {
				return err
			}
			fmt.Fprintln(g.out, whitespace+"    "+iterVar+"++")
			fmt.Fprintln(g.out, whitespace+"  }")
			fmt.Fprintln(g.out, whitespace+"} else {")
			fmt.Fprintln(g.out, whitespace+"  "+outArg+" = nil")
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

	}
}

func (g *Generator) decoderFunctionName(t reflect.Type) string {
	return g.generateFunctionName("decode", t)
}

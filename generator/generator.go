package generator

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

const (
	pkgYamly  = "github.com/KSpaceer/yamly"
	pkgDecode = "github.com/KSpaceer/yamly/decode"
	pkgEncode = "github.com/KSpaceer/yamly/encode"
)

type Generator struct {
	out *bytes.Buffer

	outputFile string
	pkgName    string
	pkgPath    string

	buildTags             string
	omitempty             bool
	disallowUnknownFields bool
	encodePointerReceiver bool

	imports map[string]string

	targetType  reflect.Type
	currentType reflect.Type

	pendingTypes   []reflect.Type
	generatedTypes map[reflect.Type]bool

	funcNames map[string]reflect.Type

	variablesCounter int
}

func New(outputFile string) *Generator {
	return &Generator{
		outputFile: outputFile,
		imports: map[string]string{
			pkgYamly:  "yamly",
			pkgDecode: "decode",
			pkgEncode: "encode",
		},
		generatedTypes: make(map[reflect.Type]bool),
		funcNames:      make(map[string]reflect.Type),
	}
}

func (g *Generator) SetPkgName(pkgName string) {
	g.pkgName = pkgName
}

func (g *Generator) SetPkgPath(pkgPath string) {
	g.pkgPath = pkgPath
}

func (g *Generator) SetBuildTags(buildTags string) {
	g.buildTags = buildTags
}

func (g *Generator) SetOmitempty(omitempty bool) {
	g.omitempty = omitempty
}

func (g *Generator) SetDisallowUnknownFields(disallow bool) {
	g.disallowUnknownFields = disallow
}

func (g *Generator) SetEncodePointerReceiver(useReceiver bool) {
	g.encodePointerReceiver = useReceiver
}

func (g *Generator) AddType(v any) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	g.addType(t)
	g.targetType = t
}

func (g *Generator) addType(t reflect.Type) {
	if g.generatedTypes[t] {
		return
	}
	for _, pendingType := range g.pendingTypes {
		if pendingType == t {
			return
		}
	}
	g.pendingTypes = append(g.pendingTypes, t)
}

func (g *Generator) Generate(w io.Writer) error {
	g.out = &bytes.Buffer{}

	for len(g.pendingTypes) > 0 {
		t := g.pendingTypes[len(g.pendingTypes)-1]
		g.pendingTypes = g.pendingTypes[:len(g.pendingTypes)-1]
		g.generatedTypes[t] = true
		g.currentType = t

		if err := g.generateDecoder(t); err != nil {
			return err
		}

		if err := g.generateEncoder(t); err != nil {
			return err
		}

		if t != g.targetType {
			continue
		}

		if err := g.generateUnmarshaler(t); err != nil {
			return err
		}

		if err := g.generateMarshaler(t); err != nil {
			return err
		}
	}
	var out bytes.Buffer
	g.generateHeader(&out)
	g.out.WriteTo(&out)

	_, err := w.Write(out.Bytes())
	return err
}

func (g *Generator) generateHeader(out io.Writer) {
	if g.buildTags != "" {
		fmt.Fprintln(out, "// +build ", g.buildTags)
		fmt.Fprintln(out)
	}
	fmt.Fprintln(out, "// Code generated by yamlygen. DO NOT EDIT.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "package", g.pkgName)
	fmt.Fprintln(out)

	byAlias := make(map[string]string, len(g.imports))
	aliases := make([]string, 0, len(g.imports))

	for path, alias := range g.imports {
		aliases = append(aliases, alias)
		byAlias[alias] = path
	}

	slices.Sort(aliases)
	fmt.Fprintln(out, "import (")
	for _, alias := range aliases {
		fmt.Fprintf(out, "  %s %q\n", alias, byAlias[alias])
	}
	fmt.Fprintln(out, ")")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "// suppress unused package warning")
	fmt.Fprintln(out, "var (")
	fmt.Fprintln(out, "  _ yamly.Marshaler")
	fmt.Fprintln(out, "  _ *encode.ASTWriter")
	fmt.Fprintln(out, "  _ *decode.ASTReader")
	fmt.Fprintln(out, ")")

	fmt.Fprintln(out)
}

func (g *Generator) extractTypeName(t reflect.Type) string {
	if t.Name() == "" {
		switch t.Kind() {
		case reflect.Pointer:
			return "*" + g.extractTypeName(t.Elem())
		case reflect.Slice:
			return "[]" + g.extractTypeName(t.Elem())
		case reflect.Array:
			return "[" + strconv.Itoa(t.Len()) + "]" + g.extractTypeName(t.Elem())
		case reflect.Map:
			return "map[" + g.extractTypeName(t.Key()) + "]" + g.extractTypeName(t.Elem())
		}
	}

	switch t.PkgPath() {
	case "":
		if t.Kind() == reflect.Struct {
			// the fields of anonymous struct may have named fields
			// and therefore we can't just use String() because
			// it does not remove the package name when it matches g.pkgPath
			nf := t.NumField()
			lines := make([]string, 0, nf)
			for i := 0; i < nf; i++ {
				f := t.Field(i)
				var line string
				if !f.Anonymous {
					line = f.Name + " "
				}
				line += g.extractTypeName(f.Type)
				tag := f.Tag
				if tag != "" {
					line += " " + escapeTag(tag)
				}
				lines = append(lines, line)
			}
			return strings.Join([]string{"struct { ", strings.Join(lines, "; "), " }"}, "")
		}
		return t.String()
	case g.pkgPath:
		return t.Name()
	default:
		return g.pkgAlias(t.PkgPath()) + "." + t.Name()
	}
}

func (g *Generator) pkgAlias(pkgPath string) string {
	pkgPath = processVendoring(pkgPath)
	if alias := g.imports[pkgPath]; alias != "" {
		return alias
	}

	alias := fixPkgAlias(filepath.Base(pkgPath))
	for i := 0; ; i++ {
		alias := alias
		if i > 0 {
			alias += strconv.Itoa(i)
		}

		var exists bool
		for _, v := range g.imports {
			if alias == v {
				exists = true
				break
			}
		}

		if !exists {
			g.imports[pkgPath] = alias
			return alias
		}
	}
}

func fixPkgAlias(alias string) string {
	alias = strings.NewReplacer(".", "_", "-", "_").Replace(alias)
	if alias[0] == 'v' {
		alias = "_" + alias
	}
	return alias
}

func processVendoring(pkgPath string) string {
	const vendorPath = "/vendor/"
	if i := strings.LastIndex(pkgPath, vendorPath); i != -1 {
		return pkgPath[i+len(vendorPath):]
	}
	return pkgPath
}

func escapeTag(tag reflect.StructTag) string {
	t := string(tag)
	if strings.ContainsRune(t, '`') {
		return strconv.Quote(t)
	}
	return "`" + t + "`"
}

func (g *Generator) generateVarName(suffixes ...string) string {
	g.variablesCounter++
	args := []any{"v", g.variablesCounter}
	for i := range suffixes {
		args = append(args, suffixes[i])
	}
	return fmt.Sprint(args...)
}

func (g *Generator) generateFunctionName(prefix string, t reflect.Type) string {
	prefix = joinFunctionNameParts(true, "yamly", prefix)
	name := joinFunctionNameParts(true, prefix, safeTypeName(t))

	if typ, ok := g.funcNames[name]; !ok || t == typ {
		g.funcNames[name] = t
		return name
	}

	for savedName, typ := range g.funcNames {
		if t == typ && strings.HasPrefix(savedName, name) {
			return savedName
		}
	}

	for i := 1; ; i++ {
		suffixedName := fmt.Sprint(name, i)
		if _, ok := g.funcNames[suffixedName]; ok {
			continue
		}
		g.funcNames[suffixedName] = t
		return suffixedName
	}
}

func safeTypeName(t reflect.Type) string {
	name := t.PkgPath()
	if t.Name() == "" {
		name += "anonymous" + anonymousStructSuffix(t)
	} else {
		name += "." + t.Name()
	}

	parts := []string{}
	singlePart := []rune{}

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			singlePart = append(singlePart, r)
		} else if len(singlePart) > 0 {
			parts = append(parts, string(singlePart))
			singlePart = singlePart[:0]
		}
	}

	if len(singlePart) > 0 {
		parts = append(parts, string(singlePart))
	}

	return joinFunctionNameParts(false, parts...)
}

func anonymousStructSuffix(t reflect.Type) string {
	if t.Kind() != reflect.Struct {
		return ""
	}

	var suffixBuilder bytes.Buffer
	for i := 0; i < t.NumField(); i++ {
		suffixBuilder.WriteString(t.Field(i).Name)
		suffixBuilder.WriteString(t.Field(i).Type.String())
	}
	hasher := fnv.New32()
	hasher.Write(suffixBuilder.Bytes())
	return fmt.Sprintf("%X", hasher.Sum32())
}

func joinFunctionNameParts(keepFirst bool, parts ...string) string {
	var buf strings.Builder
	for i := range parts {
		if i == 0 && keepFirst {
			buf.WriteString(parts[i])
		} else {
			n := len(parts[i])
			if n > 0 {
				buf.WriteString(strings.ToUpper(string(parts[i][0])))
			}
			if n > 1 {
				buf.WriteString(parts[i][1:])
			}
		}
	}
	return buf.String()
}

func getStructFields(t reflect.Type) ([]reflect.StructField, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, but got %s", t)
	}

	var (
		embeddedFields []reflect.StructField
		fields         []reflect.StructField
	)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tags := parseTags(f.Tag)
		if !f.Anonymous || tags.name != "" {
			continue
		}

		t := f.Type
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		if t.Kind() == reflect.Struct {
			fs, err := getStructFields(t)
			if err != nil {
				return nil, fmt.Errorf("error processing embedded field: %w", err)
			}
			embeddedFields = mergeStructFields(embeddedFields, fs)
		} else if (t.Kind() >= reflect.Bool && t.Kind() <= reflect.Complex128) || t.Kind() == reflect.String { // kind is basic
			if strings.Contains(f.Name, ".") || unicode.IsUpper([]rune(f.Name)[0]) {
				fields = append(fields, f)
			}
		}
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tags := parseTags(f.Tag)
		if f.Anonymous && tags.name == "" {
			continue
		}

		c := []rune(f.Name)[0]
		if unicode.IsUpper(c) {
			fields = append(fields, f)
		}
	}

	return mergeStructFields(embeddedFields, fields), nil
}

func mergeStructFields(firstFields, secondFields []reflect.StructField) []reflect.StructField {
	var fields []reflect.StructField
	used := make(map[string]bool)
	for _, f := range secondFields {
		used[f.Name] = true
		fields = append(fields, f)
	}

	for _, f := range firstFields {
		if !used[f.Name] {
			fields = append(fields, f)
		}
	}
	return fields
}

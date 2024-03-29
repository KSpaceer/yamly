package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/KSpaceer/yamly/generator/bootstrap"
	"github.com/KSpaceer/yamly/generator/parser"
)

var (
	buildTags             = flag.String("build-tags", "", "build tags to add to generated file")
	generatedType         = flag.String("type", "", "target type to generated marshaling methods")
	omitempty             = flag.Bool("omitempty", false, "omit empty fields by default")
	disallowUnknownFields = flag.Bool("disallow-unknown-fields", false, "return error if unknown field appeared in yaml")
	output                = flag.String("output", "", "name of generated file")
	encodePointerReceiver = flag.Bool("encode-pointer-receiver", false, "use pointer receiver in encode methods")
	engine                = flag.String("engine", "goyaml", "used parser engine for generated code")
	inlineEmbedded        = flag.Bool("inline-embedded", false, "inline embedded fields into YAML mapping")
)

func main() {
	flag.Parse()

	args := flag.Args()

	gofile := os.Getenv("GOFILE")

	var path string
	switch len(args) {
	case 0:
		if gofile != "" {
			path = filepath.Dir(gofile)
		} else {
			var err error
			path, err = os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	case 1:
		if isDirectory(args[0]) {
			path = args[0]
			break
		}
		fallthrough
	default:
		Usage()
		os.Exit(1)
	}

	if *generatedType == "" {
		Usage()
		os.Exit(1)
	}

	if err := generate(path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tyamlygen [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tyamlygen [flags] -type T # uses current directory\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func isDirectory(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return stat.IsDir()
}

func generate(path string) error {
	p := parser.Parser{}
	if err := p.Parse(path); err != nil {
		return fmt.Errorf("Error parsing %v: %w", path, err)
	}

	if err := os.Chdir(path); err != nil {
		return err
	}

	var outputName string
	if *output != "" {
		outputName = *output
	} else {
		outputName = toSnakeCase(*generatedType) + "_yamly.go"
	}

	if filepath.Base(outputName) != outputName {
		return fmt.Errorf("output name should only contain the file name, not path")
	}

	var trimmedBuildTags string
	if *buildTags != "" {
		trimmedBuildTags = strings.TrimSpace(*buildTags)
	}

	var engineGeneratorPackage, engineGenerator string
	switch *engine {
	case "goyaml":
		engineGeneratorPackage = "github.com/KSpaceer/yamly/engines/goyaml"
		engineGenerator = "Generator"
	case "yayamls":
		engineGeneratorPackage = "github.com/KSpaceer/yamly/engines/yayamls"
		engineGenerator = "Generator"
	default:
		return fmt.Errorf("unknown engine %q", *engine)
	}

	g := bootstrap.Generator{
		PkgPath:                p.PkgPath,
		PkgName:                p.PkgName,
		Type:                   *generatedType,
		Omitempty:              *omitempty,
		DisallowUnknownFields:  *disallowUnknownFields,
		EncodePointerReceiver:  *encodePointerReceiver,
		InlineEmbedded:         *inlineEmbedded,
		OutputName:             outputName,
		BuildTags:              trimmedBuildTags,
		EngineGeneratorPackage: engineGeneratorPackage,
		EngineGenerator:        engineGenerator,
	}

	if err := g.Generate(); err != nil {
		return fmt.Errorf("Bootstrap failed: %w", err)
	}
	return nil
}

func toSnakeCase(src string) string {
	buf := make([]rune, 0, len(src))
	var prev, cur rune
	for _, next := range src {
		switch {
		case cur == '_':
			if prev != '_' {
				buf = append(buf, '_')
			}
		case unicode.IsUpper(cur):
			if unicode.IsLower(prev) || (unicode.IsUpper(prev) && unicode.IsLower(next)) {
				buf = append(buf, '_')
			}
			buf = append(buf, unicode.ToLower(cur))
		case cur != 0:
			buf = append(buf, unicode.ToLower(cur))
		}
		prev, cur = cur, next
	}

	if len(src) > 0 {
		if unicode.IsUpper(cur) && unicode.IsLower(prev) && prev != 0 {
			buf = append(buf, '_')
		}
		buf = append(buf, unicode.ToLower(cur))
	}
	return string(buf)
}

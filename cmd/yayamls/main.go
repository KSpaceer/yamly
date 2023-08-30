package main

import (
	"flag"
	"fmt"
	"github.com/KSpaceer/yayamls/generator/parser"
	"os"
	"path/filepath"
)

var (
	buildTags             = flag.String("build-tags", "", "build tags to add to generated file")
	generatedType         = flag.String("type", "", "target type to generated marshaling methods")
	omitempty             = flag.Bool("omitempty", false, "omit empty fields by default")
	disallowUnknownFields = flag.Bool("disallow-unknown-fields", false, "return error if unknown field appeared in yaml")
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
	if err := generate(path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tyayamls [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tyayamls [flags] -type T # uses current directory\n")
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
		return fmt.Errorf("Error parsing %v: %v", path, err)
	}
}

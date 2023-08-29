package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	buildTags             = flag.String("build-tags", "", "build tags to add to generated file")
	generatedType         = flag.String("type", "", "target type to generated marshaling methods")
	omitempty             = flag.Bool("omitempty", false, "omit empty fields by default")
	outputFilename        = flag.String("output-filename", "", "specify the filename of the output")
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
		}
	default:
		path = filepath.Dir(args[0])
	}
}

func isDirectory(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return stat.IsDir()
}

func parse(args []string) error {

}

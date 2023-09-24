# yamly ![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/KSpaceer/yamly/yamly.yml) ![Codecov](https://img.shields.io/codecov/c/gh/KSpaceer/yamly) [![Go Report Card](https://goreportcard.com/badge/github.com/KSpaceer/yamly)](https://goreportcard.com/report/github.com/KSpaceer/yamly)

Package yamly provides a new way for YAML marshalling/unmarshalling, using code generation instead of runtime reflection. In this terms yamly is similar to ![easyjson](https://github.com/mailru/easyjson).
However, currently yamly is more proof of concept than really competitive and production-ready library.

## Usage

### Install

```
go get github.com/KSpaceer/yamly && go install github.com/KSpaceer/yamly/...@latest
```

### Run

```
yamlygen -type <target type> <package directory>
```

This command will generate a <target type>_yamly.go file with marshalling and unmarshalling functions for target type from package directory.

Like ![easyjson](https://github.com/mailru/easyjson) and ![ffjson](https://github.com/pquerna/ffjson), yamly code generation invokes ```go run``` on a temporary file, therefore a full Go build environment is required.

## Options

```
Flags:
  -build-tags string
    	build tags to add to generated file
  -disallow-unknown-fields
    	return error if unknown field appeared in yaml
  -encode-pointer-receiver
    	use pointer receiver in encode methods
  -engine string
    	used parser engine for generated code (default "goyaml")
  -inline-embedded
    	inline embedded fields into YAML mapping
  -omitempty
    	omit empty fields by default
  -output string
    	name of generated file
  -type string
    	target type to generated marshaling methods
```

## Struct tags

Currently yamly supports two struct tags:

- 'omitempty' - marshal field only in case it is not empty.
- 'inline' - inline the field, i.e. treat all field's nested as if they were the part of host struct.

## Engines

Yamly uses different parsing engines to generate code (i.e. engine is somewhat of 'backend' of marshalling). At this time yamly supports two engines:

- ```yayamls``` (Yet another YAML serializer) - self-made engine aiming to full coverage of YAML specification.
- ```goyaml``` - engine using ![go-yaml](https://github.com/go-yaml/yaml) as base.

## Performance

yamly is still raw and has pretty mediocre performance. With ```yayamls``` engine it is 2x times slower than ```go-yaml``` package, and with ```goyaml``` engine it only compares (not surpasses) the mentioned package.
In case of ```yayamls``` it is okay because ```yayamls``` has full coverage as main goal, but in case of ```goyaml``` it only brings overhead because of additional layer over parsing and marshalling, not mentioning difficulties of code generation.

The best way to resolve this problem is to implement new engine with direct access to lexing and tokenizing processes, which builds no AST.

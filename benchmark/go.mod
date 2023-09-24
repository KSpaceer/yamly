module github.com/KSpaceer/yamly/benchmark

go 1.21.1

require (
	github.com/KSpaceer/yamly v0.1.1
	github.com/KSpaceer/yamly/engines/goyaml v0.1.1
	github.com/KSpaceer/yamly/engines/yayamls v0.1.1
	github.com/goccy/go-yaml v1.11.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/fatih/color v1.10.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace github.com/KSpaceer/yamly => ../

replace github.com/KSpaceer/yamly/engines/yayamls => ../engines/yayamls

replace github.com/KSpaceer/yamly/engines/goyaml => ../engines/goyaml

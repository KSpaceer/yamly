module github.com/KSpaceer/yamly/test

go 1.21.1

require (
	github.com/KSpaceer/yamly v0.1.0
	github.com/KSpaceer/yamly/engines/goyaml v0.1.0
	github.com/KSpaceer/yamly/engines/yayamls v0.1.0
)

require gopkg.in/yaml.v3 v3.0.1

replace github.com/KSpaceer/yamly => ../

replace github.com/KSpaceer/yamly/engines/yayamls => ../engines/yayamls

replace github.com/KSpaceer/yamly/engines/goyaml => ../engines/goyaml

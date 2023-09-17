module github.com/KSpaceer/yamly/test

go 1.21.1

require (
	github.com/KSpaceer/yamly v0.0.0-00010101000000-000000000000
	github.com/KSpaceer/yamly/engines/goyaml v0.0.0-00010101000000-000000000000
	github.com/KSpaceer/yamly/engines/yayamls v0.0.0-00010101000000-000000000000
)

require gopkg.in/yaml.v3 v3.0.1

replace github.com/KSpaceer/yamly => ../

replace github.com/KSpaceer/yamly/engines/yayamls => ../engines/yayamls

replace github.com/KSpaceer/yamly/engines/goyaml => ../engines/goyaml

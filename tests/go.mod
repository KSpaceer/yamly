module github.com/KSpaceer/yamly/tests

go 1.21.0

require (
	github.com/KSpaceer/yamly v0.0.0-00010101000000-000000000000
	github.com/KSpaceer/yamly/engines/yayamls v0.0.0-00010101000000-000000000000
)

replace github.com/KSpaceer/yamly => ../

replace github.com/KSpaceer/yamly/engines/yayamls => ../engines/yayamls

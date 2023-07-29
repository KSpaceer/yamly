package parser

type parseOptions struct {
	tokenStreamConstructor func(string) ConfigurableTokenStream
}

type ParseOption interface {
	apply(options *parseOptions)
}

type parseOptionsFunc func(options *parseOptions)

func (f parseOptionsFunc) apply(options *parseOptions) {
	f(options)
}

func WithTokenStreamConstructor(tsConstructor func(string) ConfigurableTokenStream) ParseOption {
	return parseOptionsFunc(func(options *parseOptions) {
		options.tokenStreamConstructor = tsConstructor
	})
}

func applyOptions(opts ...ParseOption) parseOptions {
	var o parseOptions
	for _, opt := range opts {
		opt.apply(&o)
	}
	return o
}
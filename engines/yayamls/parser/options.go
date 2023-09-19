package parser

type parseOptions struct {
	tokenStreamConstructor func(string) ConfigurableTokenStream
	omitStream             bool
}

// ParseOption allows to modify parser behavior
type ParseOption interface {
	apply(options *parseOptions)
}

type parseOptionsFunc func(options *parseOptions)

func (f parseOptionsFunc) apply(options *parseOptions) {
	f(options)
}

// WithTokenStreamConstructor will make parser use the token stream constructed with given constructor function.
func WithTokenStreamConstructor(tsConstructor func(string) ConfigurableTokenStream) ParseOption {
	return parseOptionsFunc(func(options *parseOptions) {
		options.tokenStreamConstructor = tsConstructor
	})
}

// WithOmitStream will make parser to omit stream node from AST in case the stream has only one document.
func WithOmitStream() ParseOption {
	return parseOptionsFunc(func(options *parseOptions) {
		options.omitStream = true
	})
}

func applyOptions(opts ...ParseOption) parseOptions {
	var o parseOptions
	for _, opt := range opts {
		opt.apply(&o)
	}
	return o
}

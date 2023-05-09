package ifaces

type YAMLParser interface {
	ParseYAML(bytes []byte) (AST, error)
}

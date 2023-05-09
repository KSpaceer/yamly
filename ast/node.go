package ast

import (
	"fmt"
	"io"
)

type NodeType int8

const (
	UndefinedType NodeType = iota
	DocumentType
	ScalarType
	MappingType
	SequenceType
	CommentType
	DirectiveType
	TagType
)

type Node interface {
	io.Reader
	fmt.Stringer
	Type() NodeType
}

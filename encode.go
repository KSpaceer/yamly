package yayamls

import (
	"io"
	"time"
)

type Marshaler interface {
	MarshalYAML() ([]byte, error)
}

type MarshalerYAYAMLS interface {
	MarshalYAYAMLS(Inserter) error
}

type Inserter interface {
	InsertInteger(int64)

	InsertUnsigned(uint64)

	InsertBoolean(bool)

	InsertFloat(float64)

	InsertString(string)

	InsertTimestamp(time.Time)

	InsertNull()

	StartSequence()
	EndSequence()

	StartMapping()
	EndMapping()

	InsertRaw([]byte, error)
}

type Encoder interface {
	Inserter
	EncodeToString() (string, error)
	EncodeToBytes() ([]byte, error)
	EncodeTo(dst io.Writer) error
}

type TreeBuilder[T any] interface {
	Inserter
	Result() (T, error)
}

type TreeWriter[T any] interface {
	WriteTo(dst io.Writer, tree T) error
	WriteString(tree T) (string, error)
	WriteBytes(tree T) ([]byte, error)
}

type encoder[T any] struct {
	builder TreeBuilder[T]
	writer  TreeWriter[T]
}

func NewEncoder[T any](builder TreeBuilder[T], writer TreeWriter[T]) Encoder {
	return &encoder[T]{
		builder: builder,
		writer:  writer,
	}
}

func (e *encoder[T]) InsertInteger(val int64) {
	e.builder.InsertInteger(val)
}

func (e *encoder[T]) InsertUnsigned(val uint64) {
	e.builder.InsertUnsigned(val)
}

func (e *encoder[T]) InsertBoolean(val bool) {
	e.builder.InsertBoolean(val)
}

func (e *encoder[T]) InsertFloat(val float64) {
	e.builder.InsertFloat(val)
}

func (e *encoder[T]) InsertString(val string) {
	e.builder.InsertString(val)
}

func (e *encoder[T]) InsertTimestamp(val time.Time) {
	e.builder.InsertTimestamp(val)
}

func (e *encoder[T]) InsertNull() {
	e.builder.InsertNull()
}

func (e *encoder[T]) StartSequence() {
	e.builder.StartSequence()
}

func (e *encoder[T]) EndSequence() {
	e.builder.EndSequence()
}

func (e *encoder[T]) StartMapping() {
	e.builder.StartMapping()
}

func (e *encoder[T]) EndMapping() {
	e.builder.EndMapping()
}

func (e *encoder[T]) InsertRaw(data []byte, err error) {
	e.builder.InsertRaw(data, err)
}

func (e *encoder[T]) EncodeToString() (string, error) {
	tree, err := e.builder.Result()
	if err != nil {
		return "", err
	}
	return e.writer.WriteString(tree)
}

func (e *encoder[T]) EncodeToBytes() ([]byte, error) {
	tree, err := e.builder.Result()
	if err != nil {
		return nil, err
	}
	return e.writer.WriteBytes(tree)
}

func (e *encoder[T]) EncodeTo(dst io.Writer) error {
	tree, err := e.builder.Result()
	if err != nil {
		return err
	}
	return e.writer.WriteTo(dst, tree)
}

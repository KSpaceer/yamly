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
	InsertInteger(int64) error
	InsertNullableInteger(*int64) error

	InsertUnsigned(uint64) error
	InsertNullableUnsigned(*uint64) error

	InsertBoolean(bool) error
	InsertNullableBoolean(*bool) error

	InsertFloat(float64) error
	InsertNullableFloat(*float64) error

	InsertString(string) error
	InsertNullableString(*string) error

	InsertTimestamp(time.Time) error
	InsertNullableTimestamp(*time.Time) error

	InsertNull() error

	StartSequence() error
	EndSequence() error

	StartMapping() error
	EndMapping() error

	InsertRaw([]byte) error
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

func (e *encoder[T]) InsertInteger(val int64) error {
	return e.builder.InsertInteger(val)
}

func (e *encoder[T]) InsertNullableInteger(val *int64) error {
	return e.builder.InsertNullableInteger(val)
}

func (e *encoder[T]) InsertUnsigned(val uint64) error {
	return e.builder.InsertUnsigned(val)
}

func (e *encoder[T]) InsertNullableUnsigned(val *uint64) error {
	return e.builder.InsertNullableUnsigned(val)
}

func (e *encoder[T]) InsertBoolean(val bool) error {
	return e.builder.InsertBoolean(val)
}

func (e *encoder[T]) InsertNullableBoolean(val *bool) error {
	return e.builder.InsertNullableBoolean(val)
}

func (e *encoder[T]) InsertFloat(val float64) error {
	return e.builder.InsertFloat(val)
}

func (e *encoder[T]) InsertNullableFloat(val *float64) error {
	return e.builder.InsertNullableFloat(val)
}

func (e *encoder[T]) InsertString(val string) error {
	return e.builder.InsertString(val)
}

func (e *encoder[T]) InsertNullableString(val *string) error {
	return e.builder.InsertNullableString(val)
}

func (e *encoder[T]) InsertTimestamp(val time.Time) error {
	return e.builder.InsertTimestamp(val)
}

func (e *encoder[T]) InsertNullableTimestamp(val *time.Time) error {
	return e.builder.InsertNullableTimestamp(val)
}

func (e *encoder[T]) InsertNull() error {
	return e.builder.InsertNull()
}

func (e *encoder[T]) StartSequence() error {
	return e.builder.StartSequence()
}

func (e *encoder[T]) EndSequence() error {
	return e.builder.EndSequence()
}

func (e *encoder[T]) StartMapping() error {
	return e.builder.StartMapping()
}

func (e *encoder[T]) EndMapping() error {
	return e.builder.EndMapping()
}

func (e *encoder[T]) InsertRaw(data []byte) error {
	return e.builder.InsertRaw(data)
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

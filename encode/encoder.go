package encode

import (
	"io"
	"time"
)

type TreeBuilder[T any] interface {
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

	Result() (T, error)
}

type TreeWriter[T any] interface {
	WriteTo(dst io.Writer, tree T) error
	WriteString(tree T) (string, error)
	WriteBytes(tree T) ([]byte, error)
}

type Encoder[T any] struct {
	builder TreeBuilder[T]
	writer  TreeWriter[T]
}

func NewEncoder[T any](builder TreeBuilder[T], writer TreeWriter[T]) *Encoder[T] {
	return &Encoder[T]{
		builder: builder,
		writer:  writer,
	}
}

func (e *Encoder[T]) InsertInteger(val int64) error {
	return e.builder.InsertInteger(val)
}

func (e *Encoder[T]) InsertNullableInteger(val *int64) error {
	return e.builder.InsertNullableInteger(val)
}

func (e *Encoder[T]) InsertUnsigned(val uint64) error {
	return e.builder.InsertUnsigned(val)
}

func (e *Encoder[T]) InsertNullableUnsigned(val *uint64) error {
	return e.builder.InsertNullableUnsigned(val)
}

func (e *Encoder[T]) InsertBoolean(val bool) error {
	return e.builder.InsertBoolean(val)
}

func (e *Encoder[T]) InsertNullableBoolean(val *bool) error {
	return e.builder.InsertNullableBoolean(val)
}

func (e *Encoder[T]) InsertFloat(val float64) error {
	return e.builder.InsertFloat(val)
}

func (e *Encoder[T]) InsertNullableFloat(val *float64) error {
	return e.builder.InsertNullableFloat(val)
}

func (e *Encoder[T]) InsertString(val string) error {
	return e.builder.InsertString(val)
}

func (e *Encoder[T]) InsertNullableString(val *string) error {
	return e.builder.InsertNullableString(val)
}

func (e *Encoder[T]) InsertTimestamp(val time.Time) error {
	return e.builder.InsertTimestamp(val)
}

func (e *Encoder[T]) InsertNullableTimestamp(val *time.Time) error {
	return e.builder.InsertNullableTimestamp(val)
}

func (e *Encoder[T]) StartSequence() error {
	return e.builder.StartSequence()
}

func (e *Encoder[T]) EndSequence() error {
	return e.builder.EndSequence()
}

func (e *Encoder[T]) StartMapping() error {
	return e.builder.StartMapping()
}

func (e *Encoder[T]) EndMapping() error {
	return e.builder.EndMapping()
}

func (e *Encoder[T]) InsertRaw(data []byte) error {
	return e.builder.InsertRaw(data)
}

func (e *Encoder[T]) EncodeAsString() (string, error) {
	tree, err := e.builder.Result()
	if err != nil {
		return "", err
	}
	return e.writer.WriteString(tree)
}

func (e *Encoder[T]) EncodeAsBytes() ([]byte, error) {
	tree, err := e.builder.Result()
	if err != nil {
		return nil, err
	}
	return e.writer.WriteBytes(tree)
}

func (e *Encoder[T]) Encode(dst io.Writer) error {
	tree, err := e.builder.Result()
	if err != nil {
		return err
	}
	return e.writer.WriteTo(dst, tree)
}

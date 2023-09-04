package yayamls

import (
	"errors"
	"time"
)

var (
	EndOfStream = errors.New("end of YAML stream")
	Denied      = (*denyError)(nil)
)

type Unmarshaler interface {
	UnmarshalYAML([]byte) error
}

type UnmarshalerYAYAMLS interface {
	UnmarshalYAYAMLS(Decoder)
}

type CollectionState interface {
	Size() int
	HasUnprocessedItems() bool
}

type Decoder interface {
	TryNull() bool

	Integer(bitSize int) int64

	Unsigned(bitSize int) uint64

	Boolean() bool

	Float(bitSize int) float64

	String() string

	Timestamp() time.Time

	Sequence() CollectionState

	Mapping() CollectionState

	Any() any

	Raw() []byte

	Skip()

	Error() error

	AddError(err error)
}

type denyError struct {
	err error
}

func (de *denyError) Error() string {
	return de.err.Error()
}

func (de *denyError) Is(err error) bool {
	_, ok := err.(*denyError)
	return ok
}

func DenyError(err error) error {
	return &denyError{err}
}

type UnknownFieldError struct {
	Field string
}

func (ufe *UnknownFieldError) Error() string {
	return "unknown field " + ufe.Field
}

func (ufe *UnknownFieldError) Is(err error) bool {
	_, ok := err.(*UnknownFieldError)
	return ok
}

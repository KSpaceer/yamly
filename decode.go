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

type CollectionState interface {
	Size() int
	HasUnprocessedItems() bool
}

type Decoder interface {
	ExpectInteger() (int64, error)
	ExpectNullableInteger() (int64, bool, error)

	ExpectUnsigned() (uint64, error)
	ExpectNullableUnsigned() (uint64, bool, error)

	ExpectBoolean() (bool, error)
	ExpectNullableBoolean() (bool, bool, error)

	ExpectFloat() (float64, error)
	ExpectNullableFloat() (float64, bool, error)

	ExpectString() (string, error)
	ExpectNullableString() (string, bool, error)

	ExpectTimestamp() (time.Time, error)
	ExpectNullableTimestamp() (time.Time, bool, error)

	ExpectSequence() (CollectionState, error)
	ExpectNullableSequence() (CollectionState, bool, error)

	ExpectMapping() (CollectionState, error)
	ExpectNullableMapping() (CollectionState, bool, error)

	ExpectAny() (any, error)

	ExpectRaw() ([]byte, error)
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

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
	UnmarshalYAYAMLS(Decoder) error
}

type CollectionState interface {
	Size() int
	HasUnprocessedItems() bool
}

type Decoder interface {
	ExpectInteger(bitSize int) int64
	ExpectNullableInteger(bitSize int) (int64, bool)

	ExpectUnsigned(bitSize int) uint64
	ExpectNullableUnsigned(bitSize int) (uint64, bool)

	ExpectBoolean() bool
	ExpectNullableBoolean() (bool, bool)

	ExpectFloat(bitSize int) float64
	ExpectNullableFloat(bitSize int) (float64, bool)

	ExpectString() string
	ExpectNullableString() (string, bool)

	ExpectTimestamp() time.Time
	ExpectNullableTimestamp() (time.Time, bool)

	ExpectSequence() CollectionState
	ExpectNullableSequence() (CollectionState, bool)

	ExpectMapping() CollectionState
	ExpectNullableMapping() (CollectionState, bool)

	ExpectAny() any

	ExpectRaw() []byte

	Error() error
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

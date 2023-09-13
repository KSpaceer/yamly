package yamly

import (
	"errors"
	"time"
)

var (
	// EndOfStream indicates an end of YAML document (stream) during decoding.
	EndOfStream = errors.New("end of YAML stream")
	// Denied error indicates that Decoder called method was denied because current AST node
	// can't be represented as desired value.
	Denied = (*denyError)(nil)
)

// Unmarshaler interface can be implemented to customize type's behaviour when being
// unmarshaled from YAML document using raw YAML.
type Unmarshaler interface {
	UnmarshalYAML([]byte) error
}

// UnmarshalerYamly interface can be implemented to customize type's behaviour when being
// unmarshaled from YAML document using Decoder.
type UnmarshalerYamly interface {
	UnmarshalYamly(Decoder)
}

// CollectionState serves as a current state of complex node (e.g. sequence) for Decoder.
type CollectionState interface {
	// Size returns elements amount in the complex node.
	Size() int
	// HasUnprocessedItems shows if Decoder has not processed any element of the complex node yet.
	HasUnprocessedItems() bool
}

// Decoder is used to extract values from YAML AST.
type Decoder interface {
	// TryNull returns a boolean value indicating if current node is null node.
	// If it is null, proceeds to the next node.
	TryNull() bool

	// Integer extracts an integer value of given bit size from current text node.
	// If current node is not a text node or its value does not represent integer,
	// a Denied error is stored in Decoder.
	Integer(bitSize int) int64

	// Unsigned extracts an unsigned integer value of given bit size from current text node.
	// If current node is not a text node or its value does not represent unsigned integer,
	// a Denied error is stored in Decoder.
	Unsigned(bitSize int) uint64

	// Boolean extracts a boolean value of given bit size from current text node.
	// If current node is not a text node or its value does not represent boolean,
	// a Denied error is stored in Decoder.
	Boolean() bool

	// Float extracts a float value of given bit size from current text node.
	// If current node is not a text node or its value does not represent float,
	// a Denied error is stored in Decoder.
	Float(bitSize int) float64

	// String extracts a string value of given bit size from current text node.
	// If current node is not a text node, a Denied error is stored in Decoder.
	String() string

	// Timestamp extracts a time.Time value of given bit size from current text node.
	// If current node is not a text node or its value does not represent date,
	// a Denied error is stored in Decoder.
	Timestamp() time.Time

	// Sequence expects a sequence node in AST and returns a CollectionState associated with it.
	// If Decoder meets unexpected node (e.g. text or mapping), a Denied error is stored in Decoder.
	Sequence() CollectionState

	// Mapping expects a mapping node in AST and returns a CollectionState associated with it.
	// If Decoder meets unexpected node (e.g. text or sequence), a Denied error is stored in Decoder.
	Mapping() CollectionState

	// Any converts current subtree into Go value and returns it.
	Any() any

	// Raw serializes and returns current subtree as bytes.
	Raw() []byte

	// Skip skips current subtree and proceeds to the next.
	Skip()

	// Error returns a stored error.
	Error() error

	// AddError allows to add custom error to Decoder.
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

// DenyError marks given error as Denied error.
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

package yamly

import (
	"errors"
	"time"
)

var (
	// ErrEndOfStream indicates an end of YAML document (stream) during decoding.
	ErrEndOfStream = errors.New("end of YAML stream")
	// ErrDenied error indicates that Decoder called method was denied because current AST node
	// can't be represented as desired value.
	ErrDenied = (*denyError)(nil)
	// ErrUnmarshalerImplementation is used by engines to indicate that type does not
	// implement any unmarshalling interface in runtime.
	ErrUnmarshalerImplementation = errors.New("interface type is not supported: expect only interface{} " +
		"(any), yamly.Unmarshaler or engine-specific unmarshalling interfaces")
)

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
	// a ErrDenied error is stored in Decoder.
	Integer(bitSize int) int64

	// Unsigned extracts an unsigned integer value of given bit size from current text node.
	// If current node is not a text node or its value does not represent unsigned integer,
	// a ErrDenied error is stored in Decoder.
	Unsigned(bitSize int) uint64

	// Boolean extracts a boolean value of given bit size from current text node.
	// If current node is not a text node or its value does not represent boolean,
	// a ErrDenied error is stored in Decoder.
	Boolean() bool

	// Float extracts a float value of given bit size from current text node.
	// If current node is not a text node or its value does not represent float,
	// a ErrDenied error is stored in Decoder.
	Float(bitSize int) float64

	// String extracts a string value of given bit size from current text node.
	// If current node is not a text node, a ErrDenied error is stored in Decoder.
	String() string

	// Timestamp extracts a time.Time value of given bit size from current text node.
	// If current node is not a text node or its value does not represent date,
	// a ErrDenied error is stored in Decoder.
	Timestamp() time.Time

	// Sequence expects a sequence node in AST and returns a CollectionState associated with it.
	// If Decoder meets unexpected node (e.g. text or mapping), a ErrDenied error is stored in Decoder.
	Sequence() CollectionState

	// Mapping expects a mapping node in AST and returns a CollectionState associated with it.
	// If Decoder meets unexpected node (e.g. text or sequence), a ErrDenied error is stored in Decoder.
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

// ExtendedDecoder is used to extend Decoder interface with engine-specific
// methods
type ExtendedDecoder[T any] interface {
	Decoder

	// Node returns current subtree.
	Node() T
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

// DenyError marks given error as ErrDenied error.
func DenyError(err error) error {
	return &denyError{err}
}

// UnknownFieldError is used to indicate unknown field for struct
// when unmarshaler is generated with unknown field disallowing.
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

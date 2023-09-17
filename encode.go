package yamly

import (
	"errors"
	"io"
	"time"
)

var (
	// MarshalerImplementationError is used by engines to indicate that type does not
	// implement any marshalling interface in runtime.
	MarshalerImplementationError = errors.New("interface type is not supported: expect only interface{} " +
		"(any), yamly.Marshaler or engine-specific marshalling interfaces")
)

// MarshalerYamly interface can be implemented to customize type's behaviour when being
// marshaled into a YAML document inserting a subtree into builded AST.
type MarshalerYamly interface {
	MarshalYamly(Inserter)
}

// Inserter allows inserting values into an YAML AST.
// If any error occurs during insertion, further calls are no-op.
type Inserter interface {
	// InsertInteger inserts an integer into AST as text node
	InsertInteger(int64)

	// InsertUnsigned inserts an unsigned integer into AST as text node.
	InsertUnsigned(uint64)

	// InsertBoolean inserts a boolean value into AST as text node.
	InsertBoolean(bool)

	// InsertFloat inserts a float value into AST as text node.
	InsertFloat(float64)

	// InsertString inserts a string value into AST as text node.
	InsertString(string)

	// InsertTimestamp inserts a time.Time value into AST as text node.
	InsertTimestamp(time.Time)

	// InsertNull inserts a null node into AST.
	InsertNull()

	// StartSequence creates a sequence node in AST.
	// Until EndSequence is called, every inserted node will be inserted as sequence element.
	StartSequence()
	// EndSequence finishes a sequence created before with StartSequence.
	// If current node is not a sequence (i.e. StartSequence was not called or nested mapping/sequence is not finished),
	// an error is added to inserter.
	EndSequence()

	// StartMapping creates a mapping node in AST.
	// Until EndMapping is called, every odd inserted node will be inserted as mapping entry key
	// and every even inserted node will be inserted as mapping entry value.
	StartMapping()
	// EndMapping finished a mapping created before with StartMapping.
	// If current node is not a sequence (i.e. StartMapping was not called or nested mapping/sequence is not finished),
	// or there is a pending mapping entry (with created key and absent value),
	// an error is added to inserter
	EndMapping()

	// InsertRaw inserts given raw bytes in YAML as subtree into AST.
	// Also, it accepts an error to make it comfortable to call Marshaler.MarshalYAML to provide arguments.
	InsertRaw([]byte, error)
	// InsertRawText inserts given raw bytes as text node into AST.
	// Also, it accepts an error to make it comfortable to call encoding.TextMarshaler methods to provide arguments.
	InsertRawText([]byte, error)
}

// Encoder allows inserting values into an YAML AST
// and encoding builded AST as string or bytes.
type Encoder interface {
	Inserter

	// EncodeToString encodes the AST builded with Inserter methods as a string.
	EncodeToString() (string, error)
	// EncodeToBytes encodes the AST builded with Inserter methods as a byte slice.
	EncodeToBytes() ([]byte, error)
	// EncodeTo encodes the AST builded with Inserter methods and writes the serialized data into given io.Writer.
	EncodeTo(dst io.Writer) error
}

// TreeBuilder allows inserting values into a generic YAML AST
// and returning result AST.
type TreeBuilder[T any] interface {
	Inserter
	// Result returns the AST builded with Inserter methods.
	Result() (T, error)
}

// TreeWriter serializes given generic YAML AST
// as string or bytes.
type TreeWriter[T any] interface {
	// WriteTo serializes given tree and writes serialized bytes into given io.Writer.
	WriteTo(dst io.Writer, tree T) error
	// WriteString serializes given tree as string.
	WriteString(tree T) (string, error)
	// WriteBytes serializes given tree as bytes.
	WriteBytes(tree T) ([]byte, error)
}

// encoder implements Encoder using underlying TreeBuilder and TreeWriter
type encoder[T any] struct {
	builder TreeBuilder[T]
	writer  TreeWriter[T]
}

// NewEncoder returns an Encoder used to construct and serialize YAML AST tree of given type.
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

func (e *encoder[T]) InsertRawText(text []byte, err error) {
	e.builder.InsertRawText(text, err)
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

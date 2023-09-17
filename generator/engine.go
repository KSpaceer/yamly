package generator

import (
	"io"
	"reflect"
)

// EngineGenerator interface contains engine-specific methods
type EngineGenerator interface {
	// Packages returns information for packages used by engine as map
	// using package path as key and alias as value
	Packages() map[string]string

	// WarningSuppressors returns a list of types used to create stub variables
	// to suppress warnings
	WarningSuppressors() []string

	// GenerateUnmarshalers generates unmarshalling methods for target type
	// using provided decodeFuncName and typeName. The generated code is written into dst.
	// Decode function has following signature:
	//
	// func(yamly.Decoder, *typeName)
	GenerateUnmarshalers(dst io.Writer, decodeFuncName string, typeName string) error

	// GenerateMarshalers generates marshalling methods for target type
	// using provided encodeFuncName and typeName. The generated code is written into dst.
	// Encode function has following signature:
	//
	// func(yamly.Encoder, typeName)
	GenerateMarshalers(dst io.Writer, encodeFuncName string, typeName string) error

	// UnmarshalersImplementationCheck checks if provided type t implements any of engine-specific
	// unmarshalling interfaces.
	// If it does, UnmarshalersImplementationCheck generates code using implemented method, writing it into dst.
	UnmarshalersImplementationCheck(dst io.Writer, t reflect.Type, outArg string, indent int) (ImplementationResult, error)

	// MarshalersImplementationCheck checks if provided type t implements any of engine-specific
	// marshalling interfaces.
	// If it does, MarshalersImplementationCheck generates code using implemented method, writing it into dst.
	MarshalersImplementationCheck(dst io.Writer, t reflect.Type, inArg string, indent int) (ImplementationResult, error)

	// GenerateUnmarshalEmptyInterfaceAssertions generates code with type assertions for empty interface
	// (interface{} or any) using unmarshalling interfaces.
	GenerateUnmarshalEmptyInterfaceAssertions(dst io.Writer, outArg string, indent int) error

	// GenerateMarshalEmptyInterfaceAssertions generates code with type assertions for empty interface
	// (interface{} or any) using marshalling interfaces.
	GenerateMarshalEmptyInterfaceAssertions(dst io.Writer, inArg string, indent int) error
}

type ImplementationResult int8

const (
	ImplementationResultFalse ImplementationResult = iota
	ImplementationResultTrue
	ImplementationResultConditional
)

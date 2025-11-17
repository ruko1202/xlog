// Package field provides type aliases for zap.Field and field constructor functions.
package field

import "go.uber.org/zap"

// Field is a type alias for zap.Field, representing a marshaling operation.
type Field = zap.Field

// Field constructor function aliases.
var (
	// String constructs a field with a string value.
	String = zap.String

	// Int constructs a field with an int value.
	Int = zap.Int

	// Int64 constructs a field with an int64 value.
	Int64 = zap.Int64

	// Int32 constructs a field with an int32 value.
	Int32 = zap.Int32

	// Uint constructs a field with a uint value.
	Uint = zap.Uint

	// Uint64 constructs a field with a uint64 value.
	Uint64 = zap.Uint64

	// Uint32 constructs a field with a uint32 value.
	Uint32 = zap.Uint32

	// Float64 constructs a field with a float64 value.
	Float64 = zap.Float64

	// Float32 constructs a field with a float32 value.
	Float32 = zap.Float32

	// Bool constructs a field with a bool value.
	Bool = zap.Bool

	// Time constructs a field with a time.Time value.
	Time = zap.Time

	// Duration constructs a field with a time.Duration value.
	Duration = zap.Duration

	// Error constructs a field with an error value.
	Error = zap.Error

	// Any constructs a field with an arbitrary value.
	Any = zap.Any

	// Binary constructs a field with binary data.
	Binary = zap.Binary

	// ByteString constructs a field with a byte slice.
	ByteString = zap.ByteString

	// Complex64 constructs a field with a complex64 value.
	Complex64 = zap.Complex64

	// Complex128 constructs a field with a complex128 value.
	Complex128 = zap.Complex128

	// Namespace creates a namespace for subsequent fields.
	Namespace = zap.Namespace

	// Reflect constructs a field using reflection.
	Reflect = zap.Reflect

	// Stringer constructs a field with a fmt.Stringer value.
	Stringer = zap.Stringer

	// Skip constructs a no-op field.
	Skip = zap.Skip
)

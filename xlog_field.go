package xlog

import "go.uber.org/zap"

// Field is a type alias for zap.Field, representing a marshaling operation used in structured logging.
type Field = zap.Field

// Field constructor function aliases.
var (
	// StringField constructs a field with a string value.
	StringField = zap.String

	// IntField constructs a field with an int value.
	IntField = zap.Int

	// Int64Field constructs a field with an int64 value.
	Int64Field = zap.Int64

	// Int32Field constructs a field with an int32 value.
	Int32Field = zap.Int32

	// UintField constructs a field with a uint value.
	UintField = zap.Uint

	// Uint64Field constructs a field with a uint64 value.
	Uint64Field = zap.Uint64

	// Uint32Field constructs a field with a uint32 value.
	Uint32Field = zap.Uint32

	// Float64Field constructs a field with a float64 value.
	Float64Field = zap.Float64

	// Float32Field constructs a field with a float32 value.
	Float32Field = zap.Float32

	// BoolField constructs a field with a bool value.
	BoolField = zap.Bool

	// TimeField constructs a field with a time.Time value.
	TimeField = zap.Time

	// DurationField constructs a field with a time.Duration value.
	DurationField = zap.Duration

	// ErrorField constructs a field with an error value.
	ErrorField = zap.Error

	// AnyField constructs a field with an arbitrary value.
	AnyField = zap.Any

	// BinaryField constructs a field with binary data.
	BinaryField = zap.Binary

	// ByteStringField constructs a field with a byte slice.
	ByteStringField = zap.ByteString

	// Complex64Field constructs a field with a complex64 value.
	Complex64Field = zap.Complex64

	// Complex128Field constructs a field with a complex128 value.
	Complex128Field = zap.Complex128

	// NamespaceField creates a namespace for subsequent fields.
	NamespaceField = zap.Namespace

	// ReflectField constructs a field using reflection.
	ReflectField = zap.Reflect

	// StringerField constructs a field with a fmt.Stringer value.
	StringerField = zap.Stringer

	// SkipField constructs a no-op field.
	SkipField = zap.Skip
)

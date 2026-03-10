package xlog

import (
	"fmt"
	"time"
)

// FieldType represents the type of a field value.
type FieldType uint8

const (
	// UnknownType is the default field type.
	UnknownType FieldType = iota
	// StringType indicates a string field.
	StringType
	// Int64Type indicates an int64 field (also used for int, int32, etc).
	Int64Type
	// Uint64Type indicates a uint64 field (also used for uint, uint32, etc).
	Uint64Type
	// Float64Type indicates a float64 field (also used for float32).
	Float64Type
	// BoolType indicates a boolean field.
	BoolType
	// TimeType indicates a time.Time field.
	TimeType
	// DurationType indicates a time.Duration field.
	DurationType
	// ErrorType indicates an error field.
	ErrorType
	// AnyType indicates an arbitrary value (will be formatted as needed).
	AnyType
	// ArrayType indicates an array/slice field.
	ArrayType
	// ObjectType indicates a complex object field.
	ObjectType
	// BinaryType indicates a binary/byte slice field.
	BinaryType
)

// Field represents a structured logging field with a key-value pair.
// This is a universal field type that can be adapted to any logging backend.
type Field struct {
	Key       string
	Type      FieldType
	String    string
	Integer   int64
	Float     float64
	Interface interface{}
}

// String creates a string field.
func String(key, val string) Field {
	return Field{Key: key, Type: StringType, String: val}
}

// Int creates an int field.
func Int(key string, val int) Field {
	return Field{Key: key, Type: Int64Type, Integer: int64(val)}
}

// Int64 creates an int64 field.
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: Int64Type, Integer: val}
}

// Int32 creates an int32 field.
func Int32(key string, val int32) Field {
	return Field{Key: key, Type: Int64Type, Integer: int64(val)}
}

// Uint creates a uint field.
func Uint(key string, val uint) Field {
	return Field{Key: key, Type: Uint64Type, Integer: int64(val)}
}

// Uint64 creates a uint64 field.
func Uint64(key string, val uint64) Field {
	return Field{Key: key, Type: Uint64Type, Integer: int64(val)}
}

// Uint32 creates a uint32 field.
func Uint32(key string, val uint32) Field {
	return Field{Key: key, Type: Uint64Type, Integer: int64(val)}
}

// Float64 creates a float64 field.
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: Float64Type, Float: val}
}

// Float32 creates a float32 field.
func Float32(key string, val float32) Field {
	return Field{Key: key, Type: Float64Type, Float: float64(val)}
}

// Bool creates a boolean field.
func Bool(key string, val bool) Field {
	f := Field{Key: key, Type: BoolType, Integer: 0}
	if val {
		f.Integer = 1
	}
	return f
}

// Time creates a time.Time field.
func Time(key string, val time.Time) Field {
	return Field{Key: key, Type: TimeType, Integer: val.UnixNano(), Interface: val}
}

// Duration creates a time.Duration field.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: DurationType, Integer: int64(val)}
}

// Err creates an error field with key "error".
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Type: ErrorType, Interface: nil}
	}
	return Field{Key: "error", Type: ErrorType, String: err.Error(), Interface: err}
}

// NamedError creates a named error field.
func NamedError(key string, err error) Field {
	if err == nil {
		return Field{Key: key, Type: ErrorType, Interface: nil}
	}
	return Field{Key: key, Type: ErrorType, String: err.Error(), Interface: err}
}

// Any creates a field with an arbitrary value.
// The value will be formatted based on the backend's implementation.
func Any(key string, val interface{}) Field {
	return Field{Key: key, Type: AnyType, Interface: val}
}

// Strings creates a field with a string slice.
func Strings(key string, val []string) Field {
	return Field{Key: key, Type: ArrayType, Interface: val}
}

// Ints creates a field with an int slice.
func Ints(key string, val []int) Field {
	return Field{Key: key, Type: ArrayType, Interface: val}
}

// Binary creates a field with binary data.
func Binary(key string, val []byte) Field {
	return Field{Key: key, Type: BinaryType, Interface: val}
}

// Object creates a field with a complex object.
// The object will be marshaled by the backend (e.g., as JSON).
func Object(key string, val interface{}) Field {
	return Field{Key: key, Type: ObjectType, Interface: val}
}

// FormatValue formats the field value as a string for display purposes.
// This is primarily used for debugging and testing.
func (f Field) FormatValue() string {
	switch f.Type {
	case StringType:
		return f.String
	case Int64Type, Uint64Type:
		return fmt.Sprintf("%d", f.Integer)
	case Float64Type:
		return fmt.Sprintf("%f", f.Float)
	case BoolType:
		return fmt.Sprintf("%t", f.Integer == 1)
	case TimeType:
		if t, ok := f.Interface.(time.Time); ok {
			return t.Format(time.RFC3339Nano)
		}
		return fmt.Sprintf("%d", f.Integer)
	case DurationType:
		return time.Duration(f.Integer).String()
	case ErrorType:
		return f.String
	case AnyType, ArrayType, ObjectType:
		return fmt.Sprintf("%+v", f.Interface)
	case BinaryType:
		if b, ok := f.Interface.([]byte); ok {
			return fmt.Sprintf("%s", b)
		}
		return fmt.Sprintf("%v", f.Interface)
	default:
		return fmt.Sprintf("%v", f.Interface)
	}
}

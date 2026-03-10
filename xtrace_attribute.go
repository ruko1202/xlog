package xlog

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Attribute represents a key-value pair for tracing attributes.
// This is a universal attribute type that can be adapted to any tracing backend.
type Attribute struct {
	Key   string
	Type  FieldType // Reuse FieldType for consistency
	Value AttributeValue
}

// AttributeValue holds the actual value of an attribute.
type AttributeValue struct {
	String    string
	Integer   int64
	Float     float64
	Bool      bool
	Interface interface{}
}

// StringAttr creates a string attribute.
func StringAttr(key, val string) Attribute {
	return Attribute{
		Key:  key,
		Type: StringType,
		Value: AttributeValue{
			String: val,
		},
	}
}

// IntAttr creates an int attribute.
func IntAttr(key string, val int) Attribute {
	return Attribute{
		Key:  key,
		Type: Int64Type,
		Value: AttributeValue{
			Integer: int64(val),
		},
	}
}

// Int64Attr creates an int64 attribute.
func Int64Attr(key string, val int64) Attribute {
	return Attribute{
		Key:  key,
		Type: Int64Type,
		Value: AttributeValue{
			Integer: val,
		},
	}
}

// Float64Attr creates a float64 attribute.
func Float64Attr(key string, val float64) Attribute {
	return Attribute{
		Key:  key,
		Type: Float64Type,
		Value: AttributeValue{
			Float: val,
		},
	}
}

// BoolAttr creates a boolean attribute.
func BoolAttr(key string, val bool) Attribute {
	return Attribute{
		Key:  key,
		Type: BoolType,
		Value: AttributeValue{
			Bool: val,
		},
	}
}

// StringsAttr creates a string array attribute.
func StringsAttr(key string, val []string) Attribute {
	return Attribute{
		Key:  key,
		Type: ArrayType,
		Value: AttributeValue{
			Interface: val,
		},
	}
}

// IntsAttr creates an int array attribute.
func IntsAttr(key string, val []int) Attribute {
	return Attribute{
		Key:  key,
		Type: ArrayType,
		Value: AttributeValue{
			Interface: val,
		},
	}
}

// fieldsToOtelAttributes converts xlog Fields directly to OpenTelemetry attributes.
// This is a convenience function that combines FieldsToAttributes and ToOtelAttribute.
func fieldsToOtelAttributes(fields []Field) []attribute.KeyValue {
	attrs := fieldsToAttributes(fields)
	if len(attrs) == 0 {
		return nil
	}

	otelAttrs := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		otelAttrs = append(otelAttrs, toOtelAttribute(attr))
	}

	return otelAttrs
}

// toOtelAttribute converts xlog.Attribute to OpenTelemetry attribute.KeyValue.
func toOtelAttribute(a Attribute) attribute.KeyValue {
	switch a.Type {
	case StringType:
		return attribute.String(a.Key, a.Value.String)
	case Int64Type:
		return attribute.Int64(a.Key, a.Value.Integer)
	case Uint64Type:
		return attribute.Int64(a.Key, a.Value.Integer)
	case Float64Type:
		return attribute.Float64(a.Key, a.Value.Float)
	case BoolType:
		return attribute.Bool(a.Key, a.Value.Bool)
	case ArrayType:
		// Handle array types
		switch v := a.Value.Interface.(type) {
		case []string:
			return attribute.StringSlice(a.Key, v)
		case []int:
			return attribute.IntSlice(a.Key, v)
		case []int64:
			return attribute.Int64Slice(a.Key, v)
		case []float64:
			return attribute.Float64Slice(a.Key, v)
		case []bool:
			return attribute.BoolSlice(a.Key, v)
		default:
			// Fallback to string representation
			return attribute.String(a.Key, "unsupported array type")
		}
	default:
		// Fallback: convert to string
		return attribute.String(a.Key, a.Value.String)
	}
}

// fieldsToAttributes converts a slice of Fields to Attributes.
// Only supported field types are converted; unsupported types are skipped.
func fieldsToAttributes(fields []Field) []Attribute {
	if len(fields) == 0 {
		return nil
	}

	// Pre-count supported fields
	count := 0
	for _, f := range fields {
		if isFieldTypeConvertibleToAttribute(f.Type) {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	attrs := make([]Attribute, 0, count)
	for _, f := range fields {
		if isFieldTypeConvertibleToAttribute(f.Type) {
			attrs = append(attrs, fieldToAttribute(f))
		}
	}

	return attrs
}

// fieldToAttribute converts xlog.Field to xlog.Attribute.
// This is useful when you want to add log fields as span attributes.
func fieldToAttribute(f Field) Attribute {
	switch f.Type {
	case StringType:
		return StringAttr(f.Key, f.String)
	case Int64Type, Uint64Type:
		return Int64Attr(f.Key, f.Integer)
	case Float64Type:
		return Float64Attr(f.Key, f.Float)
	case BoolType:
		return BoolAttr(f.Key, f.Integer == 1)
	case TimeType:
		if t, ok := f.Interface.(time.Time); ok {
			return StringAttr(f.Key, t.Format(time.RFC3339Nano))
		}
		return Int64Attr(f.Key, f.Integer)
	case DurationType:
		return StringAttr(f.Key, time.Duration(f.Integer).String())
	case ErrorType:
		return StringAttr(f.Key, f.String)
	case ArrayType:
		// Try to preserve array types
		return Attribute{
			Key:  f.Key,
			Type: ArrayType,
			Value: AttributeValue{
				Interface: f.Interface,
			},
		}
	default:
		// Fallback: any type as string
		return StringAttr(f.Key, f.FormatValue())
	}
}

// isFieldTypeConvertibleToAttribute checks if a field type can be converted to an attribute.
func isFieldTypeConvertibleToAttribute(t FieldType) bool {
	switch t {
	case StringType, Int64Type, Uint64Type, Float64Type, BoolType,
		TimeType, DurationType, ErrorType, ArrayType:
		return true
	default:
		return false
	}
}

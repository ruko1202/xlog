package xlog

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// fieldsToOtelAttributes converts xlog Fields directly to OpenTelemetry attributes.
func fieldsToOtelAttributes(fields []Field) []attribute.KeyValue {
	if len(fields) == 0 {
		return nil
	}

	// Pre-count supported fields to avoid reallocation
	count := 0
	for i := range fields {
		if isConvertible(fields[i].Type) {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	otelAttrs := make([]attribute.KeyValue, 0, count)
	for i := range fields {
		if isConvertible(fields[i].Type) {
			otelAttrs = append(otelAttrs, fieldToOtelAttribute(fields[i]))
		}
	}

	return otelAttrs
}

// isFieldTypeConvertibleToAttribute checks if a field type can be converted to an attribute.
func isConvertible(t FieldType) bool {
	switch t {
	case StringType, Int64Type, Uint64Type, Float64Type, BoolType,
		TimeType, DurationType, ErrorType, ArrayType:
		return true
	default:
		return false
	}
}

// fieldToOtelAttribute converts a Field to OpenTelemetry attribute.KeyValue.
//
//nolint:gocyclo // switch on field types requires many cases
func fieldToOtelAttribute(f Field) attribute.KeyValue {
	switch f.Type {
	case StringType:
		return attribute.String(f.Key, f.String)
	case Int64Type, Uint64Type:
		return attribute.Int64(f.Key, f.Integer)
	case Float64Type:
		return attribute.Float64(f.Key, f.Float)
	case BoolType:
		return attribute.Bool(f.Key, f.Integer == 1)
	case TimeType:
		if t, ok := f.Interface.(time.Time); ok {
			return attribute.String(f.Key, t.Format(time.RFC3339Nano))
		}
		return attribute.Int64(f.Key, f.Integer)
	case DurationType:
		return attribute.String(f.Key, time.Duration(f.Integer).String())
	case ErrorType:
		if err, ok := f.Interface.(error); ok && err != nil {
			return attribute.String(f.Key, err.Error())
		}
		return attribute.String(f.Key, f.String)
	case ArrayType:
		// Handle array types
		switch v := f.Interface.(type) {
		case []string:
			return attribute.StringSlice(f.Key, v)
		case []int:
			return attribute.IntSlice(f.Key, v)
		case []int64:
			return attribute.Int64Slice(f.Key, v)
		case []float64:
			return attribute.Float64Slice(f.Key, v)
		case []bool:
			return attribute.BoolSlice(f.Key, v)
		default:
			// Fallback to string representation
			return attribute.String(f.Key, "unsupported array type")
		}
	default:
		// Fallback: convert to string
		return attribute.String(f.Key, f.FormatValue())
	}
}

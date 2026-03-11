package xlog

import (
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ruko1202/xlog/xfield"
)

// fieldsToOtelAttributes converts xlog Fields directly to OpenTelemetry attributes.
func fieldsToOtelAttributes(fields []xfield.Field) []attribute.KeyValue {
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
func isConvertible(t xfield.FieldType) bool {
	switch t {
	case xfield.StringType,
		xfield.Int64Type,
		xfield.Uint64Type,
		xfield.Float64Type,
		xfield.BoolType,
		xfield.TimeType,
		xfield.DurationType,
		xfield.ErrorType,
		xfield.ArrayType:
		return true
	default:
		return false
	}
}

// fieldToOtelAttribute converts a Field to OpenTelemetry attribute.KeyValue.
//
//nolint:gocyclo // switch on field types requires many cases
func fieldToOtelAttribute(f xfield.Field) attribute.KeyValue {
	switch f.Type {
	case xfield.StringType:
		return attribute.String(f.Key, f.String)
	case xfield.Int64Type, xfield.Uint64Type:
		return attribute.Int64(f.Key, f.Integer)
	case xfield.Float64Type:
		return attribute.Float64(f.Key, f.Float)
	case xfield.BoolType:
		return attribute.Bool(f.Key, f.Integer == 1)
	case xfield.TimeType:
		if t, ok := f.Interface.(time.Time); ok {
			return attribute.String(f.Key, t.Format(time.RFC3339Nano))
		}
		return attribute.Int64(f.Key, f.Integer)
	case xfield.DurationType:
		return attribute.String(f.Key, time.Duration(f.Integer).String())
	case xfield.ErrorType:
		if err, ok := f.Interface.(error); ok && err != nil {
			return attribute.String(f.Key, err.Error())
		}
		return attribute.String(f.Key, f.String)
	case xfield.ArrayType:
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

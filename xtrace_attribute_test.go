package xlog

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ruko1202/xlog/xfield"
)

func TestConvertFieldsToAttributes(t *testing.T) {
	t.Run("converts string fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.String("key1", "value1"),
			xfield.String("key2", "value2"),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.String("key1", "value1"), attrs[0])
		assert.Equal(t, attribute.String("key2", "value2"), attrs[1])
	})

	t.Run("converts integer fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Int("int", 42),
			xfield.Int64("int64", 9999),
			xfield.Int32("int32", 100),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 3, len(attrs))
		assert.Equal(t, attribute.Int64("int", 42), attrs[0])
		assert.Equal(t, attribute.Int64("int64", 9999), attrs[1])
		assert.Equal(t, attribute.Int64("int32", 100), attrs[2])
	})

	t.Run("converts bool fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Bool("bool1", true),
			xfield.Bool("bool2", false),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.Bool("bool1", true), attrs[0])
		assert.Equal(t, attribute.Bool("bool2", false), attrs[1])
	})

	t.Run("converts float fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Float64("float64", 3.14),
			xfield.Float64("another_float", 2.71),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.Key("float64"), attrs[0].Key)
		assert.InDelta(t, 3.14, attrs[0].Value.AsFloat64(), 0.001)
		assert.Equal(t, attribute.Key("another_float"), attrs[1].Key)
		assert.InDelta(t, 2.71, attrs[1].Value.AsFloat64(), 0.001)
	})

	t.Run("handles empty fields", func(t *testing.T) {
		fields := []xfield.Field{}

		attrs := fieldsToOtelAttributes(fields)
		assert.Nil(t, attrs)
	})

	t.Run("returns nil when all fields are unsupported", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Any("any1", map[string]string{"key": "value"}),
			xfield.Any("any2", []string{"a", "b", "c"}),
		}

		attrs := fieldsToOtelAttributes(fields)
		assert.Nil(t, attrs)
	})

	t.Run("ignores complex types", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.String("string", "value"),
			xfield.Any("any", map[string]string{"key": "value"}),
			xfield.Int("int", 42),
		}

		attrs := fieldsToOtelAttributes(fields)
		// Should only have string and int, Any is ignored
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.String("string", "value"), attrs[0])
		assert.Equal(t, attribute.Int64("int", 42), attrs[1])
	})

	t.Run("handles mixed field types", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.String("name", "test"),
			xfield.Int("count", 10),
			xfield.Bool("active", true),
			xfield.Float64("rate", 0.95),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 4, len(attrs))
		assert.Equal(t, attribute.String("name", "test"), attrs[0])
		assert.Equal(t, attribute.Int64("count", 10), attrs[1])
		assert.Equal(t, attribute.Bool("active", true), attrs[2])
		assert.InDelta(t, 0.95, attrs[3].Value.AsFloat64(), 0.001)
	})

	t.Run("converts time fields", func(t *testing.T) {
		now := time.Now()
		fields := []xfield.Field{
			xfield.Time("timestamp", now),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.Key("timestamp"), attrs[0].Key)
		assert.Equal(t, now.Format(time.RFC3339Nano), attrs[0].Value.AsString())
	})

	t.Run("converts duration fields", func(t *testing.T) {
		dur := 5 * time.Second
		fields := []xfield.Field{
			xfield.Duration("elapsed", dur),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.String("elapsed", "5s"), attrs[0])
	})

	t.Run("converts error fields", func(t *testing.T) {
		testErr := errors.New("test error")
		fields := []xfield.Field{
			xfield.Error(testErr),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.String("error", "test error"), attrs[0])
	})

	t.Run("converts uint64 fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Uint64("counter", 18446744073709551615),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.Key("counter"), attrs[0].Key)
		// Uint64 is stored as int64 in the field
		assert.Equal(t, int64(-1), attrs[0].Value.AsInt64())
	})

	t.Run("converts string array fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Strings("tags", []string{"tag1", "tag2", "tag3"}),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.Key("tags"), attrs[0].Key)
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, attrs[0].Value.AsStringSlice())
	})

	t.Run("converts int array fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Ints("numbers", []int{1, 2, 3}),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.Key("numbers"), attrs[0].Key)
		// OTel stores as []int64, but we can compare the slice
		slice := attrs[0].Value.AsInt64Slice()
		require.Equal(t, 3, len(slice))
		assert.Equal(t, int64(1), slice[0])
		assert.Equal(t, int64(2), slice[1])
		assert.Equal(t, int64(3), slice[2])
	})

	t.Run("converts int64 array fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Int64s("ids", []int64{100, 200, 300}),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.Key("ids"), attrs[0].Key)
		assert.Equal(t, []int64{100, 200, 300}, attrs[0].Value.AsInt64Slice())
	})

	t.Run("converts float64 array fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Float64s("rates", []float64{1.1, 2.2, 3.3}),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.Key("rates"), attrs[0].Key)
		slice := attrs[0].Value.AsFloat64Slice()
		require.Equal(t, 3, len(slice))
		assert.InDelta(t, 1.1, slice[0], 0.001)
		assert.InDelta(t, 2.2, slice[1], 0.001)
		assert.InDelta(t, 3.3, slice[2], 0.001)
	})

	t.Run("converts bool array fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.Bools("flags", []bool{true, false, true}),
		}

		attrs := fieldsToOtelAttributes(fields)
		require.Equal(t, 1, len(attrs))
		assert.Equal(t, attribute.Key("flags"), attrs[0].Key)
		assert.Equal(t, []bool{true, false, true}, attrs[0].Value.AsBoolSlice())
	})
}

func TestIsFieldTypeConvertibleToAttribute(t *testing.T) {
	t.Run("supported types", func(t *testing.T) {
		supportedTypes := []xfield.FieldType{
			xfield.StringType,
			xfield.Int64Type,
			xfield.Uint64Type,
			xfield.BoolType,
			xfield.Float64Type,
			xfield.TimeType,
			xfield.DurationType,
			xfield.ErrorType,
			xfield.ArrayType,
		}

		for _, ft := range supportedTypes {
			assert.True(t, isConvertible(ft), "type %v should be supported", ft)
		}
	})

	t.Run("unsupported types", func(t *testing.T) {
		unsupportedTypes := []xfield.FieldType{
			xfield.BinaryType,
			xfield.ObjectType,
			xfield.AnyType,
		}

		for _, ft := range unsupportedTypes {
			assert.False(t, isConvertible(ft), "type %v should NOT be supported", ft)
		}
	})
}

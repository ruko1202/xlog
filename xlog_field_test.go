package xlog

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFieldCreation(t *testing.T) {
	t.Run("String field", func(t *testing.T) {
		f := String("key", "value")
		assert.Equal(t, "key", f.Key)
		assert.Equal(t, StringType, f.Type)
		assert.Equal(t, "value", f.String)
		assert.Equal(t, "value", f.FormatValue())
	})

	t.Run("Int field", func(t *testing.T) {
		f := Int("key", 42)
		assert.Equal(t, "key", f.Key)
		assert.Equal(t, Int64Type, f.Type)
		assert.Equal(t, int64(42), f.Integer)
		assert.Equal(t, "42", f.FormatValue())
	})

	t.Run("Int32 field", func(t *testing.T) {
		f := Int32("key", 123456789)
		assert.Equal(t, Int64Type, f.Type)
		assert.Equal(t, int64(123456789), f.Integer)
		assert.Equal(t, "123456789", f.FormatValue())
	})

	t.Run("Int64 field", func(t *testing.T) {
		f := Int64("key", 123456789)
		assert.Equal(t, Int64Type, f.Type)
		assert.Equal(t, int64(123456789), f.Integer)
		assert.Equal(t, "123456789", f.FormatValue())
	})

	t.Run("Uint field", func(t *testing.T) {
		f := Uint("key", 987654321)
		assert.Equal(t, Uint64Type, f.Type)
		assert.Equal(t, int64(987654321), f.Integer)
		assert.Equal(t, "987654321", f.FormatValue())
	})

	t.Run("Uint32 field", func(t *testing.T) {
		f := Uint32("key", 987654321)
		assert.Equal(t, Uint64Type, f.Type)
		assert.Equal(t, int64(987654321), f.Integer)
		assert.Equal(t, "987654321", f.FormatValue())
	})

	t.Run("Uint64 field", func(t *testing.T) {
		f := Uint64("key", 987654321)
		assert.Equal(t, Uint64Type, f.Type)
		assert.Equal(t, int64(987654321), f.Integer)
		assert.Equal(t, "987654321", f.FormatValue())
	})

	t.Run("Float32 field", func(t *testing.T) {
		f := Float32("key", 3.14)
		assert.Equal(t, Float64Type, f.Type)
		assert.InDelta(t, 3.14, f.Float, 0.001)
		assert.Equal(t, "3.140000", f.FormatValue())
	})

	t.Run("Float64 field", func(t *testing.T) {
		f := Float64("key", 3.14)
		assert.Equal(t, Float64Type, f.Type)
		assert.InDelta(t, 3.14, f.Float, 0.001)
		assert.Equal(t, "3.140000", f.FormatValue())
	})

	t.Run("Bool field true", func(t *testing.T) {
		f := Bool("key", true)
		assert.Equal(t, BoolType, f.Type)
		assert.Equal(t, int64(1), f.Integer)
		assert.Equal(t, "true", f.FormatValue())
	})

	t.Run("Bool field false", func(t *testing.T) {
		f := Bool("key", false)
		assert.Equal(t, BoolType, f.Type)
		assert.Equal(t, int64(0), f.Integer)
		assert.Equal(t, "false", f.FormatValue())
	})

	t.Run("Bools field", func(t *testing.T) {
		f := Bools("key", []bool{true, false})
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, int64(0), f.Integer)
		assert.Equal(t, "[true false]", f.FormatValue())
	})

	t.Run("Time field", func(t *testing.T) {
		now := time.Now()
		f := Time("key", now)
		assert.Equal(t, TimeType, f.Type)
		assert.Equal(t, now.UnixNano(), f.Integer)
		assert.Equal(t, now, f.Interface)
		assert.Equal(t, now.Format(time.RFC3339Nano), f.FormatValue())
	})

	t.Run("Duration field", func(t *testing.T) {
		d := 5 * time.Second
		f := Duration("key", d)
		assert.Equal(t, DurationType, f.Type)
		assert.Equal(t, int64(d), f.Integer)
		assert.Equal(t, "5s", f.FormatValue())
	})

	t.Run("Durations field", func(t *testing.T) {
		f := Durations("key", []time.Duration{time.Millisecond, time.Second})
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, "[1ms 1s]", f.FormatValue())
	})

	t.Run("Err field", func(t *testing.T) {
		err := errors.New("test error")
		f := Err(err)
		assert.Equal(t, "error", f.Key)
		assert.Equal(t, ErrorType, f.Type)
		assert.Equal(t, "", f.String)
		assert.Equal(t, err, f.Interface)
		assert.Equal(t, err.Error(), f.FormatValue())
	})

	t.Run("Err field nil", func(t *testing.T) {
		f := Err(nil)
		assert.Equal(t, ErrorType, f.Type)
		assert.Nil(t, f.Interface)
		assert.Equal(t, "", f.FormatValue())
	})

	t.Run("NamedError field", func(t *testing.T) {
		err := errors.New("custom error")
		f := NamedError("custom_key", err)
		assert.Equal(t, "custom_key", f.Key)
		assert.Equal(t, ErrorType, f.Type)
		assert.Equal(t, "custom error", f.String)
		assert.Equal(t, err.Error(), f.FormatValue())
	})

	t.Run("NamedError field with nil", func(t *testing.T) {
		f := NamedError("custom_key", nil)
		assert.Equal(t, "custom_key", f.Key)
		assert.Equal(t, ErrorType, f.Type)
		assert.Equal(t, "", f.String)
		assert.Equal(t, "", f.FormatValue())
	})

	t.Run("Any field", func(t *testing.T) {
		type custom struct {
			Value string
		}
		obj := custom{Value: "test"}
		f := Any("key", obj)
		assert.Equal(t, AnyType, f.Type)
		assert.Equal(t, obj, f.Interface)
		assert.Equal(t, "{Value:test}", f.FormatValue())
	})

	t.Run("Strings field", func(t *testing.T) {
		vals := []string{"a", "b", "c"}
		f := Strings("key", vals)
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, vals, f.Interface)
		assert.Equal(t, "[a b c]", f.FormatValue())
	})

	t.Run("Ints field", func(t *testing.T) {
		vals := []int{1, 2, 3}
		f := Ints("key", vals)
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, vals, f.Interface)
		assert.Equal(t, "[1 2 3]", f.FormatValue())
	})

	t.Run("Int32s field", func(t *testing.T) {
		vals := []int32{1, 2, 3}
		f := Int32s("key", vals)
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, vals, f.Interface)
		assert.Equal(t, "[1 2 3]", f.FormatValue())
	})

	t.Run("Int64s field", func(t *testing.T) {
		vals := []int64{1, 2, 3}
		f := Int64s("key", vals)
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, vals, f.Interface)
		assert.Equal(t, "[1 2 3]", f.FormatValue())
	})

	t.Run("UInts field", func(t *testing.T) {
		vals := []uint{1, 2, 3}
		f := UInts("key", vals)
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, vals, f.Interface)
		assert.Equal(t, "[1 2 3]", f.FormatValue())
	})

	t.Run("UInt32s field", func(t *testing.T) {
		vals := []uint32{1, 2, 3}
		f := UInt32s("key", vals)
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, vals, f.Interface)
		assert.Equal(t, "[1 2 3]", f.FormatValue())
	})

	t.Run("UInt64s field", func(t *testing.T) {
		vals := []uint64{1, 2, 3}
		f := UInt64s("key", vals)
		assert.Equal(t, ArrayType, f.Type)
		assert.Equal(t, vals, f.Interface)
		assert.Equal(t, "[1 2 3]", f.FormatValue())
	})

	t.Run("Binary field", func(t *testing.T) {
		data := []byte("some bytes")
		f := Binary("key", data)
		assert.Equal(t, BinaryType, f.Type)
		assert.Equal(t, data, f.Interface)
		assert.Equal(t, "some bytes", f.FormatValue())
	})

	t.Run("Object field", func(t *testing.T) {
		type custom struct {
			Name string
			Age  int
		}
		obj := custom{Name: "test", Age: 30}
		f := Object("key", obj)
		assert.Equal(t, ObjectType, f.Type)
		assert.Equal(t, obj, f.Interface)
		assert.Equal(t, "{Name:test Age:30}", f.FormatValue())
	})
}

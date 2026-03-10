package xlog

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFieldToZapField(t *testing.T) {
	t.Run("converts all basic types", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			field Field
		}{
			{
				name:  "String",
				field: String("key", "value"),
			}, {
				name:  "Int64",
				field: Int64("key", 123),
			}, {
				name:  "Uint64",
				field: Uint64("key", 456),
			}, {
				name:  "Float64",
				field: Float64("key", 3.14),
			}, {
				name:  "Bool",
				field: Bool("key", true),
			}, {
				name:  "Time",
				field: Time("key", time.Now()),
			}, {
				name:  "Duration",
				field: Duration("key", 5*time.Second),
			}, {
				name:  "Error",
				field: Err(errors.New("test")),
			}, {
				name:  "Binary",
				field: Binary("key", []byte("a")),
			}, {
				name:  "Any",
				field: Any("key", []byte("a")),
			}, {
				name:  "Any",
				field: Object("key", []byte("a")),
			}, {
				name:  "Strings",
				field: Strings("key", []string{"a", "b", "c"}),
			}, {
				name:  "Ints",
				field: Ints("key", []int{1, 2, 3}),
			}, {
				name:  "Int32s",
				field: Int32s("key", []int32{1, 2, 3}),
			}, {
				name:  "Int64s",
				field: Int64s("key", []int64{1, 2, 3}),
			}, {
				name:  "UInts",
				field: UInts("key", []uint{1, 2, 3}),
			}, {
				name:  "UInt32s",
				field: UInt32s("key", []uint32{1, 2, 3}),
			}, {
				name:  "UInt64s",
				field: UInt64s("key", []uint64{1, 2, 3}),
			}, {
				name:  "Float32s",
				field: Float32s("key", []float32{1.1, 2.2, 3.3}),
			}, {
				name:  "Float64s",
				field: Float64s("key", []float64{1.1, 2.2, 3.3}),
			}, {
				name:  "Bools",
				field: Bools("key", []bool{true, false}),
			}, {
				name:  "Bools",
				field: Durations("key", []time.Duration{time.Second, time.Millisecond}),
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				zapField := fieldToZapField(tc.field)
				assert.Equal(t, tc.field.Key, zapField.Key)
				assert.Equal(t, tc.field.String, zapField.String)
			})
		}
	})
}

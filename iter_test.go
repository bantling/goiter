package goiter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArraySliceIterFunc(t *testing.T) {
	// Empty array
	iterFunc := ArraySliceIterFunc(reflect.ValueOf([0]int{}))

	_, next := iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element array
	iterFunc = ArraySliceIterFunc(reflect.ValueOf([1]int{1}))

	val, next := iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Two element array
	iterFunc = ArraySliceIterFunc(reflect.ValueOf([2]int{1, 2}))

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 2, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty slice
	iterFunc = ArraySliceIterFunc(reflect.ValueOf([]int{}))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element slice
	iterFunc = ArraySliceIterFunc(reflect.ValueOf([]int{1}))

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Two element slice
	iterFunc = ArraySliceIterFunc(reflect.ValueOf([]int{1, 2}))

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 2, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	iterFunc = ArraySliceIterFunc(reflect.ValueOf([]interface{}{3, 4}))

	val, next = iterFunc()
	assert.Equal(t, 3, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 4, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Non-array/slice
	func() {
		defer func() {
			assert.Equal(t, "ArraySliceIterFunc argument must be an array or slice", recover())
		}()

		ArraySliceIterFunc(reflect.ValueOf(1))

		assert.Fail(t, "Must panic on non-array/slice")
	}()
}

func TestIterablesFunc(t *testing.T) {
	// No Iterables
	iterFunc := IterablesFunc([]Iterable{})

	_, next := iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One nil Iterable
	iterFunc = IterablesFunc([]Iterable{nil})

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One empty Iterable
	iterFunc = IterablesFunc([]Iterable{Of()})

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One Iterable of one value
	iterFunc = IterablesFunc([]Iterable{Of(1)})

	val, next := iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One Iterable of two values
	iterFunc = IterablesFunc([]Iterable{Of(1, 2)})

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 2, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Mix of nil, empty, non-empty Iterables
	iterFunc = IterablesFunc([]Iterable{nil, Of(), Of(1), nil, Of(2), Of(), Of(3), Of(4, 5)})

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 2, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 3, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 4, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 5, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)
}

func TestMapIterFunc(t *testing.T) {
	// Empty map
	iterFunc := MapIterFunc(reflect.ValueOf(map[int]int{}))

	_, next := iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element map
	iterFunc = MapIterFunc(reflect.ValueOf(map[int]int{1: 2}))

	val, next := iterFunc()
	assert.Equal(t, KeyValue{Key: 1, Value: 2}, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Two element map
	expected := map[int]int{1: 2, 3: 4}
	iterFunc = MapIterFunc(reflect.ValueOf(expected))
	m := map[int]int{}

	val, next = iterFunc()
	kv := val.(KeyValue)
	m[kv.Key.(int)] = kv.Value.(int)
	assert.True(t, next)

	val, next = iterFunc()
	kv = val.(KeyValue)
	m[kv.Key.(int)] = kv.Value.(int)
	assert.True(t, next)

	assert.Equal(t, expected, m)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Non-map
	func() {
		defer func() {
			assert.Equal(t, "MapIterFunc argument must be a map", recover())
		}()

		MapIterFunc(reflect.ValueOf(1))

		assert.Fail(t, "Must panic on non-map")
	}()
}

func TestNoValueIterFunc(t *testing.T) {
	iterFunc := NoValueIterFunc

	_, next := iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)
}

func TestSingleValueIterFunc(t *testing.T) {
	// One element
	iterFunc := SingleValueIterFunc(reflect.ValueOf(5))

	val, next := iterFunc()
	assert.Equal(t, 5, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)
}

func TestElementsIterFunc(t *testing.T) {
	// Empty array
	iterFunc := ElementsIterFunc(reflect.ValueOf([0]int{}))

	_, next := iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty slice
	iterFunc = ElementsIterFunc(reflect.ValueOf([]int{}))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element array
	iterFunc = ElementsIterFunc(reflect.ValueOf([1]int{1}))

	val, next := iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty slice
	iterFunc = ElementsIterFunc(reflect.ValueOf([]int{}))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element slice
	iterFunc = ElementsIterFunc(reflect.ValueOf([]int{1}))

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty Iterable
	iterFunc = ElementsIterFunc(reflect.ValueOf(Of()))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element Iterable
	iterFunc = ElementsIterFunc(reflect.ValueOf(Of(1)))

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty map
	iterFunc = ElementsIterFunc(reflect.ValueOf(map[int]int{}))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element map
	iterFunc = ElementsIterFunc(reflect.ValueOf(map[int]int{1: 2}))

	val, next = iterFunc()
	assert.Equal(t, KeyValue{Key: 1, Value: 2}, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Nil ptr
	iterFunc = ElementsIterFunc(reflect.ValueOf((*int)(nil)))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Single value
	iterFunc = ElementsIterFunc(reflect.ValueOf(5))

	val, next = iterFunc()
	assert.Equal(t, 5, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)
}

func TestDelayedIterFunc(t *testing.T) {
	iterFunc := DelayedIterFunc(func() func() (interface{}, bool) {
		return SingleValueIterFunc(reflect.ValueOf(1))
	},
	)

	val, next := iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)
}

func TestOf(t *testing.T) {
	// Empty items
	iter := Of()

	next := iter.Next()
	assert.False(t, next)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	// One item
	iter = Of(5)

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 5, iter.Value())

	next = iter.Next()
	assert.False(t, next)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	// Two items
	iter = Of(5, []int{6, 7})

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 5, iter.Value())

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, []int{6, 7}, iter.Value())

	next = iter.Next()
	assert.False(t, next)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()
}

func TestOfElements(t *testing.T) {
	// Slice
	iter := OfElements([]int{5, 6})

	next := iter.Next()
	assert.True(t, next)
	assert.Equal(t, 5, iter.Value())

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 6, iter.Value())

	next = iter.Next()
	assert.False(t, next)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	// Iterable
	iter = OfElements(Of(1))

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 1, iter.Value())

	next = iter.Next()
	assert.False(t, next)

	// Map
	iter = OfElements(map[int]int{1: 2})

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, KeyValue{1, 2}, iter.Value())

	next = iter.Next()
	assert.False(t, next)

	// Nil
	iter = OfElements(nil)

	next = iter.Next()
	assert.False(t, next)

	// One item
	iter = OfElements(5)

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 5, iter.Value())

	next = iter.Next()
	assert.False(t, next)
}

func TestOfIterables(t *testing.T) {
	// Empty iterables
	iter := OfIterables()

	next := iter.Next()
	assert.False(t, next)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	// One iterable
	iter = OfIterables(Of(5))

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 5, iter.Value())

	next = iter.Next()
	assert.False(t, next)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	// Mix of nil, empty, non-empty Iterables
	iter = OfIterables(nil, Of(5), Of(), Of(6))

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 5, iter.Value())

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 6, iter.Value())

	next = iter.Next()
	assert.False(t, next)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()
}

func TestBoolValue(t *testing.T) {
	var (
		v1   = bool(true)
		iter = Of(v1)
	)

	next := iter.Next()
	assert.True(t, next)
	var v2 bool = iter.BoolValue()
	assert.Equal(t, bool(true), v2)
}

func TestComplexValue(t *testing.T) {
	var (
		v1   = complex128(1 + 2i)
		iter = Of(v1)
	)

	next := iter.Next()
	assert.True(t, next)
	var v2 complex128 = iter.ComplexValue()
	assert.Equal(t, complex128(1+2i), v2)
}

func TestFloatValue(t *testing.T) {
	var (
		v1   = float64(1.25)
		iter = Of(v1)
	)

	next := iter.Next()
	assert.True(t, next)
	var v2 float64 = iter.FloatValue()
	assert.Equal(t, float64(1.25), v2)
}

func TestIntValue(t *testing.T) {
	var (
		v1   = int64(1)
		iter = Of(v1)
	)

	next := iter.Next()
	assert.True(t, next)
	var v2 int64 = iter.IntValue()
	assert.Equal(t, int64(1), v2)
}

func TestUintValue(t *testing.T) {
	var (
		v1   = uint64(1)
		iter = Of(v1)
	)

	next := iter.Next()
	assert.True(t, next)
	var v2 uint64 = iter.UintValue()
	assert.Equal(t, uint64(1), v2)
}

func TestStringValue(t *testing.T) {
	var (
		v1   = "1"
		iter = Of(v1)
	)

	next := iter.Next()
	assert.True(t, next)
	var v2 string = iter.StringValue()
	assert.Equal(t, "1", v2)
}

func TestValueOfType(t *testing.T) {
	var (
		v1   = "1"
		iter = Of(v1)
	)

	next := iter.Next()
	assert.True(t, next)
	var v2 string = iter.ValueOfType("").(string)
	assert.Equal(t, "1", v2)
}

func TestForLoop(t *testing.T) {
	func() {
		var (
			iter     = Of(5, []int{6, 7})
			idx      = 0
			expected = []interface{}{5, []int{6, 7}}
		)

		for iter.Next() {
			assert.Equal(t, expected[idx], iter.Value())
			idx++
		}

		assert.Equal(t, 2, idx)

		func() {
			defer func() {
				assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
			}()

			iter.Next()
			assert.Fail(t, "Must panic")
		}()
	}()

	func() {
		var (
			iter     *Iter
			idx      = 0
			expected = []int{5}
		)

		for iter = OfElements(5); iter.Next(); {
			assert.Equal(t, expected[idx], iter.Value())
			idx++
		}

		assert.Equal(t, 1, idx)

		func() {
			defer func() {
				assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
			}()

			iter.Next()
			assert.Fail(t, "Must panic")
		}()
	}()
}

func TestIterIsIterable(t *testing.T) {
	var (
		iter     = Of(0)
		iterable = Iterable(iter)
		it       = iterable.Iter()
	)
	assert.True(t, it == iter)
	assert.True(t, it.Next())
	assert.Equal(t, 0, it.Value())
	assert.False(t, it.Next())

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		it.Next()
		assert.Fail(t, "Must panic")
	}()

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()
}

func TestSplitIntoRows(t *testing.T) {
	// Split with n = 5 items per subslice
	var (
		iter  = Of()
		split = iter.SplitIntoRows(5)
	)
	assert.Equal(t, [][]interface{}{}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1)
	split = iter.SplitIntoRows(5)
	assert.Equal(t, [][]interface{}{{1}}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1, 2, 3, 4)
	split = iter.SplitIntoRows(5)
	assert.Equal(t, [][]interface{}{{1, 2, 3, 4}}, split)

	iter = Of(1, 2, 3, 4, 5)
	split = iter.SplitIntoRows(5)
	assert.Equal(t, [][]interface{}{{1, 2, 3, 4, 5}}, split)

	iter = Of(1, 2, 3, 4, 5, 6)
	split = iter.SplitIntoRows(5)
	assert.Equal(t, [][]interface{}{{1, 2, 3, 4, 5}, {6}}, split)

	iter = Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	split = iter.SplitIntoRows(5)
	assert.Equal(t, [][]interface{}{{1, 2, 3, 4, 5}, {6, 7, 8, 9, 10}}, split)

	iter = Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	split = iter.SplitIntoRows(5)
	assert.Equal(t, [][]interface{}{{1, 2, 3, 4, 5}, {6, 7, 8, 9, 10}, {11}}, split)

	// Split with n = 1 items per subslice corner case
	iter = Of()
	split = iter.SplitIntoRows(1)
	assert.Equal(t, [][]interface{}{}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1)
	split = iter.SplitIntoRows(1)
	assert.Equal(t, [][]interface{}{{1}}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1, 2)
	split = iter.SplitIntoRows(1)
	assert.Equal(t, [][]interface{}{{1}, {2}}, split)

	// Die if n < 1
	func() {
		defer func() {
			assert.Equal(t, "cols must be > 0", recover())
		}()

		iter.SplitIntoRows(0)
		assert.Fail(t, "Must panic")
	}()
}

func TestSplitIntoRowsOf(t *testing.T) {
	// Split with n = 5 items per subslice
	var (
		iter  = Of()
		split = iter.SplitIntoRowsOf(5, 0)
	)
	assert.Equal(t, [][]int{}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1)
	split = iter.SplitIntoRowsOf(5, 0)
	assert.Equal(t, [][]int{{1}}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1, 2, 3, 4)
	split = iter.SplitIntoRowsOf(5, 0)
	assert.Equal(t, [][]int{{1, 2, 3, 4}}, split)

	iter = Of(1, 2, 3, 4, 5)
	split = iter.SplitIntoRowsOf(5, 0)
	assert.Equal(t, [][]int{{1, 2, 3, 4, 5}}, split)

	iter = Of(1, 2, 3, 4, 5, 6)
	split = iter.SplitIntoRowsOf(5, 0)
	assert.Equal(t, [][]int{{1, 2, 3, 4, 5}, {6}}, split)

	iter = Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	split = iter.SplitIntoRowsOf(5, 0)
	assert.Equal(t, [][]int{{1, 2, 3, 4, 5}, {6, 7, 8, 9, 10}}, split)

	iter = Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	split = iter.SplitIntoRowsOf(5, 0)
	assert.Equal(t, [][]int{{1, 2, 3, 4, 5}, {6, 7, 8, 9, 10}, {11}}, split)

	// Split into a type that requires conversion
	iter = Of(uint(1), uint(2))
	split = iter.SplitIntoRowsOf(5, 0)
	assert.Equal(t, [][]int{{1, 2}}, split)

	// Split with n = 1 items per subslice corner case
	iter = Of()
	split = iter.SplitIntoRowsOf(1, 0)
	assert.Equal(t, [][]int{}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1)
	split = iter.SplitIntoRowsOf(1, 0)
	assert.Equal(t, [][]int{{1}}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1, 2)
	split = iter.SplitIntoRowsOf(1, 0)
	assert.Equal(t, [][]int{{1}, {2}}, split)

	// Die if n < 1
	func() {
		defer func() {
			assert.Equal(t, "cols must be > 0", recover())
		}()

		iter.SplitIntoRowsOf(0, 0)
		assert.Fail(t, "Must panic")
	}()

	// Die if value is nil
	func() {
		defer func() {
			assert.Equal(t, "value cannot be nil", recover())
		}()

		iter.SplitIntoRowsOf(1, nil)
		assert.Fail(t, "Must panic")
	}()
}

func TestSplitIntoColumns(t *testing.T) {
	// Split with n = 5 columns per subslice
	var (
		iter  = Of()
		split = iter.SplitIntoColumns(5)
	)
	assert.Equal(t, [][]interface{}{}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1)
	split = iter.SplitIntoColumns(5)
	assert.Equal(t, [][]interface{}{{1}}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1, 2, 3, 4)
	split = iter.SplitIntoColumns(5)
	assert.Equal(t, [][]interface{}{{1}, {2}, {3}, {4}}, split)

	iter = Of(1, 2, 3, 4, 5)
	split = iter.SplitIntoColumns(5)
	assert.Equal(t, [][]interface{}{{1}, {2}, {3}, {4}, {5}}, split)

	iter = Of(1, 2, 3, 4, 5, 6)
	split = iter.SplitIntoColumns(5)
	assert.Equal(t, [][]interface{}{{1, 6}, {2}, {3}, {4}, {5}}, split)

	iter = Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	split = iter.SplitIntoColumns(5)
	assert.Equal(t, [][]interface{}{{1, 6}, {2, 7}, {3, 8}, {4, 9}, {5, 10}}, split)

	iter = Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	split = iter.SplitIntoColumns(5)
	assert.Equal(t, [][]interface{}{{1, 6, 11}, {2, 7}, {3, 8}, {4, 9}, {5, 10}}, split)

	// Split with n = 1 columns per subslice corner case
	iter = Of()
	split = iter.SplitIntoColumns(1)
	assert.Equal(t, [][]interface{}{}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1)
	split = iter.SplitIntoColumns(1)
	assert.Equal(t, [][]interface{}{{1}}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1, 2)
	split = iter.SplitIntoColumns(1)
	assert.Equal(t, [][]interface{}{{1, 2}}, split)

	// Die if n < 1
	func() {
		defer func() {
			assert.Equal(t, "rows must be > 0", recover())
		}()

		iter.SplitIntoColumns(0)
		assert.Fail(t, "Must panic")
	}()
}

func TestSplitIntoColumnsOf(t *testing.T) {
	// Split with n = 5 columns per subslice
	var (
		iter  = Of()
		split = iter.SplitIntoColumnsOf(5, 0)
	)
	assert.Equal(t, [][]int{}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1)
	split = iter.SplitIntoColumnsOf(5, 0)
	assert.Equal(t, [][]int{{1}}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1, 2, 3, 4)
	split = iter.SplitIntoColumnsOf(5, 0)
	assert.Equal(t, [][]int{{1}, {2}, {3}, {4}}, split)

	iter = Of(1, 2, 3, 4, 5)
	split = iter.SplitIntoColumnsOf(5, 0)
	assert.Equal(t, [][]int{{1}, {2}, {3}, {4}, {5}}, split)

	iter = Of(1, 2, 3, 4, 5, 6)
	split = iter.SplitIntoColumnsOf(5, 0)
	assert.Equal(t, [][]int{{1, 6}, {2}, {3}, {4}, {5}}, split)

	iter = Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	split = iter.SplitIntoColumnsOf(5, 0)
	assert.Equal(t, [][]int{{1, 6}, {2, 7}, {3, 8}, {4, 9}, {5, 10}}, split)

	iter = Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	split = iter.SplitIntoColumnsOf(5, 0)
	assert.Equal(t, [][]int{{1, 6, 11}, {2, 7}, {3, 8}, {4, 9}, {5, 10}}, split)

	// Split into a type that requires conversion
	iter = Of(uint(1), uint(2))
	split = iter.SplitIntoColumnsOf(5, 0)
	assert.Equal(t, [][]int{{1}, {2}}, split)

	// Split with n = 1 columns per subslice corner case
	iter = Of()
	split = iter.SplitIntoColumnsOf(1, 0)
	assert.Equal(t, [][]int{}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1)
	split = iter.SplitIntoColumnsOf(1, 0)
	assert.Equal(t, [][]int{{1}}, split)

	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()

	iter = Of(1, 2)
	split = iter.SplitIntoColumnsOf(1, 0)
	assert.Equal(t, [][]int{{1, 2}}, split)

	// Die if n < 1
	func() {
		defer func() {
			assert.Equal(t, "rows must be > 0", recover())
		}()

		iter.SplitIntoColumnsOf(0, 0)
		assert.Fail(t, "Must panic")
	}()

	// Die if value is nil
	func() {
		defer func() {
			assert.Equal(t, "value cannot be nil", recover())
		}()

		iter.SplitIntoColumnsOf(1, nil)
		assert.Fail(t, "Must panic")
	}()
}

func TestToSlice(t *testing.T) {
	assert.Equal(t, []interface{}{}, Of().ToSlice())
	assert.Equal(t, []interface{}{1}, Of(1).ToSlice())
	assert.Equal(t, []interface{}{1, 2}, Of(1, 2).ToSlice())

	iter := Of()
	iter.ToSlice()
	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()
}

func TestToSliceOf(t *testing.T) {
	assert.Equal(t, []int{}, Of().ToSliceOf(0))
	assert.Equal(t, []int{1}, Of(1).ToSliceOf(0))
	assert.Equal(t, []int{1, 2}, Of(1, 2).ToSliceOf(0))

	iter := Of()
	iter.ToSliceOf(0)
	func() {
		defer func() {
			assert.Equal(t, "Iter.Next called on exhausted iterator", recover())
		}()

		iter.Next()
		assert.Fail(t, "Must panic")
	}()
}

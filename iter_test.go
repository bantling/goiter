package goiter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestIterSupplier struct {
	bool
}

func (TestIterSupplier) Iter() *Iter {
	return Of(10)
}

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

	// Non-array/slice
	func() {
		defer func() {
			assert.Equal(t, "ArraySliceIterFunc argument must be an array or slice", recover())
		}()

		ArraySliceIterFunc(reflect.ValueOf(1))

		assert.Fail(t, "Must panic on non-array/slice")
	}()
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

func TestIterFunc(t *testing.T) {
	// Nil Iter
	iterFunc := IterFunc(nil)

	_, next := iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty Iter
	iterFunc = IterFunc(Of())

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Iter of a value
	iterFunc = IterFunc(Of(1))

	val, next := iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Iter of two values
	iterFunc = IterFunc(Of(1, 2))

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
}

func TestChildrenIterFunc(t *testing.T) {
	// Empty items
	iterFunc := ChildrenIterFunc()

	_, next := iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty array
	iterFunc = ChildrenIterFunc(reflect.ValueOf([0]int{}))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element array
	iterFunc = ChildrenIterFunc(reflect.ValueOf([1]int{1}))

	val, next := iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty slice
	iterFunc = ChildrenIterFunc(reflect.ValueOf([]int{}))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element slice
	iterFunc = ChildrenIterFunc(reflect.ValueOf([]int{1}))

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty map
	iterFunc = ChildrenIterFunc(reflect.ValueOf(map[int]int{}))

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element map
	iterFunc = ChildrenIterFunc(reflect.ValueOf(map[int]int{1: 2}))

	val, next = iterFunc()
	assert.Equal(t, KeyValue{Key: 1, Value: 2}, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty *Iter
	iterFunc = ChildrenIterFunc(Of())

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element *Iter
	iterFunc = ChildrenIterFunc(Of(1))

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element IterSupplier
	iterFunc = ChildrenIterFunc(TestIterSupplier{})

	val, next = iterFunc()
	assert.Equal(t, 10, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element
	iterFunc = ChildrenIterFunc(reflect.ValueOf(5))

	val, next = iterFunc()
	assert.Equal(t, 5, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Empty Array/Slice/Map/*Iter, element
	iterFunc = ChildrenIterFunc([0]int{}, []int{}, map[int]int{}, Of(), 6)

	val, next = iterFunc()
	assert.Equal(t, 6, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// One element Array/Slice/Map/*Iter/IterSupplier, element
	iterFunc = ChildrenIterFunc([1]int{1}, []int{2}, map[int]int{3: 4}, Of(5), TestIterSupplier{}, 6)

	val, next = iterFunc()
	assert.Equal(t, 1, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 2, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, KeyValue{Key: 3, Value: 4}, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 5, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 10, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 6, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	// Two items, three elements
	iterFunc = ChildrenIterFunc(5, []int{6, 7})

	val, next = iterFunc()
	assert.Equal(t, 5, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 6, val)
	assert.True(t, next)

	val, next = iterFunc()
	assert.Equal(t, 7, val)
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

func TestOfChildren(t *testing.T) {
	// Empty items
	iter := OfChildren()

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
	iter = OfChildren(5)

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

	// Three items, four values
	iter = OfChildren(5, []int{6, 7}, Of(8))

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 5, iter.Value())

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 6, iter.Value())

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 7, iter.Value())

	next = iter.Next()
	assert.True(t, next)
	assert.Equal(t, 8, iter.Value())

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
			expected = []int{5, 6, 7}
		)

		for iter = OfChildren(5, []int{6, 7}); iter.Next(); {
			assert.Equal(t, expected[idx], iter.Value())
			idx++
		}

		assert.Equal(t, 3, idx)

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

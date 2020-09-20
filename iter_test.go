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

func TestSingleValueIterFunc (t *testing.T) {
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
	
	// One element
	iterFunc = ChildrenIterFunc(reflect.ValueOf(5))

	val, next = iterFunc()
	assert.Equal(t, 5, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)
	
	// Empty Array/Slice/Map, element
	iterFunc = ChildrenIterFunc([0]int{}, []int{}, map[int]int{}, 5)

	val, next = iterFunc()
	assert.Equal(t, 5, val)
	assert.True(t, next)

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)
	
	// One element Array/Slice/Map, element
	iterFunc = ChildrenIterFunc([1]int{1}, []int{2}, map[int]int{3: 4}, 5)

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

	_, next = iterFunc()
	assert.False(t, next)

	_, next = iterFunc()
	assert.False(t, next)
}

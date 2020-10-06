package goiter

import (
	"reflect"
)

// ArraySliceIterFunc iterates an array or slice
func ArraySliceIterFunc(arraySlice reflect.Value) func() (interface{}, bool) {
	if (arraySlice.Kind() != reflect.Array) && (arraySlice.Kind() != reflect.Slice) {
		panic("ArraySliceIterFunc argument must be an array or slice")
	}

	var (
		num = arraySlice.Len()
		idx = 0
	)

	return func() (interface{}, bool) {
		if idx == num {
			// Exhausted all values - don't care how many calls are made once exhausted
			return nil, false
		}

		// Return value at current index, and increment index for next time
		val := arraySlice.Index(idx).Interface()
		idx++
		return val, true
	}
}

// Iterable is a supplier of an Iter
type Iterable interface {
	Iter() *Iter
}

// IterablesFunc iterates the values of any number of Iterables in the order passed
func IterablesFunc(iterables []Iterable) func() (interface{}, bool) {
	var (
		num         = len(iterables)
		idx         = 0
		theIterable Iterable
		theIter     *Iter
	)

	return func() (interface{}, bool) {
		// Continue to return values from current iter until it is empty
		if theIter != nil {
			if theIter.Next() {
				return theIter.Value(), true
			}

			// Nilify current iter once it is empty
			theIter = nil
		}

		// Search any remaining iters for the next non-nil non-empty iter, if one exists
		for (theIter == nil) && (idx < num) {
			if theIterable, idx = iterables[idx], idx+1; theIterable != nil {
				if theIter = theIterable.Iter(); (theIter != nil) && theIter.Next() {
					return theIter.Value(), true
				}
			}

			// Nilify iter in case it is non-nil and empty
			theIter = nil
		}

		// No values left to iterate
		return nil, false
	}
}

// KeyValue contains a key value pair from a map
type KeyValue struct {
	Key   interface{}
	Value interface{}
}

// MapIterFunc iterates a map
func MapIterFunc(aMap reflect.Value) func() (interface{}, bool) {
	if aMap.Kind() != reflect.Map {
		panic("MapIterFunc argument must be a map")
	}

	var (
		mapIter = aMap.MapRange()
		done    bool
	)

	return func() (interface{}, bool) {
		// Return immediately if further calls are made after last key was iterated
		if done {
			return nil, false
		}

		// Advance MapIter to next key/value pair, if any
		if !mapIter.Next() {
			// Exhausted all values
			done = true
			return nil, false
		}

		// Return next key/value pair
		val := KeyValue{
			Key:   mapIter.Key().Interface(),
			Value: mapIter.Value().Interface(),
		}
		return val, true
	}
}

// NoValueIterFunc always returns (nil, false)
func NoValueIterFunc() (interface{}, bool) {
	return nil, false
}

// SingleValueIterFunc iterates a single value
func SingleValueIterFunc(aVal reflect.Value) func() (interface{}, bool) {
	done := false

	return func() (interface{}, bool) {
		if done {
			return nil, false
		}

		// First call returns the wrapped value given
		done = true
		return aVal.Interface(), true
	}
}

// ElementsIterFunc returns an iterator function that iterates the elements of the item passed.
// The item is handled as follows:
// - Array or Slice: returns ArraySliceIterFunc(item)
// - Iterable: returns IterFunc(item)
// - Map: returns MapIterFunc(item)
// - Nil ptr: returns NoValueIterFunc
// - Otherwise returns SingleValueIterFunc(item)
func ElementsIterFunc(item reflect.Value) func() (interface{}, bool) {
	switch item.Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		return ArraySliceIterFunc(item)
	case reflect.Map:
		return MapIterFunc(item)
	default:
		if iterableObj, isa := item.Interface().(Iterable); isa {
			return IterablesFunc([]Iterable{iterableObj})
		}

		if (item.Kind() == reflect.Ptr) && item.IsNil() {
			return NoValueIterFunc
		}

		return SingleValueIterFunc(item)
	}
}

// Iter is an iterator of values of an arbitrary type.
// Technically, the values can be different types, but that is usually undesirable.
type Iter struct {
	iter       func() (interface{}, bool)
	nextCalled bool
	value      interface{}
}

// NewIter constructs an Iter from an iterating function.
// The function must returns (nextItem, true) for every item available to iterate, then return (invalid, false) on the next call after the last item.
// Once the function returns a false bool value, it will never be called again.
// Panics if iter is nil.
func NewIter(iter func() (interface{}, bool)) *Iter {
	if iter == nil {
		panic("NewIter requires an iterator")
	}

	return &Iter{iter: iter}
}

// Of constructs an Iter that iterates the items passed.
// If any item is an array/slice/map/Iterable, it will be handled the same as any other type - the whole array/slice/map/Iterable will iterated as a single value.
func Of(items ...interface{}) *Iter {
	return NewIter(ArraySliceIterFunc(reflect.ValueOf(items)))
}

// OfElements constructs an Iter that iterates the elements of the item passed.
// See ElementsIterFunc for details of how different types are handled.
func OfElements(item interface{}) *Iter {
	if item == nil {
		// Can't call reflect.ValueOf(nil)
		return NewIter(NoValueIterFunc)
	}

	return NewIter(ElementsIterFunc(reflect.ValueOf(item)))
}

// OfIterables constructs an Iter that iterates each Iterable passed.
func OfIterables(iterables ...Iterable) *Iter {
	return NewIter(IterablesFunc(iterables))
}

// Next returns true if there is another item to be read by Value.
// Once Next returns false, further calls to Next or Value panic.
func (it *Iter) Next() bool {
	// Die if iterator already exhausted
	if it.iter == nil {
		panic("Iter.Next called on exhausted iterator")
	}

	// Try to get next item
	if value, haveIt := it.iter(); haveIt {
		// If we have it, keep the value for call to Value() and return true
		it.nextCalled = true
		it.value = value
		return true
	}

	// First call with no more items, mark done and return false
	it.iter = nil
	return false
}

// Value returns the value retrieved by the prior call to Next.
// In the case of iterating a map, each value will be returned as a KeyValue instance, passed by value.
// Panics if the iterator is exhausted.
// Panics if Next has not been called since the last time Value was called.
func (it *Iter) Value() interface{} {
	if it.iter == nil {
		panic("Iter.Value called on exhausted iteraror")
	}

	if !it.nextCalled {
		panic("Iter.Next has to be called before iter.Value")
	}

	// Clear nextCalled flag
	it.nextCalled = false
	return it.value
}

// Iter is the Iterable interface.
// By implementing Iterable, algorithms can be written against only Iterable, and accept *Iter or Iterable instances.
// Returns pointer, as all callers to this iter are exhausting the same set of data.
// As a rule, there should only be one owner of this iterator.
func (it *Iter) Iter() *Iter {
	return it
}

// SplitIntoRows splits the iterator into rows of at most the number of columns specified.
// This operation will exhaust the iter.
// Panics if the iter has already been exhausted.
// Panics if cols = 0.
func (it *Iter) SplitIntoRows(cols uint) [][]interface{} {
	if cols == 0 {
		panic("cols must be > 0")
	}

	var (
		split = [][]interface{}{}
		row   = make([]interface{}, 0, cols)
		idx   uint
	)

	for it.Next() {
		val := it.Value()
		row = append(row, val)
		idx++

		if idx == cols {
			split = append(split, row)
			row = make([]interface{}, 0, cols)
			idx = 0
		}
	}

	// If len == 0, must be a corner case: no items, or an exact multiple of n items.
	// Otherwise, row contains a partial slice of the last < n items.
	if len(row) > 0 {
		split = append(split, row)
	}

	return split
}

// SplitIntoColumns splits the iterator into columns with at most the number of rows specified.
// This method is simply the transpose operation of SplitIntoRows.
// This operation will exhaust the iter.
// Panics if the iter has already been exhausted.
// Panics if rows = 0.
func (it *Iter) SplitIntoColumns(rows uint) [][]interface{} {
	if rows == 0 {
		panic("rows must be > 0")
	}

	var (
		split = [][]interface{}{}
		idx   uint
	)

	// Start by creating up to the specified number of rows with one element each
	for idx = 0; idx < rows; idx++ {
		if !it.Next() {
			// Less elements than specified number of rows, return the one element rows we have
			return split
		}

		split = append(split, []interface{}{it.Value()})
	}

	// Populate columns top to bottom with remaining elements
	for idx = 0; it.Next(); {
		split[idx] = append(split[idx], it.Value())

		if idx++; idx == rows {
			idx = 0
		}
	}

	return split
}

// ToSlice collects the elements into a slice
func (it *Iter) ToSlice() []interface{} {
	slice := []interface{}{}

	for it.Next() {
		slice = append(slice, it.Value())
	}

	return slice
}

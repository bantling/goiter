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

// IterFunc returns an iterator func that iterates the values of an *Iter
func IterFunc(iter *Iter) func() (interface{}, bool) {
	theIter := iter

	return func() (interface{}, bool) {
		if theIter != nil {
			if theIter.Next() {
				return theIter.Value(), true
			}

			theIter = nil
		}

		return nil, false
	}
}

// Iterable is a supplier of an Iter
type Iterable interface {
	Iter() *Iter
}

// ChildrenIterFunc returns an iterator function that iterates the items passed, if any.
// It is valid to not pass any items, the Iter will simply return false on first call to Next.
// Items are handled as follows:
// - A precheck will handle a reflect.Value the same as an unwrapped value
// - Slice: the elements of the slice are iterated non-recursively
// - Map: the key/value pairs of the map are iterated non-recursively, and returned as KeyValue objects
// - Nil ptr: Skipped
// - Iterable: the Iter() method is called to get an *Iter, which will be iterated.
func ChildrenIterFunc(items ...interface{}) func() (interface{}, bool) {
	var (
		num  = len(items)
		idx  = 0
		iter func() (interface{}, bool)
	)

	return func() (interface{}, bool) {
		for {
			if iter != nil {
				// Try getting next value of current item being iterated
				if val, haveValue := iter(); haveValue {
					// Have another value, return it
					return val, true
				}

				// Exhausted current iter, try next item
				iter = nil
				idx++
			}

			// iter must be nil
			if idx == num {
				// Exhausted all items - don't care how many calls are made once exhausted
				return nil, false
			}

			// CurrentIter must be nil or exhausted
			// Need to get the iterator func for value(s) of the next item passed
			var (
				item    = items[idx]
				itemVal reflect.Value
				isa     bool
			)

			if itemVal, isa = item.(reflect.Value); !isa {
				itemVal = reflect.ValueOf(item)
			}

			if iterableObj, isa := itemVal.Interface().(Iterable); isa {
				// IterSupplier could be value or pointer receiver
				iter = IterFunc(iterableObj.Iter())
			} else {
				switch itemVal.Kind() {
				case reflect.Array:
					fallthrough
				case reflect.Slice:
					iter = ArraySliceIterFunc(itemVal)
				case reflect.Map:
					iter = MapIterFunc(itemVal)
				case reflect.Ptr:
					if itemVal.IsNil() {
						// Try next item
						idx++
						continue
					}
					fallthrough
				default:
					iter = SingleValueIterFunc(itemVal)
				}
			}
			// Next iteration will now have a non-nil iter, which may be for an empty slice or map.
			// We'll just keep going through items until we find a non-empty item or run out of items.
		}
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

// OfChildren constructs an Iter that iterates the children of the items passed.
// If any item is an array/slice/map/Iterable, then the values contained in it will be iterated non-recursively.
// An item of any other type will just be iterated as a single value.
func OfChildren(items ...interface{}) *Iter {
	return NewIter(ChildrenIterFunc(items...))
}

// Next returns true if there is another item to be read by Value.
// Once Next returns false, further calls to Next or Value panic.
func (i *Iter) Next() bool {
	// Die if iterator already exhausted
	if i.iter == nil {
		panic("Iter.Next called on exhausted iterator")
	}

	// Try to get next item
	if value, haveIt := i.iter(); haveIt {
		// If we have it, keep the value for call to Value() and return true
		i.nextCalled = true
		i.value = value
		return true
	}

	// First call with no more items, mark done and return false
	i.iter = nil
	return false
}

// Value returns the value retrieved by the prior call to Next.
// In the case of iterating a map, each value will be returned as a KeyValue instance, passed by value.
// Panics if the iterator is exhausted.
// Panics if Next has not been called since the last time Value was called.
func (i *Iter) Value() interface{} {
	if i.iter == nil {
		panic("Iter.Value called on exhausted iteraror")
	}

	if !i.nextCalled {
		panic("Iter.Next has to be called before iter.Value")
	}

	// Clear nextCalled flag
	i.nextCalled = false
	return i.value
}

// Iter is the Iterable interface.
// By implementing Iterable, algorithms can be written against only Iterable, and accept *Iter or Iterable instances.
// Returns pointer, as all callers to this iter are exhausting the same set of data.
// As a rule, there should only be one owner of this iterator.
func (i *Iter) Iter() *Iter {
	return i
}

// Split the iterator into slices of at most size n.
// This operation will exhaust the iter, and will panic if Next() or Value() is called after it.
// Panics if n = 0.
func (i *Iter) Split(n uint) [][]interface{} {
	if n == 0 {
		panic("n must be > 0")
	}

	var (
		split   = [][]interface{}{}
		current = make([]interface{}, 0, n)
		idx     uint
	)

	for i.Next() {
		val := i.Value()
		current = append(current, val)
		idx++

		if idx == n {
			split = append(split, current)
			current = make([]interface{}, 0, n)
			idx = 0
		}
	}

	// If len == 0, must be a corner case: no items, or an exact multiple of n items.
	// Otherwise, current contains a partial slice of the last < n items.
	if len(current) > 0 {
		split = append(split, current)
	}

	return split
}

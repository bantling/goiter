// SPDX-License-Identifier: Apache-2.0

package goiter

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"
)

// Error constants
const (
	InvalidUTF8EncodingError            = "Invalid UTF 8 encoding"
	ErrArraySliceIterFuncArg            = "ArraySliceIterFunc argument must be an array or slice"
	ErrMapIterFuncArg                   = "MapIterFunc argument must be a map"
	ErrNewIterNeedsIterator             = "NewIter requires an iterator"
	ErrNextExhaustedIter                = "Iter.Next called on exhausted iterator"
	ErrValueExhaustedIter               = "Iter.Value called on exhausted iterator"
	ErrValueNextFirst                   = "Iter.Next has to be called before iter.Value"
	ErrValueCannotBeNil                 = "value cannot be nil"
	ErrUnreadExhaustedIter              = "Iter.Unread called on exhausted iterator"
	ErrColsGreaterThanZero              = "cols must be > 0"
	ErrRowsGreaterThanZero              = "rows must be > 0"
	ErrIterableGeneratorCannotBeNil     = "Iterable.Generator cannot be nil"
	ErrIterableGeneratorCannotReturnNil = "Iterable.Generator cannot return a nil iterating function"
)

var (
	zeroUTF8Buffer = []byte{0, 0, 0, 0}
)

// ==== Iterator function generators

// ArraySliceIterFunc iterates an array or slice outermost dimension.
// EG, if an [][]int is passed, the iterator returns []int values.
// Panics if the value is not an array or slice.
func ArraySliceIterFunc(arraySlice reflect.Value) func() (interface{}, bool) {
	if (arraySlice.Kind() != reflect.Array) && (arraySlice.Kind() != reflect.Slice) {
		panic(ErrArraySliceIterFuncArg)
	}

	var (
		num = arraySlice.Len()
		idx int
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
		panic(ErrMapIterFuncArg)
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

// ElementsIterFunc returns an iterator function that iterates the elements of the item passed non-recursively.
// The item is handled as follows:
// - Array or Slice: returns ArraySliceOuterIterFunc(item)
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
		if (item.Kind() == reflect.Ptr) && item.IsNil() {
			return NoValueIterFunc
		}

		return SingleValueIterFunc(item)
	}
}

// ReaderIterFunc iterates the bytes of an io.Reader.
// For each byte in the Reader, returns (byte, true).
// When eof read, returns (0, false).
// When any other error occurs, panics with the error.
func ReaderIterFunc(src io.Reader) func() (interface{}, bool) {
	buf := make([]byte, 1)

	return func() (interface{}, bool) {
		if _, err := src.Read(buf); err != nil {
			if err != io.EOF {
				panic(err)
			}

			return 0, false
		}

		return buf[0], true
	}
}

// ReaderToRunesIterFunc iterates the bytes of an io.Reader, and interprets them as UTF-8 runes.
// For each valid rune contained in the Reader, returns (rune, true).
// When EOF read, returns (utf8.RuneError, false).
// When any other error occurs (including invalid UTF-8 encoding), panics with the error.
func ReaderToRunesIterFunc(src io.Reader) func() (interface{}, bool) {
	// UTF-8 requires at most 4 bytes for a code point
	var (
		buf    = make([]byte, 4)
		bufPos int
	)

	return func() (interface{}, bool) {
		// Read next up to 4 bytes from reader into subslice of buffer, after any remaining bytes from last read
		if _, err := src.Read(buf[bufPos:]); (err != nil) && (err != io.EOF) {
			panic(err)
		}

		// If first byte is 0 after reading, must have emptied source and returned all runes
		if buf[0] == 0 {
			return utf8.RuneError, false
		}

		// Decode up to 4 bytes for next code point
		r, rl := utf8.DecodeRune(buf)
		if r == utf8.RuneError {
			panic(InvalidUTF8EncodingError)
		}

		// Shift any remaining unused bytes back to the begining of the buffer
		copy(buf, buf[rl:])

		// Next time read up to as many bytes as were shifted from source, overwriting remaining bytes
		bufPos = 4 - rl

		// Clear out the unused bytes at the end, in case we don't have enough bytes left to fill them
		copy(buf[bufPos:], zeroUTF8Buffer)

		return r, true
	}
}

// ReaderToLinesIterFunc iterates the bytes of an io.Reader, and interprets them as runes.
// Runes are read until an EOL sequence occurs (CR, LF, CRLF) or EOF occurs.
// For each line contained in the Reader, returns (string, true), where the string does not contain an EOL sequence.
// After the last line has been returned, all further calls return ("", false).
// When any other error occurs (including invalid UTF-8 encoding), panics with the error.
func ReaderToLinesIterFunc(src io.Reader) func() (interface{}, bool) {
	// Use ReaderToRunesIterFunc to read individual runes until a line is read
	var (
		runesIter = ReaderToRunesIterFunc(src)
		str       strings.Builder
		lastCR    bool
	)

	return func() (interface{}, bool) {
		str.Reset()

		for {
			codePoint, haveIt := runesIter()

			if !haveIt {
				if str.Len() > 0 {
					return str.String(), true
				}

				return "", false
			}

			if codePoint == '\r' {
				lastCR = true
				return str.String(), true
			}

			if codePoint == '\n' {
				if lastCR {
					lastCR = false
					continue
				}

				return str.String(), true
			}

			str.WriteRune(codePoint.(rune))
		}
	}
}

// FlattenArraySlice flattens an array or slice of any number of dimensions into a new slice of one dimension.
// EG, an [][]int{{1, 2}, {3, 4, 5}} is flattened into an []interface{}{1,2,3,4,5}.
// Note that in case where the element type is interface{}, a mixture of values and arrays/slices could be used.
// EG, an []interface{}{1, [2]int{2, 3}, [][]string{{"4", "5"}, {"6", "7", "8"}}} is flattened into []interface{}{1, 2, 3, "4", "5", "6", "7", "8"}.
// Panics if the value is not an array or slice.
func FlattenArraySlice(value interface{}) []interface{} {
	arraySlice := reflect.ValueOf(value)
	if (arraySlice.Kind() != reflect.Array) && (arraySlice.Kind() != reflect.Slice) {
		panic("FlattenArraySlice argument must be an array or slice")
	}

	// Make a one dimensional slice
	result := []interface{}{}

	// Recursive function
	var f func(reflect.Value)
	f = func(currentArraySlice reflect.Value) {
		// Iterate current array or slice
		for i, num := 0, currentArraySlice.Len(); i < num; i++ {
			val := reflect.ValueOf(currentArraySlice.Index(i).Interface())

			// Recurse sub-arrays/slices
			if (val.Kind() == reflect.Array) || (val.Kind() == reflect.Slice) {
				f(val)
			} else {
				result = append(result, val.Interface())
			}
		}
	}
	f(arraySlice)

	return result
}

// FlattenArraySliceAsType flattens an array or slice of any number of dimensions into a new slice of one dimension,
// where the slice type is the same as the given element.
// EG, an [][]int{{1, 2}, {3, 4, 5}} can be flattened into an []int{}{1,2,3,4,5}.
// Note that in case where the element type is interface{}, a mixture of values and arrays/slices could be used.
// EG, an []interface{}{1, [2]int{2, 3}, [][]uint{{4, 5}, {6, 7, 8}}} can be flattened into []int{}{1, 2, 3, 4, 5, 6, 7, 8}.
// Panics if the value is not an array or slice.
func FlattenArraySliceAsType(value interface{}, elementVal interface{}) interface{} {
	arraySlice := reflect.ValueOf(value)
	if (arraySlice.Kind() != reflect.Array) && (arraySlice.Kind() != reflect.Slice) {
		panic("FlattenArraySliceAs value must be an array or slice")
	}

	// Make a one dimensional slice that has the same type as the type of elementVal
	var (
		typ    = reflect.TypeOf(elementVal)
		result = reflect.MakeSlice(reflect.SliceOf(typ), 0, 0)
	)

	// Recursive function
	var f func(reflect.Value)
	f = func(currentArraySlice reflect.Value) {
		// Iterate current array or slice
		for i, num := 0, currentArraySlice.Len(); i < num; i++ {
			val := reflect.ValueOf(currentArraySlice.Index(i).Interface())

			// Recurse sub-arrays/slices
			if (val.Kind() == reflect.Array) || (val.Kind() == reflect.Slice) {
				f(val)
			} else {
				result = reflect.Append(result, val.Convert(typ))
			}
		}
	}
	f(arraySlice)

	return result.Interface()
}

// ==== Iter

// Iter is an iterator of values of an arbitrary type.
// Technically, the values can be different types, but that is usually undesirable.
type Iter struct {
	iter       func() (interface{}, bool)
	nextCalled bool
	value      interface{}
	buffer     []interface{}
}

// NewIter constructs an Iter from an iterating function.
// The function must returns (nextItem, true) for every item available to iterate, then return (invalid, false) on the next call after the last item.
// Once the function returns a false bool value, it will never be called again.
// Panics if iter is nil.
func NewIter(iter func() (interface{}, bool)) *Iter {
	if iter == nil {
		panic(ErrNewIterNeedsIterator)
	}

	return &Iter{iter: iter}
}

// Of constructs an Iter that iterates the items passed.
// If any item is an array/slice/map/Iterable, it will be handled the same as any other type - the whole array/slice/map/Iterable will iterated as a single value.
func Of(items ...interface{}) *Iter {
	return NewIter(ArraySliceIterFunc(reflect.ValueOf(items)))
}

// OfFlatten constructs an Iter that flattens a multi-dimensional array or slice into a new one-dimensional slice.
// See FlattenArraySlice.
func OfFlatten(items interface{}) *Iter {
	if items == nil {
		// Can't call reflect.ValueOf(nil)
		return NewIter(NoValueIterFunc)
	}

	return NewIter(ArraySliceIterFunc(reflect.ValueOf(FlattenArraySlice(items))))
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

// OfReader constructs an Iter that iterates the bytes of a reader.
// See ReaderIterFunc for details.
func OfReader(src io.Reader) *Iter {
	return NewIter(ReaderIterFunc(src))
}

// OfReaderRunes constructs an Iter that iterates the runes of a reader.
// See ReaderToRunesIterFunc for details.
func OfReaderRunes(src io.Reader) *Iter {
	return NewIter(ReaderToRunesIterFunc(src))
}

// OfReaderLines constructs an Iter that iterates the lines of a reader.
// See ReaderToLinesIterFunc for details.
func OfReaderLines(src io.Reader) *Iter {
	return NewIter(ReaderToLinesIterFunc(src))
}

// Next returns true if there is another item to be read by Value.
// Once Next returns false, further calls to Next or Value panic.
func (it *Iter) Next() bool {
	// Die if iterator already exhausted
	if it.iter == nil {
		panic(ErrNextExhaustedIter)
	}

	// Check buffer first in case items have been unread
	if l := len(it.buffer); l > 0 {
		it.nextCalled = true
		it.value = it.buffer[l-1]
		it.buffer = it.buffer[:l-1]
		return true
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
		panic(ErrValueExhaustedIter)
	}

	if !it.nextCalled {
		panic(ErrValueNextFirst)
	}

	// Clear nextCalled flag
	it.nextCalled = false
	return it.value
}

// ValueOfType reads the value and converts it to a value with the same type as the given value.
// EG, if an int is passed, it converts the value to an int.
// The result will have to be type asserted.
// Panics is value is nil.
// Panics if Value() method panics.
// Panics if the value is not convertible to the type of the given value.
func (it *Iter) ValueOfType(value interface{}) interface{} {
	if value == nil {
		panic(ErrValueCannotBeNil)
	}

	return reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(value)).Interface()
}

// NextValue retrieves the next value for cases where you know the iterator has another value.
// Panics if Next() or Value() panics.
func (it *Iter) NextValue() interface{} {
	it.Next()
	return it.Value()
}

// NextValueOfType retrieves the next value with the same type as the given value for cases where you know the iterator has another value.
// Panics if Next() or ValueOfType() panics.
func (it *Iter) NextValueOfType(value interface{}) interface{} {
	it.Next()
	return it.ValueOfType(value)
}

// BoolValue reads the value and converts it to a bool.
// Panics if Value() method panics.
// Panics if the value is not convertible to a bool.
func (it *Iter) BoolValue() bool {
	return reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(true)).Bool()
}

// NextBoolValue retrieves the next value as a bool for cases where you know the iterator has another value.
// Panics if Next() or BoolValue() panics.
func (it *Iter) NextBoolValue() bool {
	it.Next()
	return it.BoolValue()
}

// ByteValue reads the value and converts it to a byte.
// Panics if Value() method panics.
// Panics if the value is not convertible to a byte.
func (it *Iter) ByteValue() byte {
	return byte(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(byte(0))).Uint())
}

// NextByteValue retrieves the next value as a byte for cases where you know the iterator has another value.
// Panics if Next() or ByteValue() panics.
func (it *Iter) NextByteValue() byte {
	it.Next()
	return it.ByteValue()
}

// RuneValue reads the value and converts it to a rune.
// Panics if Value() method panics.
// Panics if the value is not convertible to a rune.
func (it *Iter) RuneValue() rune {
	return rune(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(rune(0))).Int())
}

// NextRuneValue retrieves the next value as a rune for cases where you know the iterator has another value.
// Panics if Next() or RuneValue() panics.
func (it *Iter) NextRuneValue() rune {
	it.Next()
	return it.RuneValue()
}

// IntValue reads the value and converts it to an int.
// Panics if Value() method panics.
// Panics if the value is not convertible to an int.
func (it *Iter) IntValue() int {
	return int(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(0)).Int())
}

// NextIntValue retrieves the next value as an int for cases where you know the iterator has another value.
// Panics if Next() or IntValue() panics.
func (it *Iter) NextIntValue() int {
	it.Next()
	return it.IntValue()
}

// Int8Value reads the value and converts it to an int8.
// Panics if Value() method panics.
// Panics if the value is not convertible to an int8.
func (it *Iter) Int8Value() int8 {
	return int8(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(int8(0))).Int())
}

// NextInt8Value retrieves the next value as an int8 for cases where you know the iterator has another value.
// Panics if Next() or Int8Value() panics.
func (it *Iter) NextInt8Value() int8 {
	it.Next()
	return it.Int8Value()
}

// Int16Value reads the value and converts it to an int16.
// Panics if Value() method panics.
// Panics if the value is not convertible to an int16.
func (it *Iter) Int16Value() int16 {
	return int16(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(int16(0))).Int())
}

// NextInt16Value retrieves the next value as an int16 for cases where you know the iterator has another value.
// Panics if Next() or Int16Value() panics.
func (it *Iter) NextInt16Value() int16 {
	it.Next()
	return it.Int16Value()
}

// Int32Value reads the value and converts it to an int32.
// Panics if Value() method panics.
// Panics if the value is not convertible to an int32.
func (it *Iter) Int32Value() int32 {
	return int32(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(int32(0))).Int())
}

// NextInt32Value retrieves the next value as an int32 for cases where you know the iterator has another value.
// Panics if Next() or Int32Value() panics.
func (it *Iter) NextInt32Value() int32 {
	it.Next()
	return it.Int32Value()
}

// Int64Value reads the value and converts it to an int64.
// Panics if Value() method panics.
// Panics if the value is not convertible to an int64.
func (it *Iter) Int64Value() int64 {
	return reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(int64(0))).Int()
}

// NextInt64Value retrieves the next value as an int64 for cases where you know the iterator has another value.
// Panics if Next() or Int64Value() panics.
func (it *Iter) NextInt64Value() int64 {
	it.Next()
	return it.Int64Value()
}

// UintValue reads the value and converts it to a uint.
// Panics if Value() method panics.
// Panics if the value is not convertible to a uint.
func (it *Iter) UintValue() uint {
	return uint(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(uint(0))).Uint())
}

// NextUintValue retrieves the next value as a uint for cases where you know the iterator has another value.
// Panics if Next() or UintValue() panics.
func (it *Iter) NextUintValue() uint {
	it.Next()
	return it.UintValue()
}

// Uint8Value reads the value and converts it to a uint8.
// Panics if Value() method panics.
// Panics if the value is not convertible to a uint8.
func (it *Iter) Uint8Value() uint8 {
	return uint8(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(uint8(0))).Uint())
}

// NextUint8Value retrieves the next value as a uint8 for cases where you know the iterator has another value.
// Panics if Next() or Uint8Value() panics.
func (it *Iter) NextUint8Value() uint8 {
	it.Next()
	return it.Uint8Value()
}

// Uint16Value reads the value and converts it to a uint16.
// Panics if Value() method panics.
// Panics if the value is not convertible to a uint16.
func (it *Iter) Uint16Value() uint16 {
	return uint16(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(uint16(0))).Uint())
}

// NextUint16Value retrieves the next value as a uint16 for cases where you know the iterator has another value.
// Panics if Next() or Uint16Value() panics.
func (it *Iter) NextUint16Value() uint16 {
	it.Next()
	return it.Uint16Value()
}

// Uint32Value reads the value and converts it to a uint32.
// Panics if Value() method panics.
// Panics if the value is not convertible to a uint32.
func (it *Iter) Uint32Value() uint32 {
	return uint32(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(uint32(0))).Uint())
}

// NextUint32Value retrieves the next value as a uint32 for cases where you know the iterator has another value.
// Panics if Next() or Uint32Value() panics.
func (it *Iter) NextUint32Value() uint32 {
	it.Next()
	return it.Uint32Value()
}

// Uint64Value reads the value and converts it to a uint64.
// Panics if Value() method panics.
// Panics if the value is not convertible to a uint64.
func (it *Iter) Uint64Value() uint64 {
	return reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(uint64(0))).Uint()
}

// NextUint64Value retrieves the next value as a uint64 for cases where you know the iterator has another value.
// Panics if Next() or Uint64Value() panics.
func (it *Iter) NextUint64Value() uint64 {
	it.Next()
	return it.Uint64Value()
}

// Float32Value reads the value and converts it to a float32.
// Panics if Value() method panics.
// Panics if the value is not convertible to a float32.
func (it *Iter) Float32Value() float32 {
	return float32(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(float32(0))).Float())
}

// NextFloat32Value retrieves the next value as a float32 for cases where you know the iterator has another value.
// Panics if Next() or Float32Value() panics.
func (it *Iter) NextFloat32Value() float32 {
	it.Next()
	return it.Float32Value()
}

// Float64Value reads the value and converts it to a float64.
// Panics if Value() method panics.
// Panics if the value is not convertible to a float64.
func (it *Iter) Float64Value() float64 {
	return reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(float64(0))).Float()
}

// NextFloat64Value retrieves the next value as a float64 for cases where you know the iterator has another value.
// Panics if Next() or Float64Value() panics.
func (it *Iter) NextFloat64Value() float64 {
	it.Next()
	return it.Float64Value()
}

// Complex64Value reads the value and converts it to a complex64.
// Panics if Value() method panics.
// Panics if the value is not convertible to a complex64.
func (it *Iter) Complex64Value() complex64 {
	return complex64(reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(complex64(0))).Complex())
}

// NextComplex64Value retrieves the next value as a complex64 for cases where you know the iterator has another value.
// Panics if Next() or Complex64Value() panics.
func (it *Iter) NextComplex64Value() complex64 {
	it.Next()
	return it.Complex64Value()
}

// Complex128Value reads the value and converts it to a complex128.
// Panics if Value() method panics.
// Panics if the value is not convertible to a complex128.
func (it *Iter) Complex128Value() complex128 {
	return reflect.ValueOf(it.Value()).Convert(reflect.TypeOf(complex128(0))).Complex()
}

// NextComplex128Value retrieves the next value as a complex128 for cases where you know the iterator has another value.
// Panics if Next() or Complex128Value() panics.
func (it *Iter) NextComplex128Value() complex128 {
	it.Next()
	return it.Complex128Value()
}

// StringValue reads the value and converts it to a string.
// Panics if Value() method panics.
// Panics if the value is not convertible to a string.
func (it *Iter) StringValue() string {
	return fmt.Sprintf("%s", reflect.ValueOf(it.Value()).Convert(reflect.TypeOf("")))
}

// NextStringValue retrieves the next value as a string for cases where you know the iterator has another value.
// Panics if Next() or StringValue() panics.
func (it *Iter) NextStringValue() string {
	it.Next()
	return it.StringValue()
}

// Unread places the given value at the end of an internal buffer of unread values.
// It is up to the caller to unread correctly.
// Example:
// - the source has values 1,2,3
// - values 1,2 have been iterated, 3 has not
// - caller can choose to unread 2, so that Next/Value returns 2 from buffer without consulting source
// - calling Next again consults source to read 3
// - caller can unread 3,2,1, so that Next/Value returns 1,2,3 without consulting source
// - calling Next again returns false
// There is nothing preventing the caller from reading 1,2,3 and unreading 1,2,3 causing Next/Value to return 3,2,1.
// Panics if the iterator is exhausted.
func (it *Iter) Unread(val interface{}) {
	// Die if iterator already exhausted
	if it.iter == nil {
		panic(ErrUnreadExhaustedIter)
	}

	it.buffer = append(it.buffer, val)
}

// SplitIntoRows splits the iterator into rows of at most the number of columns specified.
// Since the number of items to iterate is not known, the algorithm fills across the first row from left to right,
// then fills across the second row, and so on.
// The original ordering is retained by iterating each row from left to right.
// If the number of items <= the number of columns, a single row is returned.
// This operation will exhaust the iter.
// Panics if the iter has already been exhausted.
// Panics if cols = 0.
func (it *Iter) SplitIntoRows(cols uint) [][]interface{} {
	if cols == 0 {
		panic(ErrColsGreaterThanZero)
	}

	var (
		split = [][]interface{}{}
		row   = make([]interface{}, 0, cols)
		idx   uint
	)

	for it.Next() {
		row = append(row, it.Value())
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

// SplitIntoRowsOf is a version of SplitIntoRows where the slice type is the same as the type of the given value.
// EG, if a value of type int is passed, a [][]int is returned.
// This operation will exhaust the iter.
// Panics if the iter has already been exhausted.
// Panics if cols = 0.
// Panics is value is nil.
// Panics if any value is not convertible to the type of the given value.
func (it *Iter) SplitIntoRowsOf(cols uint, value interface{}) interface{} {
	if cols == 0 {
		panic(ErrColsGreaterThanZero)
	}

	if value == nil {
		panic(ErrValueCannotBeNil)
	}

	var (
		intCols = int(cols)
		typ     = reflect.TypeOf(value)
		split   = reflect.MakeSlice(reflect.SliceOf(reflect.SliceOf(typ)), 0, 0)
		row     = reflect.MakeSlice(reflect.SliceOf(typ), 0, intCols)
		idx     uint
	)

	for it.Next() {
		row = reflect.Append(row, reflect.ValueOf(it.Value()).Convert(typ))
		idx++

		if idx == cols {
			split = reflect.Append(split, row)
			row = reflect.MakeSlice(reflect.SliceOf(typ), 0, intCols)
			idx = 0
		}
	}

	// If len == 0, must be a corner case: no items, or an exact multiple of n items.
	// Otherwise, row contains a partial slice of the last < n items.
	if row.Len() > 0 {
		split = reflect.Append(split, row)
	}

	return split.Interface()
}

// SplitIntoColumns splits the iterator into columns with at most the number of rows specified.
// The algorithm reads all the items into a slice first to determine the number of them and ensures that each row has the same number of columns, except for a remainder spread across one or more rows.
// EG, if 23 items exist and rows = 5, 23 / 5 = 4 r 3, so the first 3 rows have 5 items (4 + 1 from remainder), the last 2 have 4: 3 * 5 + 2 * 4 = 15 + 8 = 23.
// If the number of items <= the number of rows, then the number of rows = number of items, 1 item per row.
// This method is simply the transpose operation of SplitIntoRows.
// This operation will exhaust the iter.
// Panics if the iter has already been exhausted.
// Panics if rows = 0.
func (it *Iter) SplitIntoColumns(rows uint) [][]interface{} {
	if rows == 0 {
		panic(ErrRowsGreaterThanZero)
	}

	// Collect values into a slice first
	var (
		values         = it.ToSlice()
		numValues      = len(values)
		numRows        = int(rows)
		numItems, rmdr = numValues / numRows, numValues % numRows
		start, end     int
		split          = [][]interface{}{}
	)

	if numValues < numRows {
		// Fewer items than requested number of rows, actual number of rows = number of items
		numRows = numValues

		// Each row has 1 item, no remainder
		numItems, rmdr = 1, 0
	}

	for i := 0; i < numRows; i++ {
		// start, end = indexes for a subslice of values for this row
		end = start + numItems
		if rmdr > 0 {
			// Add one extra item from remainder
			end++
			rmdr--
		}
		split = append(split, values[start:end])

		// next row start index is current row end index (start is inclusive, end is exclusive)
		start = end
	}

	return split
}

// SplitIntoColumnsOf is a version of SplitIntoColumns where the slice type is the same as the type of the given value.
// This operation will exhaust the iter.
// Panics if the iter has already been exhausted.
// Panics if rows = 0.
// Panics if value is nil.
// Panics if any value is not convertible to the type of the given value.
func (it *Iter) SplitIntoColumnsOf(rows uint, value interface{}) interface{} {
	if rows == 0 {
		panic(ErrRowsGreaterThanZero)
	}

	if value == nil {
		panic(ErrValueCannotBeNil)
	}

	// Collect values into a slice first
	var (
		values         = it.ToSlice()
		numValues      = len(values)
		numRows        = int(rows)
		numItems, rmdr = numValues / numRows, numValues % numRows
		start, end     int
		typ            = reflect.TypeOf(value)
	)

	if numValues < numRows {
		// Fewer items than requested number of rows, actual number of rows = number of items
		numRows = numValues

		// Each row has 1 item, no remainder
		numItems, rmdr = 1, 0
	}

	// Allocate number of rows now we know for sure how many there are
	split := reflect.MakeSlice(reflect.SliceOf(reflect.SliceOf(typ)), numRows, numRows)

	for i := 0; i < numRows; i++ {
		// start, end = indexes for a subslice of values for this row
		end = start + numItems
		if rmdr > 0 {
			// Add one extra item from remainder
			end++
			rmdr--
		}

		row := reflect.MakeSlice(reflect.SliceOf(typ), end-start, end-start)
		for j, colIdx := start, 0; j < end; j, colIdx = j+1, colIdx+1 {
			row.Index(colIdx).Set(reflect.ValueOf(values[j]).Convert(typ))
		}
		split.Index(i).Set(row)

		// next row start index is current row end index (start is inclusive, end is exclusive)
		start = end
	}

	return split.Interface()
}

// ToSlice collects the elements into a slice
func (it *Iter) ToSlice() []interface{} {
	slice := []interface{}{}

	for it.Next() {
		slice = append(slice, it.Value())
	}

	return slice
}

// ToSliceOf returns a slice of all elements, where the slice type is the same as the type of the given value.
// EG, if a value of type int is passed, a []int is returned.
// Panics if value is nil.
// Panics if any value is not convertible to the type of the given value.
func (it *Iter) ToSliceOf(value interface{}) interface{} {
	if value == nil {
		panic(ErrValueCannotBeNil)
	}

	var (
		typ   = reflect.TypeOf(value)
		slice = reflect.MakeSlice(reflect.SliceOf(typ), 0, 0)
	)

	for it.Next() {
		slice = reflect.Append(slice, reflect.ValueOf(it.Value()).Convert(typ))
	}

	return slice.Interface()
}

// ==== Iterable

// Iterable is a generator of Iter that allows iterating the same data structure any number of times.
// There are no thread safety guarantees in the generated Iter instances returned.
// The generator may be replaced with a new generator at any time, to iterate a different data structure.
type Iterable struct {
	Generator func() func() (interface{}, bool)
}

// IterableArraySliceFunc is an iterating function generator for an array or slice
func IterableArraySliceFunc(items interface{}) func() func() (interface{}, bool) {
	return func() func() (interface{}, bool) { return ArraySliceIterFunc(reflect.ValueOf(items)) }
}

// IterableFlattenFunc is an iterating function generator that flattens a multi-dimensional array or slice into a single dimension
func IterableFlattenFunc(items interface{}) func() func() (interface{}, bool) {
	if items == nil {
		// Can't call reflect.ValueOf(nil)
		return func() func() (interface{}, bool) { return NoValueIterFunc }
	}

	return func() func() (interface{}, bool) {
		return ArraySliceIterFunc(reflect.ValueOf(FlattenArraySlice(items)))
	}
}

// IterableElementsFunc is an iterating function generator that iterates the elements of the value passed.
// See ElementsIterFunc for the types it handles.
func IterableElementsFunc(items interface{}) func() func() (interface{}, bool) {
	if items == nil {
		// Can't call reflect.ValueOf(nil)
		return func() func() (interface{}, bool) { return NoValueIterFunc }
	}

	return func() func() (interface{}, bool) { return ElementsIterFunc(reflect.ValueOf(items)) }
}

// IterablesFunc is an iterating function generator that iterates the values of any number of Iterables in the order passed.
// As each Iterable.Iter() is exhausted, the next Iterable.Iter() is used.
// If the slice is empty, the generated Iter will immediately return false on the first call to Next().
func IterablesFunc(iterables []*Iterable) func() func() (interface{}, bool) {
	return func() func() (interface{}, bool) {
		var (
			num         = len(iterables)
			idx         = 0
			theIterable *Iterable
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

			// Search any remaining iters for the next non-empty generated *Iter, if any
			for idx < num {
				theIterable = iterables[idx]
				idx++
				if theIterable != nil {
					if theIter = theIterable.Iter(); theIter.Next() {
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
}

// IterableOfGenerator constructs an Iterable from an iterating function generator.
// Panics if the provided generator is nil.
func IterableOfGenerator(generator func() func() (interface{}, bool)) *Iterable {
	if generator == nil {
		panic(ErrIterableGeneratorCannotBeNil)
	}

	return &Iterable{Generator: generator}
}

// IterableOf constructs an Iterable from the items passed.
// See IterableArraySliceFunc.
func IterableOf(items ...interface{}) *Iterable {
	return &Iterable{Generator: IterableArraySliceFunc(items)}
}

// IterableOfFlatten constructs an Iterable that iterates a multi-dimensional array or slice flattened into a new one-dimensional slice.
// See IterableFlattenFunc.
func IterableOfFlatten(items interface{}) *Iterable {
	return &Iterable{Generator: IterableFlattenFunc(items)}
}

// IterableOfElements constructs an Iterable that iterates the elements of the item passed.
// See IterableElementsFunc.
func IterableOfElements(items interface{}) *Iterable {
	return &Iterable{Generator: IterableElementsFunc(items)}
}

// IterableOfIterables constructs an Iterable that iterates the iterables passed.
// See IterablesFunc.
func IterableOfIterables(iterables ...*Iterable) *Iterable {
	return &Iterable{Generator: IterablesFunc(iterables)}
}

// Iter generates a new Iter instance for the underlying data structure.
// Panics if the generator is nil, or the generators returns a nil *Iter.
func (it Iterable) Iter() *Iter {
	if it.Generator == nil {
		panic(ErrIterableGeneratorCannotBeNil)
	}

	iterFunc := it.Generator()
	if iterFunc == nil {
		panic(ErrIterableGeneratorCannotReturnNil)
	}

	return NewIter(iterFunc)
}

// Of replaces the generator with a new generator that iterates the items passed.
// See IterableOf constructor function.
func (it *Iterable) Of(items ...interface{}) {
	it.Generator = IterableArraySliceFunc(items)
}

// OfFlatten replaces the generator with a new generator that iterates a multi-dimensional array or slice flattened into a new one-dimensional slice.
// See IterableFlatten constructor function.
func (it *Iterable) OfFlatten(items interface{}) {
	it.Generator = IterableFlattenFunc(items)
}

// OfElements replaces the generator with a new generator that iterates the elements of the item passed.
// See IterableElements constructor function.
func (it *Iterable) OfElements(items interface{}) {
	it.Generator = IterableElementsFunc(items)
}

// OfIterables replaces the generator with a new generator that iterates the iterables passed.
// See Iterables constructor function.
func (it *Iterable) OfIterables(items ...*Iterable) {
	it.Generator = IterablesFunc(items)
}

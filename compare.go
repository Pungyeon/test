package test

import (
	"reflect"
	"testing"
)

var evalTypeFn map[reflect.Kind]func(a, b reflect.Value) error

func init() {
	evalTypeFn = map[reflect.Kind]func(a, b reflect.Value) error{
		reflect.String: isStringEqual,
		reflect.Bool: isBoolEqual,
		reflect.Struct: isStructEqual,
		reflect.Map: isMapEqual,
		reflect.Ptr: isPointerDeepEqual,
		reflect.Interface: isInterfaceEqual,

		// Integer
		reflect.Int:    isIntEqual,
		reflect.Int8:   isIntEqual,
		reflect.Int16:  isIntEqual,
		reflect.Int32:  isIntEqual,
		reflect.Int64:  isIntEqual,
		reflect.Uint:   isIntEqual,
		reflect.Uint8:  isIntEqual,
		reflect.Uint16: isIntEqual,
		reflect.Uint32: isIntEqual,
		reflect.Uint64: isIntEqual,

		// Float
		reflect.Float32: isFloatEqual,
		reflect.Float64: isFloatEqual,

		// Complex
		reflect.Complex64:  isComplexEqual,
		reflect.Complex128: isComplexEqual,

		// Slice & Array
		reflect.Slice: isSliceEqual,
		reflect.Array:         isSliceEqual,

		// Pointers
		reflect.Func:   isPointerEqual,
		reflect.UnsafePointer: isPointerEqual,
		reflect.Chan:    isPointerEqual,
		reflect.Uintptr:       isPointerEqual,

		// Unsupported
		reflect.Invalid: unsupported,
	}
}

func isInterfaceEqual(a, b reflect.Value) error {
	return equal(a.Elem(), b.Elem())
}

func isPointerDeepEqual(a, b reflect.Value) error {
	return equal(refToVal(a), refToVal(b))
}

func isMapEqual(a, b reflect.Value) error {
	for _, key := range a.MapKeys() {
		if err := equal(a.MapIndex(key), b.MapIndex(key)); err != nil {
			return err
		}
	}
	return nil
}

func isStructEqual(a, b reflect.Value) error {
	for i := 0; i < a.NumField(); i++ {
		if err := equal(a.Field(i), b.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func isBoolEqual(a, b reflect.Value) error {
	return isEqual(a.Bool() == b.Bool(), a, b)
}

func isStringEqual(a ,b reflect.Value) error {
	return isEqual(a.String() == b.String(), a, b)
}

type Comparison struct {
	t *testing.T
}

func NewComparison() Comparison {
	return Comparison{}
}

func (c Comparison) Equal(a, b interface{}) error {
	return equal(reflect.ValueOf(a), reflect.ValueOf(b))
}

func equal(a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return NewError(ErrDifferingTypes, DifferentTypesFmt, a.Kind(), a, b.Kind(), b)
	}
	fn, ok := evalTypeFn[a.Kind()]
	if !ok {
		return NewError(ErrUnsupportedType, "(kind: %v, type: %v, value: %v)", a.Kind(), a.Type(), a)
	}
	return fn(a, b)
}

func isEqual(statement bool, a, b reflect.Value) error {
	if !statement {
		return NewError(ErrNotEqual, "%v != %v", a, b)
	}
	return nil
}

func isIntEqual(a, b reflect.Value) error {
	return isEqual(a.Int() == b.Int(), a, b)
}

func isFloatEqual(a, b reflect.Value) error {
	return isEqual(a.Float() == b.Float(), a, b)
}

func isComplexEqual(a, b reflect.Value) error {
	return isEqual(a.Complex() == b.Complex(), a, b)
}

func isSliceEqual(a, b reflect.Value) error {
	for i := 0; i < a.Len(); i++ {
		if err := equal(a.Index(i), b.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func isPointerEqual(a, b reflect.Value) error {
	return isEqual(a.Pointer() == b.Pointer(), a, b)
}

func refToVal(a reflect.Value) reflect.Value {
	for a.Kind() == reflect.Ptr {
		a = a.Elem()
	}
	return a
}

func unsupported(a, b reflect.Value) error {
	return NewError(ErrUnsupportedType, "(kind: %v, type: %v, value: %v)", a.Kind(), a.Type(), a)
}


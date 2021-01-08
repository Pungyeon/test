package test

import (
	"reflect"
	"testing"
)

type Comparison struct {
	t *testing.T
}

func NewComparison(t *testing.T) Comparison {
	return Comparison{
		t: t,
	}
}

func isIntEqual(a, b reflect.Value) error {
	return isEqual(a.Int() == b.Int(), a, b)
}

func isFloatEqual(a, b reflect.Value) error {
	return isEqual(a.Float() == b.Float(), a, b)
}

var evalTypeFn = map[reflect.Kind]func(a, b reflect.Value) error{
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

	reflect.String: func(a, b reflect.Value) error {
		return isEqual(a.String() == b.String(), a, b)
	},
	reflect.Bool: func(a, b reflect.Value) error {
		return isEqual(a.Bool() == b.Bool(), a, b)
	},
	reflect.Array:         unsupported,
	reflect.Slice:         unsupported,
	reflect.Struct:        unsupported,
	reflect.Ptr:           unsupported,
	reflect.Invalid:       unsupported,
	reflect.Uintptr:       unsupported,
	reflect.Complex64:     unsupported,
	reflect.Complex128:    unsupported,
	reflect.Chan:          unsupported,
	reflect.Func:          unsupported,
	reflect.Interface:     unsupported,
	reflect.Map:           unsupported,
	reflect.UnsafePointer: unsupported,
}

func unsupported(a, b reflect.Value) error {
	return ErrUnsupportedType
}

func isEqual(statement bool, a, b reflect.Value) error {
	if !statement {
		return NewError(ErrNotEqual, "%v != %v", a, b)
	}
	return nil
}

func (c Comparison) Equal(a, b interface{}) error {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	fn, ok := evalTypeFn[va.Kind()]
	if !ok {
		return NewError(ErrUnsupportedType, "(kind: %v, type: %v, value: %v)", va.Kind(), va.Type(), a)
	}
	return fn(va, vb)
}

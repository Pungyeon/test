package test

import (
	"fmt"
	"reflect"
	"testing"
)

var evalTypeFn map[reflect.Kind]func(a, b reflect.Value) Result

func init() {
	evalTypeFn = map[reflect.Kind]func(a, b reflect.Value) Result{
		reflect.String:    isStringEqual,
		reflect.Bool:      isBoolEqual,
		reflect.Struct:    isStructEqual,
		reflect.Map:       isMapEqual,
		reflect.Ptr:       isPointerDeepEqual,
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
		reflect.Array: isSliceEqual,

		// Pointers
		reflect.Func:          isPointerEqual,
		reflect.UnsafePointer: isPointerEqual,
		reflect.Chan:          isPointerEqual,
		reflect.Uintptr:       isPointerEqual,

		// Unsupported
		reflect.Invalid: unsupported,
	}
}

type Result struct {
	Error    error
	a        reflect.Value
	b        reflect.Value
	Children []Result
	Field    string
}

func (r Result) Cmp() string {
	if r.Error != nil {
		return "!="
	}
	return "=="
}

func (r Result) String() string {
	return fmt.Sprintf("%s: (%v) %v %v %v", r.Field, r.a.Kind(), r.a, r.Cmp(), r.b)
}

func (r Result) Print(tabs string) error {
	if r.Error != nil {
		return r.Error
	}
	if len(r.Children) > 0 {
		fmt.Printf("(%v::%v)\n", r.a.Type().PkgPath(), r.a.Type().Name())
	}
	for _, child := range r.Children {
		if len(child.Children) == 0 {
			fmt.Println(tabs + child.String())
		} else {
			fmt.Printf("%s: ", child.Field)
		}
		if err := child.Print(tabs + "\t"); err != nil {
			return err
		}
	}
	return nil
}

type Comparison struct {
	t *testing.T
}

func NewComparison() Comparison {
	return Comparison{}
}

func (c Comparison) Equal(a, b interface{}) error {
	result := equal(reflect.ValueOf(a), reflect.ValueOf(b))
	return result.Print("")
}

func equal(a, b reflect.Value) Result {
	if a.Kind() != b.Kind() {
		return Result{
			Error: NewError(ErrDifferingTypes, DifferentTypesFmt, a.Kind(), a, b.Kind(), b),
		}
	}
	fn, ok := evalTypeFn[a.Kind()]
	if !ok {
		return Result{
			Error: NewError(ErrUnsupportedType, "(kind: %v, type: %v, value: %v)", a.Kind(), a.Type(), a),
		}
	}
	return fn(a, b)
}

func isEqual(statement bool, a, b reflect.Value) Result {
	if !statement {
		return Result{
			Error: NewError(ErrNotEqual, "%v != %v", a, b),
			a:     a,
			b:     b,
		}
	}
	return Result{a: a, b: b}
}

func isIntEqual(a, b reflect.Value) Result {
	return isEqual(a.Int() == b.Int(), a, b)
}

func isFloatEqual(a, b reflect.Value) Result {
	return isEqual(a.Float() == b.Float(), a, b)
}

func isComplexEqual(a, b reflect.Value) Result {
	return isEqual(a.Complex() == b.Complex(), a, b)
}

func isSliceEqual(a, b reflect.Value) Result {
	for i := 0; i < a.Len(); i++ {
		result := equal(a.Index(i), b.Index(i))
		if result.Error != nil {
			return result
		}
	}
	return Result{}
}

func isPointerEqual(a, b reflect.Value) Result {
	return isEqual(a.Pointer() == b.Pointer(), a, b)
}

func isInterfaceEqual(a, b reflect.Value) Result {
	return equal(a.Elem(), b.Elem())
}

func isPointerDeepEqual(a, b reflect.Value) Result {
	return equal(refToVal(a), refToVal(b))
}

func isMapEqual(a, b reflect.Value) Result {
	for _, key := range a.MapKeys() {
		result := equal(a.MapIndex(key), b.MapIndex(key))
		if result.Error != nil {
			return result
		}
	}
	return Result{}
}

func isStructEqual(a, b reflect.Value) Result {
	result := Result{a: a, b: b}
	for i := 0; i < a.NumField(); i++ {
		r := equal(a.Field(i), b.Field(i))
		r.Field = a.Type().Field(i).Name
		result.Children = append(result.Children, r)
		if r.Error != nil {
			return result
		}
	}
	return result
}

func isBoolEqual(a, b reflect.Value) Result {
	return isEqual(a.Bool() == b.Bool(), a, b)
}

func isStringEqual(a, b reflect.Value) Result {
	return isEqual(a.String() == b.String(), a, b)
}
func refToVal(a reflect.Value) reflect.Value {
	for a.Kind() == reflect.Ptr {
		a = a.Elem()
	}
	return a
}

func unsupported(a, b reflect.Value) Result {
	return Result{
		Error: NewError(ErrUnsupportedType, "(kind: %v, type: %v, value: %v)", a.Kind(), a.Type(), a),
	}
}

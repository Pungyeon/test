package test

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

var (
	colorReset  = "\033[0m"
	colorGrey   = "\033[90m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
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
	Type     string
}

func (r Result) AssertString() string {
	if r.Error == nil {
		return "OK"
	}
	return "FAIL"
}

func (r Result) Cmp() (string, string) {
	if r.Error != nil {
		return colorRed, "!="
	}
	return colorReset, "=="
}

func (r Result) ComparisonString() string {
	color, comparator := r.Cmp()
	return fmt.Sprintf("%s(%v)%s %v %v %v",
		colorGrey, r.a.Kind(), color, r.a, comparator, r.b)
}

// This should take a writer, instead of just writing to stdout
func (r Result) Print(w io.Writer, debug bool, tabs string) error {
	if debug == false && r.Error == nil {
		return nil
	}
	if len(r.Children) == 0 {
		_, _ = w.Write([]byte(r.ComparisonString() + "\n"))
		return r.Error
	}
	_, _ = w.Write([]byte(fmt.Sprintf("%s%s%s[%v]\n", colorYellow, r.Type, colorGrey, r.AssertString())))
	for _, child := range r.Children {
		if debug == false && child.Error == nil {
			continue
		}
		_, _ = w.Write([]byte(fmt.Sprintf("%s%s%s: ", tabs, colorCyan, child.Field)))
		if err := child.Print(w, debug, tabs+"\t"); err != nil {
			return err
		}
	}
	return nil
}

type Option func(*Comparison)

func WithDebug() Option {
	return func(c *Comparison) {
		c.debug = true
	}
}

func WithWriter(w io.Writer) Option {
	return func(c *Comparison) {
		c.writer = w
	}
}

type Comparison struct {
	debug  bool
	writer io.Writer
}

func DefaultComparison() Comparison {
	return Comparison{
		debug:  false,
		writer: os.Stdout,
	}
}

func NewComparison(opts ...Option) Comparison {
	c := DefaultComparison()

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

func (c Comparison) Equal(a, b interface{}) error {
	result := equal(reflect.ValueOf(a), reflect.ValueOf(b))
	return result.Print(c.writer, c.debug, "")
}

func equal(a, b reflect.Value) Result {
	if a.Kind() != b.Kind() {
		return Result{
			a: a, b: b,
			Error: NewError(ErrDifferingTypes, DifferentTypesFmt, a.Kind(), a, b.Kind(), b),
		}
	}
	fn, ok := evalTypeFn[a.Kind()]
	if !ok {
		return Result{
			a: a, b: b,
			Error: NewError(ErrUnsupportedType, "(kind: %v, type: %v, value: %v)", a.Kind(), a.Type(), a),
		}
	}
	return fn(a, b)
}

func isEqual(statement bool, a, b reflect.Value) Result {
	if !statement {
		return Result{
			a: a, b: b,
			Error: NewError(ErrNotEqual, "%v != %v", a, b),
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
	result := Result{
		a: a, b: b,
		Type: fmt.Sprintf("(%v)", a.Type().String()),
	}
	for i := 0; i < a.Len(); i++ {
		r := equal(a.Index(i), b.Index(i))
		r.Field = strconv.FormatInt(int64(i), 10)
		result.Children = append(result.Children, r)
		if r.Error != nil {
			result.Error = r.Error
			return result
		}
	}
	return result
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
	result := Result{a: a, b: b,
		Type: fmt.Sprintf("(%v)", a.Type().String()),
	}
	for _, key := range a.MapKeys() {
		r := equal(a.MapIndex(key), b.MapIndex(key))
		r.Field = key.String()
		result.Children = append(result.Children, r)
		if r.Error != nil {
			result.Error = r.Error
			return result
		}
	}
	return result
}

func isStructEqual(a, b reflect.Value) Result {
	result := Result{
		a: a, b: b,
		Type: fmt.Sprintf("(%v::%v)", a.Type().PkgPath(), a.Type().Name()),
	}

	for i := 0; i < a.NumField(); i++ {
		r := equal(a.Field(i), b.Field(i))
		r.Field = a.Type().Field(i).Name
		result.Children = append(result.Children, r)
		if r.Error != nil {
			result.Error = r.Error
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

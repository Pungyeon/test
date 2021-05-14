package test

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"testing"
)

var (
	colorReset  = "\033[0m"
	colorGrey   = "\033[90m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

type equalityEvalMap map[reflect.Kind]func(a, b reflect.Value) CmpResult

func newEqualityEvalMap(assertion *Assertion) equalityEvalMap {
	return map[reflect.Kind]func(a, b reflect.Value) CmpResult{
		reflect.String:    isStringEqual,
		reflect.Bool:      isBoolEqual,
		reflect.Struct:    assertion.isStructEqual,
		reflect.Map:       assertion.isMapEqual,
		reflect.Ptr:       assertion.isPointerDeepEqual,
		reflect.Interface: assertion.isInterfaceEqual,

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
		reflect.Slice: assertion.isSliceEqual,
		reflect.Array: assertion.isSliceEqual,

		// Pointers
		reflect.Func:          isPointerEqual,
		reflect.UnsafePointer: isPointerEqual,
		reflect.Chan:          isPointerEqual,
		reflect.Uintptr:       isPointerEqual,

		// Unsupported
		reflect.Invalid: unsupported,
	}
}

type CmpResult struct {
	Error    error
	a        reflect.Value
	b        reflect.Value
	Children []CmpResult
	Field    string
	Type     string
}

func (r CmpResult) AssertString() string {
	if r.Error == nil {
		return "OK"
	}
	return "FAIL"
}

func (r CmpResult) Cmp() (string, string) {
	if r.Error != nil {
		return colorRed, "!="
	}
	return colorReset, "=="
}

func (r CmpResult) ComparisonString() string {
	color, comparator := r.Cmp()
	return fmt.Sprintf("%s(%v)%s %v %v %v",
		colorGrey, r.a.Kind(), color, r.a, comparator, r.b)
}

// This should take a writer, instead of just writing to stdout
func (r CmpResult) Print(w io.Writer, debug bool, tabs string) error {
	defer fmt.Printf(colorReset)
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

type Option func(*Assertion)

func WithDebug() Option {
	return func(c *Assertion) {
		c.debug = true
	}
}

func WithWriter(w io.Writer) Option {
	return func(c *Assertion) {
		c.writer = w
	}
}

func WithIgnoredFields(fields ...string) Option {
	return func(c *Assertion) {
		for _, field := range fields {
			c.ignore[field] = true
		}
	}
}

type Assertion struct {
	testing testing.TB
	writer  io.Writer
	ignore  map[string]bool
	debug   bool
	evalMap equalityEvalMap
}

func DefaultAssertion(t testing.TB) Assertion {
	assertion := Assertion{
		testing: t,
		debug:   false,
		writer:  os.Stdout,
		ignore:  make(map[string]bool),
	}

	assertion.evalMap = newEqualityEvalMap(&assertion)
	return assertion
}

func NewAssertion(t testing.TB, opts ...Option) Assertion {
	c := DefaultAssertion(t)

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

func (c Assertion) Size(v interface{}, expected int) {
	val := reflect.ValueOf(v)
	if val.Len() != expected {
		c.testing.Fatalf("unexpected size of %v: (expected) %v != %v (actual)",
			val.Type().String(), val.Len(), expected)
	}
}

func (c Assertion) NotNil(err error) {
	if err != nil {
		c.testing.Fatal(err)
	}
}

func (c Assertion) Equal(a, b interface{}) error {
	return c.equal(reflect.ValueOf(a), reflect.ValueOf(b)).
		Print(c.writer, c.debug, "")
}

func (c *Assertion) equal(a, b reflect.Value) CmpResult {
	if a.Kind() != b.Kind() {
		return CmpResult{
			a: a, b: b,
			Error: NewError(ErrDifferingTypes, DifferentTypesFmt, a.Kind(), a, b.Kind(), b),
		}
	}
	fn, ok := c.evalMap[a.Kind()]
	if !ok {
		return CmpResult{
			a: a, b: b,
			Error: NewError(ErrUnsupportedType, "(kind: %v, type: %v, value: %v)", a.Kind(), a.Type(), a),
		}
	}
	return fn(a, b)
}

func isEqual(statement bool, a, b reflect.Value) CmpResult {
	if !statement {
		return CmpResult{
			a: a, b: b,
			Error: NewError(ErrNotEqual, "%v != %v", a, b),
		}
	}
	return CmpResult{a: a, b: b}
}

func isIntEqual(a, b reflect.Value) CmpResult {
	return isEqual(a.Int() == b.Int(), a, b)
}

func isFloatEqual(a, b reflect.Value) CmpResult {
	return isEqual(a.Float() == b.Float(), a, b)
}

func isComplexEqual(a, b reflect.Value) CmpResult {
	return isEqual(a.Complex() == b.Complex(), a, b)
}

func (c *Assertion) isSliceEqual(a, b reflect.Value) CmpResult {
	result := CmpResult{
		a: a, b: b,
		Type: fmt.Sprintf("(%v)", a.Type().String()),
	}
	for i := 0; i < a.Len(); i++ {
		r := c.equal(a.Index(i), b.Index(i))
		r.Field = strconv.FormatInt(int64(i), 10)
		result.Children = append(result.Children, r)
		if r.Error != nil {
			result.Error = r.Error
			return result
		}
	}
	return result
}

func isPointerEqual(a, b reflect.Value) CmpResult {
	return isEqual(a.Pointer() == b.Pointer(), a, b)
}

func (c *Assertion) isInterfaceEqual(a, b reflect.Value) CmpResult {
	return c.equal(a.Elem(), b.Elem())
}

func (c *Assertion) isPointerDeepEqual(a, b reflect.Value) CmpResult {
	return c.equal(refToVal(a), refToVal(b))
}

func (c *Assertion) isMapEqual(a, b reflect.Value) CmpResult {
	result := CmpResult{a: a, b: b,
		Type: fmt.Sprintf("(%v)", a.Type().String()),
	}
	for _, key := range a.MapKeys() {
		r := c.equal(a.MapIndex(key), b.MapIndex(key))
		r.Field = key.String()
		result.Children = append(result.Children, r)
		if r.Error != nil {
			result.Error = r.Error
			return result
		}
	}
	return result
}

func (c *Assertion) isStructEqual(a, b reflect.Value) CmpResult {
	result := CmpResult{
		a: a, b: b,
		Type: fmt.Sprintf("(%v::%v)", a.Type().PkgPath(), a.Type().Name()),
	}

	for i := 0; i < a.NumField(); i++ {
		field := fmt.Sprintf("%s::%s", a.Type().Name(), a.Type().Field(i).Name)
		if _, ok := c.ignore[field]; ok {
			fmt.Printf("ignoring: %s\n", field)
			continue
		}
		r := c.equal(a.Field(i), b.Field(i))
		r.Field = a.Type().Field(i).Name
		result.Children = append(result.Children, r)
		if r.Error != nil {
			result.Error = r.Error
			return result
		}
	}
	return result
}

func isBoolEqual(a, b reflect.Value) CmpResult {
	return isEqual(a.Bool() == b.Bool(), a, b)
}

func isStringEqual(a, b reflect.Value) CmpResult {
	return isEqual(a.String() == b.String(), a, b)
}
func refToVal(a reflect.Value) reflect.Value {
	for a.Kind() == reflect.Ptr {
		a = a.Elem()
	}
	return a
}

func unsupported(a, _ reflect.Value) CmpResult {
	return CmpResult{
		Error: NewError(ErrUnsupportedType, "(kind: %v, type: %v, value: %v)", a.Kind(), a.Type(), a),
	}
}

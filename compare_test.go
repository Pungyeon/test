package test

import (
	"bytes"
	"errors"
	"fmt"
	"math/cmplx"
	"sync"
	"testing"
	"time"
)

type TestStruct struct {
	name string
	nest TestNest
}

type TestNest struct {
	age int
}

type TestInterfaceProp struct {
	v interface{}
}

type BigStruct struct {
	Name  string
	Value TestInterfaceProp
	Inner InnerStruct
	Test  TestStruct
}

type InnerStruct struct {
	Name   string
	Values []int
}

func TestStructCompare(t *testing.T) {
	var (
		expectedOut = "\x1b[33m(github.com/pungyeon/test::BigStruct)\x1b[90m[FAIL]\n\x1b[36mInner: \x1b[33m(github.com/pungyeon/test::InnerStruct)\x1b[90m[FAIL]\n\t\x1b[36mValues: \x1b[33m([]int)\x1b[90m[FAIL]\n\t\t\x1b[36m7: \x1b[90m(int)\x1b[31m 7 != 8\n"
		a           = BigStruct{
			Name: "Big Struct",
			Value: TestInterfaceProp{
				v: &TestNest{
					age: 23,
				},
			},
			Inner: InnerStruct{
				Name:   "Inner Struct",
				Values: []int{1, 2, 3, 4, 5, 6, 7, 7, 9, 10},
			},
			Test: TestStruct{
				name: "Test Struct",
				nest: TestNest{
					age: 39,
				},
			},
		}
	)

	out := bytes.Buffer{}
	if err := NewAssertion(t, WithWriter(&out)).Equal(a, BigStruct{
		Name: "Big Struct",
		Value: TestInterfaceProp{
			v: &TestNest{
				age: 23,
			},
		},
		Inner: InnerStruct{
			Name:   "Inner Struct",
			Values: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		Test: TestStruct{
			name: "Test Struct",
			nest: TestNest{
				age: 39,
			},
		},
	}); !errors.Is(err, ErrNotEqual) {
		t.Fatal(err)
	}

	if out.String() != expectedOut {
		t.Fatalf("got: %#v\nexpected: %#v", out.String(), expectedOut)
	}
}

func TestStructCompareDebug(t *testing.T) {
	var (
		expectedOut = "\x1b[33m(github.com/pungyeon/test::BigStruct)\x1b[90m[OK]\n\x1b[36mName: \x1b[90m(string)\x1b[0m Big Struct == Big Struct\n\x1b[36mValue: \x1b[33m(github.com/pungyeon/test::TestInterfaceProp)\x1b[90m[OK]\n\t\x1b[36mv: \x1b[33m(github.com/pungyeon/test::TestNest)\x1b[90m[OK]\n\t\t\x1b[36mage: \x1b[90m(int)\x1b[0m 23 == 23\n\x1b[36mInner: \x1b[33m(github.com/pungyeon/test::InnerStruct)\x1b[90m[OK]\n\t\x1b[36mName: \x1b[90m(string)\x1b[0m Inner Struct == Inner Struct\n\t\x1b[36mValues: \x1b[33m([]int)\x1b[90m[OK]\n\t\t\x1b[36m0: \x1b[90m(int)\x1b[0m 1 == 1\n\t\t\x1b[36m1: \x1b[90m(int)\x1b[0m 2 == 2\n\t\t\x1b[36m2: \x1b[90m(int)\x1b[0m 3 == 3\n\t\t\x1b[36m3: \x1b[90m(int)\x1b[0m 4 == 4\n\t\t\x1b[36m4: \x1b[90m(int)\x1b[0m 5 == 5\n\t\t\x1b[36m5: \x1b[90m(int)\x1b[0m 6 == 6\n\t\t\x1b[36m6: \x1b[90m(int)\x1b[0m 7 == 7\n\t\t\x1b[36m7: \x1b[90m(int)\x1b[0m 8 == 8\n\t\t\x1b[36m8: \x1b[90m(int)\x1b[0m 9 == 9\n\t\t\x1b[36m9: \x1b[90m(int)\x1b[0m 10 == 10\n\x1b[36mTest: \x1b[33m(github.com/pungyeon/test::TestStruct)\x1b[90m[OK]\n\t\x1b[36mname: \x1b[90m(string)\x1b[0m Test Struct == Test Struct\n\t\x1b[36mnest: \x1b[33m(github.com/pungyeon/test::TestNest)\x1b[90m[OK]\n\t\t\x1b[36mage: \x1b[90m(int)\x1b[0m 39 == 39\n"
		a           = BigStruct{
			Name: "Big Struct",
			Value: TestInterfaceProp{
				v: &TestNest{
					age: 23,
				},
			},
			Inner: InnerStruct{
				Name:   "Inner Struct",
				Values: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			Test: TestStruct{
				name: "Test Struct",
				nest: TestNest{
					age: 39,
				},
			},
		}
	)
	var out bytes.Buffer
	if err := NewAssertion(t,
		WithDebug(),
		WithWriter(&out),
	).Equal(a, BigStruct{
		Name: "Big Struct",
		Value: TestInterfaceProp{
			v: &TestNest{
				age: 23,
			},
		},
		Inner: InnerStruct{
			Name:   "Inner Struct",
			Values: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		Test: TestStruct{
			name: "Test Struct",
			nest: TestNest{
				age: 39,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if out.String() != expectedOut {
		t.Fatalf("got: %#v\nexpected: %#v", out.String(), expectedOut)
	}
}

func TestCompare(t *testing.T) {
	var (
		cmp        Assertion   = NewAssertion(t)
		aInterface interface{} = TestStruct{"Lasse", TestNest{23}}
		bInterface interface{} = TestStruct{"Lasse", TestNest{23}}
		cInterface interface{} = TestStruct{"Basse", TestNest{24}}
		aChannel   chan int    = make(chan int)
		bChannel   chan int    = make(chan int)
	)

	tt := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected error
	}{
		{"integer equal", 1, 1, nil},
		{"integer !equal", 1, 2, ErrNotEqual},
		{"float equal", 1.0, 1.0, nil},
		{"float !equal", 1.0, 2.0, ErrNotEqual},
		{"string equal", "Lasse", "Lasse", nil},
		{"string !equal", "Lasse", "Basse", ErrNotEqual},
		{"bool equal", true, true, nil},
		{"bool !equal", true, false, ErrNotEqual},
		{"complex equal", cmplx.Rect(2.0, 2.0), cmplx.Rect(2.0, 2.0), nil},
		{"complex !equal", cmplx.Rect(2.0, 2.0), cmplx.Rect(2.0, 3.0), ErrNotEqual},
		{"slice equal", []int{1, 2, 3}, []int{1, 2, 3}, nil},
		{"slice !equal", []int{1, 2, 3}, []int{1, 2, 4}, ErrNotEqual},
		{"map equal",
			map[string]int{
				"one": 1,
				"two": 2,
			},
			map[string]int{
				"two": 2,
				"one": 1,
			},
			nil},
		{"map !equal",
			map[string]int{
				"one": 1,
				"two": 2,
			},
			map[string]int{
				"one": 1,
				"two": 9,
			},
			ErrNotEqual},
		{"map w. interface equal",
			map[string]interface{}{
				"one": "ichi",
				"two": 2,
			},
			map[string]interface{}{
				"two": 2,
				"one": "ichi",
			},
			nil},
		{"map w. interface !equal",
			map[string]interface{}{
				"one": "ichi",
				"two": 2,
			},
			map[string]interface{}{
				"two": 2,
				"one": "eins",
			},
			ErrNotEqual},
		{"struct equal",
			TestStruct{"Lasse", TestNest{23}},
			TestStruct{"Lasse", TestNest{23}},
			nil},
		{"struct !equal",
			TestStruct{"Lasse", TestNest{23}},
			TestStruct{"Lasse", TestNest{24}},
			ErrNotEqual},
		{"ptr equal",
			&TestStruct{"Lasse", TestNest{23}},
			&TestStruct{"Lasse", TestNest{23}},
			nil},
		{"ptr !equal",
			&TestStruct{"Lasse", TestNest{23}},
			&TestStruct{"Basse", TestNest{23}},
			ErrNotEqual},
		{"interface equal", aInterface, bInterface, nil},
		{"interface !equal", aInterface, cInterface, ErrNotEqual},
		{"struct w. interface equal", TestInterfaceProp{1}, TestInterfaceProp{1}, nil},
		{"struct w. interface !equal", TestInterfaceProp{1}, TestInterfaceProp{2}, ErrNotEqual},
		{"func equal", cmp.equal, cmp.equal, nil},
		{"func !equal", cmp.equal, unsupported, ErrNotEqual},
		{"chan equal", aChannel, aChannel, nil},
		{"chan !equal", aChannel, bChannel, ErrNotEqual},
		{"differing types", 1, "ding", ErrDifferingTypes},
	}

	for _, tf := range tt {
		if err := cmp.Equal(tf.a, tf.b); !errors.Is(err, tf.expected) {
			t.Fatalf("test (%s) failed: %v", tf.name, err)
		}
	}
}

func TestIgnoreField(t *testing.T) {
	var (
		a = BigStruct{
			Name: "Big Struct",
			Value: TestInterfaceProp{
				v: &TestNest{
					age: 23,
				},
			},
			Inner: InnerStruct{
				Name:   "Inner Struct",
				Values: []int{1, 2, 3, 4, 5, 6, 7, 7, 9, 10},
			},
		}

		expected = BigStruct{
			Name: "Big Struct",
			Value: TestInterfaceProp{
				v: &TestNest{
					age: 23,
				},
			},
			Inner: InnerStruct{
				Name:   "Different",
				Values: []int{1, 2, 3, 4, 5, 6, 7, 7, 9, 10},
			},
		}
	)

	if err := NewAssertion(t, WithIgnoredFields("InnerStruct::Name")).
		Equal(a, expected); err != nil {
		t.Fatal(err)
	}
}

type Credentials struct {
	key string
}

var creds *Credentials

var initialiser sync.Once

func Initialise() {
	initialiser.Do(func() {
		creds = &Credentials{
			key: time.Now().String(),
		}
	})
}

func TestContextStuff(t *testing.T) {
	Initialise()
	tmp := creds.key

	Initialise()

	if tmp != creds.key {
		fmt.Println("sync once does not work")
	}
}

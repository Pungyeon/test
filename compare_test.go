package test

import (
	"errors"
	"fmt"
	"math/cmplx"
	"testing"
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
	a := BigStruct{
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
	if err := NewComparison().Equal(a, BigStruct{
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
}

func TestCompare(t *testing.T) {
	var (
		cmp        Comparison  = NewComparison()
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
		{"func equal", equal, equal, nil},
		{"func !equal", equal, unsupported, ErrNotEqual},
		{"chan equal", aChannel, aChannel, nil},
		{"chan !equal", aChannel, bChannel, ErrNotEqual},
		{"differing types", 1, "ding", ErrDifferingTypes},
	}

	for _, tf := range tt {
		fmt.Println()
		if err := cmp.Equal(tf.a, tf.b); !errors.Is(err, tf.expected) {
			t.Fatalf("test (%s) failed: %v", tf.name, err)
		}
	}
}

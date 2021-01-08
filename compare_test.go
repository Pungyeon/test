package test

import (
	"errors"
	"testing"
)

func TestCompare(t *testing.T) {
	cmp := NewComparison(t)

	tt := []struct{
		name string
		a interface{}
		b interface{}
		expected error
	}{
		{"integer equal", 1, 1, nil},
		{"integer !equal", 1, 2, ErrNotEqual},
		{"float equal", 1.0, 1.0, nil},
		{"float !equal", 1.0, 2.0, ErrNotEqual},
		{"string equal", 1.0, 1.0, nil},
		{"string !equal", 1.0, 2.0, ErrNotEqual},
		{"bool equal", true, true, nil},
		{"bool !equal", true, false, ErrNotEqual},
	}

	for _, tf := range tt {
		if err := cmp.Equal(tf.a, tf.b); !errors.Is(err, tf.expected) {
			t.Fatal(err)
		}
	}
}
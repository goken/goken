package test

import (
	"jxck/assert"
	"testing"
)

func TestInt(t *testing.T) {
	type I int
	var actual I = 1
	expected := 2
	assert.Equal(t, actual, expected)
}

func TestString(t *testing.T) {
	type S string
	var actual S = "aaa"
	expected := "bbb"
	assert.Equal(t, actual, expected)
}

func TestStruct(t *testing.T) {
	type Foo struct {
		name string
		age  int
	}

	actual := Foo{"john", 20}
	expected := Foo{"emily", 22}
	assert.Equal(t, actual, expected)
}

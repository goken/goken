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

func TestBool(t *testing.T) {
	type B bool
	var actual B = false
	expected := true
	assert.Equal(t, actual, expected)
}

func TestSlice(t *testing.T) {
	type S []int
	var actual S = []int{1, 2, 3}
	expected := []int{1, 2, 3, 4}
	assert.Equal(t, actual, expected)
}

func TestStruct(t *testing.T) {
	type Foo struct {
		name  string
		age   int
		veget bool
		lang  []string
	}

	actual := Foo{"john", 20, true, []string{"ja", "en"}}
	expected := Foo{"emily", 22, false, []string{"ja", "en", "ch"}}

	assert.Equal(t, actual, expected)
}

func TestNestedStruct(t *testing.T) {
	type Foo struct {
		name  string
		age   int
		veget bool
		lang  []string
	}

	john := Foo{"john", 20, true, []string{"ja", "en"}}
	emily := Foo{"emily", 22, false, []string{"ja", "en", "ch"}}

	type Bar struct {
		Foo
		class int
	}

	actual := Bar{john, 1}
	expected := Bar{emily, 1}

	assert.Equal(t, actual, expected)
}

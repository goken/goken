package main

import (
	"testing"
)

func TestSelectFrom(t *testing.T) {
	actual := Select("id").From("employee").Build()

	expected := "SELECT id FROM employee"
	if actual != expected {
		t.Errorf("Actual: %q, Expected; %q", actual, expected)
	}
}

func TestSelectFromWhere(t *testing.T) {
	actual := Select("id").From("employee").Where("age == 20").Build()

	expected := "SELECT id FROM employee WHERE age == 20"
	if actual != expected {
		t.Errorf("Actual: %q, Expected; %q", actual, expected)
	}
}

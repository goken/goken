package assert

import (
	"fmt"
	"strings"
	"testing"
)

var values []*capturedIdent

func takeAll() (capturedIdents []*capturedIdent) {
	capturedIdents = values
	values = make([]*capturedIdent, 0)
	return
}

type capturedIdent struct {
	value interface {}
	name string
}

func Capt(value interface{}, name string) interface{} {
	captured := new(capturedIdent)
	captured.name = name
	captured.value = value
	values = append(values, captured)
	return value
}

func PowerOk(t *testing.T, result bool, original string) {
	capturedIdents := takeAll()
	if !result {
		pairs := make([]string, len(capturedIdents))
		for i, c := range capturedIdents {
			pairs[i] = fmt.Sprintf("   %s => %v", c.name, c.value)
		}
		t.Errorf("Assertion Failed.\nASSERT:\n   %s\nWHERE:\n%s", original, strings.Join(pairs, "\n"))
	}
}

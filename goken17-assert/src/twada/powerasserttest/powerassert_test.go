package powerasserttest

import (
	"testing"
	"twada/power/assert"
)

func TestNotPowered(t *testing.T) {
	hoge := "hoge"
	fuga := "fuga"
	assert.Ok(t, hoge == fuga)
}

func TestEmpoweredAssert(t *testing.T) {
	hoge := "foo"
	fuga := "bar"
	assert.PowerOk(t, assert.Capt(hoge, "hoge") == assert.Capt(fuga, "fuga"), "powerassert.Ok(t, hoge == fuga)")
}

func TestEmpoweredAssertWithLiteral(t *testing.T) {
	piyo := "toto"
	assert.PowerOk(t, assert.Capt(piyo, "piyo") == "fuga", "powerassert.Ok(t, piyo == \"fuga\")")
}

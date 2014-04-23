package testee

import (
	"hogedigo/assert"
	"hogedigo/assert/is"
	"testing"
)

func TestOk(t *testing.T) {
	assert.That(t, IntValue(9), is.EqualTo(9))
	assert.That(t, FloatValue(99.9), is.EqualTo(99.9))
	assert.That(t, StrValue("aaa"), is.EqualTo("aaa"))
	assert.That(t, StructValue("aaa", "bbb"), is.EqualTo(struct{ a, b string }{"aaa", "bbb"}))
	assert.That(t, StrValue("aaa"), is.EqualTo("aaa"))
	assert.That(t, StrValue("hello gopher!!"), is.Contains("gopher"))
	assert.That(t, StrValue("hello gopher!!"), is.Contains("gopher").And(is.Contains("hello")))
	assert.That(t, StrValue("hello gopher!!"), is.Contains("gopher").And(
		is.Contains("hoge").Or(is.Contains("hello"))))
}

func TestIntValueNg(t *testing.T) {
	assert.That(t, IntValue(9), is.EqualTo(10))
}

func TestFloatValueNg(t *testing.T) {
	assert.That(t, FloatValue(9.876), is.EqualTo(9.875))
}

func TestStrValueNg(t *testing.T) {
	assert.That(t, StrValue("aaa"), is.EqualTo("aab"))
}

func TestStrValueContainsNg(t *testing.T) {
	assert.That(t, StrValue("hello gopher!!"), is.Contains("golang"))
}

func TestStructValueNg(t *testing.T) {
	assert.That(t, StructValue("aaa", "bbb"), is.EqualTo(struct{ a, b string }{"aaa", "bbc"}))
}

func TestLogical(t *testing.T) {
	assert.That(t, StrValue("hello gopher!!"), is.Contains("gopher").And(
		is.Contains("hoge").Or(is.Contains("moke"))))
}

package instrumentor

import (
	"testing"
	"go/ast"
	"go/parser"
	"go/token"
)

func inteq(t *testing.T, actual, expected int, message string) {
	if actual != expected {
		t.Errorf("\nexpected: %v\n     got: %v\n%s", expected, actual, message)
	}
}

func streq(t *testing.T, actual, expected string, message string) {
	if actual != expected {
		t.Errorf("\nexpected: %v\n     got: %v\n%s", expected, actual, message)
	}
}

func ok(t *testing.T, actual bool, message string) {
	if !actual {
		t.Errorf("failed: %s", message)
	}
}

func parse(x string) ast.Expr {
	expr, err := parser.ParseExpr(x)
	if err != nil {
		panic(err)
	}
	return expr
}


func TestSingleIdent(t *testing.T) {
	expr := parse("assert.Ok(t, val)")
	v := NewVisitor()
//	ast.Print(token.NewFileSet(), expr)
	ast.Walk(v, expr)
	inteq(t, v.nodeStack.Count(), 0, "stack")
	ok(t, !v.capturing, "capturing")
	streq(t, v.original, "", "original")
	result,_ := ToCode(expr)
	streq(t, result, "assert.PowerOk(t, assert.Capt(val, \"val\"), \"assert.Ok(t, val)\")", "output")
}


func TestBinaryExpression(t *testing.T) {
	expr := parse("assert.Ok(t, foo == bar)")
	v := NewVisitor()
	ast.Walk(v, expr)
	inteq(t, v.nodeStack.Count(), 0, "stack")
	ok(t, !v.capturing, "capturing")
	streq(t, v.original, "", "original")
	result,_ := ToCode(expr)
	streq(t, result, "assert.PowerOk(t, assert.Capt(foo, \"foo\") == assert.Capt(bar, \"bar\"), \"assert.Ok(t, foo == bar)\")", "output")
}


func TestAstSpike(t *testing.T) {
	src := `package p

import (
	"testing"
	"power/assert"
)

func TestTargetMethod(t *testing.T) {
	hoge := "foo"
	fuga := "bar"
	assert.Ok(t, hoge == fuga)
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(err)
	}
	// ast.Print(fset, f)

	v := NewVisitor()
	ast.Walk(v, f)

	inteq(t, v.nodeStack.Count(), 0, "stack")
	ok(t, !v.capturing, "capturing")
	streq(t, v.original, "", "original")

	expected := `package p

import (
	"testing"
	"power/assert"
)

func TestTargetMethod(t *testing.T) {
	hoge := "foo"
	fuga := "bar"
	assert.PowerOk(t, assert.Capt(hoge, "hoge") == assert.Capt(fuga, "fuga"), "assert.Ok(t, hoge == fuga)")
}
`
	result,_ := ToCode(f)
	streq(t, result, expected, "output")

}

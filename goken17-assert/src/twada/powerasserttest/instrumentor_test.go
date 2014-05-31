package powerasserttest

import (
	"twada/power/instrumentor"
	"testing"
	"go/ast"
	"go/parser"
	"go/token"
)

func streq(t *testing.T, actual, expected string, message string) {
	if actual != expected {
		t.Errorf("\nexpected: %v\n     got: %v\n%s", expected, actual, message)
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
	v := instrumentor.NewVisitor()
//	ast.Print(token.NewFileSet(), expr)
	ast.Walk(v, expr)
	result,_ := instrumentor.ToCode(expr)
	streq(t, result, "assert.PowerOk(t, assert.Capt(val, \"val\"), \"assert.Ok(t, val)\")", "output")
}


func TestBinaryExpression(t *testing.T) {
	expr := parse("assert.Ok(t, foo == bar)")
	v := instrumentor.NewVisitor()
	ast.Walk(v, expr)
	result,_ := instrumentor.ToCode(expr)
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

	v := instrumentor.NewVisitor()
	ast.Walk(v, f)

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
	result,_ := instrumentor.ToCode(f)
	streq(t, result, expected, "output")
}

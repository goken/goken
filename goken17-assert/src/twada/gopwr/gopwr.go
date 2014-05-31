package main

import (
	"io"
	"io/ioutil"
	"os"
	"go/ast"
	"go/parser"
	"go/token"
	"twada/power/instrumentor"
)

// How to generate powered test code:
// $ GOPATH=`pwd` go run src/twada/gopwr/gopwr.go < src/twada/exam/bt_test.go
func main() {
	in := os.Stdin
	out := os.Stdout

	src, err := ioutil.ReadAll(in)
	if err != nil {
		os.Exit(2)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		os.Exit(2)
	}

	v := instrumentor.NewVisitor()
	ast.Walk(v, f)
	res,_ := instrumentor.ToCode(f)

	io.WriteString(out, res)
}

package instrumentor

import (
	"fmt"
	"go/ast"
//	"go/parser"
	"go/token"
	"go/printer"
	"bytes"
)

const (
	DEBUG = false
)

func trace(a ...interface{}) (n int, err error) {
	if DEBUG {
		return fmt.Println(a...)
	}
	return
}

func log(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func ToCode(node interface{}) (code string, err error) {
	buffer := new(bytes.Buffer)
	err = printer.Fprint(buffer, token.NewFileSet(), node)
	code = buffer.String()
	return
}

func captIdent(ident *ast.Ident) (modified *ast.CallExpr) {
	modified = &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X: &ast.Ident{Name: "assert"},
			Sel: &ast.Ident{Name: "Capt"},
		},
		Args: []ast.Expr{
			ident,
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", ident.Name)},
		},
	}
	return
}

func captExpr(node ast.Expr) (modified *ast.CallExpr, ok bool) {
	switch n := node.(type) {
	case *ast.Ident:
		modified = captIdent(n)
		ok = true
	}
	return
}


func NewVisitor() *powerVisitor {
	v := &powerVisitor{}
	v.nodeStack = NewStack()
	return v
}


type powerVisitor struct {
	nodeStack Stack
	capturing bool
	capturingCall *ast.CallExpr
	original string
}


func (v *powerVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		v.leave()
		return nil
	}
	return v.enter(node)
}


func (v *powerVisitor) enter(node ast.Node) ast.Visitor {
	v.nodeStack.Push(node)
	trace("enter node. stack size: ", v.nodeStack.Count(), " pushed: ", node)

	if selector, ok := node.(*ast.SelectorExpr); ok {
		if pkg, ok := selector.X.(*ast.Ident); ok {
			if pkg.Name == "assert" && selector.Sel.Name == "Ok" {
				v.startCapturing()
				return nil
			}
		}
	}

	return v
}


func (v *powerVisitor) leave() {
	leavingNode := v.nodeStack.Pop()
	trace("leave node. stack size: ", v.nodeStack.Count(), " poped: ", leavingNode)

	if v.capturing == true {
		switch n := leavingNode.(type) {
		case *ast.BinaryExpr:
			v.leaveBinaryExpr(n)
		case *ast.CallExpr:
			if n == v.capturingCall {
				v.finishCapturing(n)
			}
		}
	}
}


func (v *powerVisitor) leaveBinaryExpr(binaryExpr *ast.BinaryExpr) {
	if modified, ok := captExpr(binaryExpr.X); ok {
		binaryExpr.X = modified
	}
	if modified, ok := captExpr(binaryExpr.Y); ok {
		binaryExpr.Y = modified
	}
}


func (v *powerVisitor) startCapturing() {
	v.capturing = true
	popped := v.nodeStack.Pop()
	trace("start capturing. stack size: ", v.nodeStack.Count(), " poped: ", popped)
	target := v.nodeStack.Peek()
	if call, ok := target.(*ast.CallExpr); ok {
		v.capturingCall = call
		trace("rewinding node. stack size ", v.nodeStack.Count(), " peeked: ", target)
		v.original, _ = ToCode(call)
	} else {
		panic(ok)
	}
}


func (v *powerVisitor) finishCapturing(n *ast.CallExpr) {
	trace("stop capturing. stack size: ", v.nodeStack.Count(), " poped: ", n)
	if selector, ok := n.Fun.(*ast.SelectorExpr); ok {
		if selector.Sel.Name == "Ok" {
			selector.Sel.Name = "PowerOk"
			arg := n.Args[1]
			if modified, ok := captExpr(arg); ok {
				n.Args[1] = modified
			}
			n.Args = append(n.Args, &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", v.original)})
		}
	}
//	modified, _ := ToCode(n)
//	log("-r='", v.original, "->", modified, "'")
	v.capturingCall = nil
	v.original = ""
	v.capturing = false
}

package codegen

import (
	"testing"

	"github.com/agatan/kaleigo/ast"
)

func TestGenFun(t *testing.T) {
	g := NewGenerator("test")
	value, err := g.GenFun(&ast.Function{
		Prototype: &ast.Prototype{
			Name: "f",
			Args: []string{"x", "y"},
		},
		Body: &ast.BinaryExpr{
			Op:  '+',
			LHS: &ast.VariableExpr{Name: "x"},
			RHS: &ast.VariableExpr{Name: "y"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if value.IsNil() {
		t.Fatalf("generated llvm.Value is nil")
	}
}

func TestGenExpr(t *testing.T) {
	g := NewGenerator("test")
	value, err := g.GenExpr(&ast.NumberExpr{Val: 0.0})
	if err != nil {
		t.Fatal(err)
	}
	if value.IsNil() {
		t.Fatalf("generated llvm.Value is nil")
	}
}

func TestGenExtern(t *testing.T) {
	g := NewGenerator("test")
	value, err := g.GenProto(&ast.Prototype{Name: "cos", Args: []string{"x"}})
	if err != nil {
		t.Fatal(err)
	}
	if value.IsNil() {
		t.Fatalf("generated llvm.Value is nil")
	}
	value, err = g.GenExpr(&ast.CallExpr{
		Callee: "cos",
		Args: []ast.Expr{
			&ast.NumberExpr{Val: 0.0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if value.IsNil() {
		t.Fatalf("generated llvm.Value is nil")
	}
}

func TestGenIfExpr(t *testing.T) {
	g := NewGenerator("test")
	value, err := g.GenFun(&ast.Function{
		Prototype: &ast.Prototype{
			Name: "testif",
			Args: []string{},
		},
		Body: &ast.IfExpr{
			Cond: &ast.NumberExpr{Val: 1.0},
			Then: &ast.NumberExpr{Val: 2.0},
			Else: &ast.NumberExpr{Val: 3.0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if value.IsNil() {
		t.Fatalf("generated llvm.Value from if expression is nil")
	}
}

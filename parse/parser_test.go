package parse

import (
	"reflect"
	"testing"

	"github.com/agatan/kaleigo/ast"
)

func TestParseExpression(t *testing.T) {
	p := New("test", "1 + 2 * 3")
	expr := p.ParseExpression()
	var expected ast.Expr = &ast.BinaryExpr{
		Op:  '+',
		LHS: &ast.NumberExpr{Val: 1.0},
		RHS: &ast.BinaryExpr{
			Op:  '*',
			LHS: &ast.NumberExpr{Val: 2.0},
			RHS: &ast.NumberExpr{Val: 3.0},
		},
	}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("binary expression precedence is wrong: expected %#v, actual %#v", expected, expr)
	}
}

func TestParseExtern(t *testing.T) {
	p := New("test", "extern pow(x, y)")
	actual := p.ParseExtern()
	expected := &ast.Prototype{
		Name: "pow",
		Args: []string{"x", "y"},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("external function declaration parsing is wrong")
	}
}

func TestParseDefinition(t *testing.T) {
	p := New("test", "def add(x, y) x + y")
	actual := p.ParseDefinition()
	expected := &ast.Function{
		Prototype: &ast.Prototype{
			Name: "add",
			Args: []string{"x", "y"},
		},
		Body: &ast.BinaryExpr{
			Op:  '+',
			LHS: &ast.VariableExpr{Name: "x"},
			RHS: &ast.VariableExpr{Name: "y"},
		},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("function definition parsing is wrong")
	}
}

func TestParseCall(t *testing.T) {
	p := New("test", "f()\n")
	actual := p.ParseExpression()
	expected := &ast.CallExpr{
		Callee: "f",
		Args:   []ast.Expr{},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("function definition parsing is wrong")
	}
}

func TestParseIf(t *testing.T) {
	p := New("test", "if 2 < 3 then 1 else 2")
	actual := p.ParseExpression()
	expected := &ast.IfExpr{
		Cond: &ast.BinaryExpr{
			Op:  '<',
			LHS: &ast.NumberExpr{Val: 2},
			RHS: &ast.NumberExpr{Val: 3},
		},
		Then: &ast.NumberExpr{Val: 1},
		Else: &ast.NumberExpr{Val: 2},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("if expression parsing is wrong")
	}
}

func TestParseFor(t *testing.T) {
	p := New("test", "for i = 1, i < n, 1.0 in i")
	actual := p.ParseExpression()
	expected := &ast.ForExpr{
		Var:   "i",
		Start: &ast.NumberExpr{Val: 1.0},
		End: &ast.BinaryExpr{
			Op:  '<',
			LHS: &ast.VariableExpr{Name: "i"},
			RHS: &ast.VariableExpr{Name: "n"},
		},
		Step: &ast.NumberExpr{Val: 1.0},
		Body: &ast.VariableExpr{Name: "i"},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("for expression parsing is wrong")
	}
}

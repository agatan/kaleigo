package parse

import (
	"reflect"
	"testing"

	"github.com/agatan/kaleigo/ast"
)

func TestParseExpression(t *testing.T) {
	p := New("test", "1 + 2 * 3")
	expr := p.parseExpression()
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

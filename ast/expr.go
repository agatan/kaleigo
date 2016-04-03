package ast

// Expr represents kaleigo's ast node.
type Expr interface {
	ExprKind() ExprType
}

// ExprType identifies ast node's type.
type ExprType int

const (
	ExprError ExprType = iota
	ExprNumber
	ExprUnary
	ExprBinary
	ExprCall
	ExprVariable
	ExprBlock
	ExprIf
	ExprFor
)

type (
	NumberExpr struct {
		Val float64
	}

	VariableExpr struct {
		Name string
	}

	BinaryExpr struct {
		Op  rune
		LHS Expr
		RHS Expr
	}

	CallExpr struct {
		Callee string
		Args   []Expr
	}

	BlockExpr struct {
		Exprs []Expr
	}

	IfExpr struct {
		Cond Expr
		Then Expr
		Else Expr
	}

	ForExpr struct {
		Var   string
		Start Expr
		End   Expr
		Step  Expr
		Body  Expr
	}
)

func (*NumberExpr) ExprKind() ExprType   { return ExprNumber }
func (*VariableExpr) ExprKind() ExprType { return ExprVariable }
func (*BinaryExpr) ExprKind() ExprType   { return ExprBinary }
func (*CallExpr) ExprKind() ExprType     { return ExprCall }
func (*BlockExpr) ExprKind() ExprType    { return ExprBlock }
func (*IfExpr) ExprKind() ExprType       { return ExprIf }
func (*ForExpr) ExprKind() ExprType      { return ExprFor }

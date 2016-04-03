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
)

func (*NumberExpr) ExprKind() ExprType   { return ExprNumber }
func (*VariableExpr) ExprKind() ExprType { return ExprVariable }
func (*BinaryExpr) ExprKind() ExprType   { return ExprBinary }
func (*CallExpr) ExprKind() ExprType     { return ExprCall }
func (*BlockExpr) ExprKind() ExprType    { return ExprBlock }

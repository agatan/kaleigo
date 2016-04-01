package ast

type Function struct {
	*Prototype
	Body Expr
}

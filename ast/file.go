package ast

type File struct {
	Name    string
	Externs []*Prototype
	Defs    []*Function
	Exprs   []Expr
}

// CreateMain creates dummy main function that contains all of toplevel expressions.
func (f *File) CreateMain() *Function {
	return &Function{
		Prototype: &Prototype{
			Name: "__kaleigo_main",
			Args: []string{},
		},
		Body: &BlockExpr{
			f.Exprs,
		},
	}
}

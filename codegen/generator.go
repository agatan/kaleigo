package codegen

import (
	"fmt"

	"github.com/agatan/kaleigo/ast"

	"llvm.org/llvm/bindings/go/llvm"
)

// Generator holds all information for llvm code generation.
type Generator struct {
	ctx     llvm.Context
	mod     llvm.Module
	builder llvm.Builder
	values  map[string]llvm.Value
}

// New creates a new llvm code generator
func New(name string) *Generator {
	return &Generator{
		ctx:     llvm.GlobalContext(),
		mod:     llvm.NewModule(name),
		builder: llvm.NewBuilder(),
		values:  make(map[string]llvm.Value),
	}
}

func (g *Generator) error(err error) error {
	return err
}

func (g *Generator) errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func (g *Generator) GenExpr(expr ast.Expr) (val llvm.Value, err error) {
	switch e := expr.(type) {
	case *ast.NumberExpr:
		return llvm.ConstFloat(llvm.DoubleType(), e.Val), nil
	case *ast.VariableExpr:
		v, ok := g.values[e.Name]
		if !ok {
			err = g.errorf("unknown variable name : %q", e.Name)
		}
		return v, err
	case *ast.BinaryExpr:
		l, err := g.GenExpr(e.LHS)
		if err != nil {
			return l, err
		}
		r, err := g.GenExpr(e.RHS)
		if err != nil {
			return r, err
		}

		switch e.Op {
		case '+':
			return g.builder.CreateFAdd(l, r, "addtmp"), nil
		case '-':
			return g.builder.CreateFSub(l, r, "subtmp"), nil
		case '*':
			return g.builder.CreateFMul(l, r, "multmp"), nil
		case '<':
			l = g.builder.CreateFCmp(llvm.FloatULT, l, r, "cmptmp")
			return g.builder.CreateUIToFP(l, llvm.DoubleType(), "booltmp"), nil

		default:
			err = fmt.Errorf("invalid binary operator: %q", e.Op)
			return val, err
		}
	case *ast.CallExpr:
		f := g.mod.NamedFunction(e.Callee)
		if f.IsNil() {
			return val, fmt.Errorf("unknown function referenced: %q", e.Callee)
		}

		if f.ParamsCount() != len(e.Args) {
			return val, fmt.Errorf("incorrect number of arguments passed for %q. %d expected, but %d given", e.Callee, f.ParamsCount(), len(e.Args))
		}

		args := []llvm.Value{}
		for _, arg := range e.Args {
			v, err := g.GenExpr(arg)
			if err != nil {
				return val, err
			}
			args = append(args, v)
		}

		return g.builder.CreateCall(f, args, "calltmp"), nil
	default:
		panic("internal compiler error")
	}
}

func (g *Generator) GenProto(p *ast.Prototype) (llvm.Value, error) {
	doubles := []llvm.Type{}
	for _ = range p.Args {
		doubles = append(doubles, llvm.DoubleType())
	}
	ft := llvm.FunctionType(llvm.DoubleType(), doubles, false)
	f := llvm.AddFunction(g.mod, p.Name, ft)
	if f.IsNil() {
		return f, fmt.Errorf("function is nil: %q", p.Name)
	}
	for i, arg := range f.Params() {
		arg.SetName(p.Args[i])
	}
	return f, nil
}

func (g *Generator) GenFun(f *ast.Function) (llvm.Value, error) {
	var err error
	ff := g.mod.NamedFunction(f.Name)
	if ff.IsNil() {
		ff, err = g.GenProto(f.Prototype)
		if err != nil {
			return ff, err
		}
	}

	bb := llvm.AddBasicBlock(ff, "entry")
	g.builder.SetInsertPointAtEnd(bb)
	g.values = make(map[string]llvm.Value)

	for _, arg := range ff.Params() {
		g.values[arg.Name()] = arg
	}

	body, err := g.GenExpr(f.Body)
	if err != nil {
		ff.EraseFromParentAsFunction()
		return body, err
	}

	g.builder.CreateRet(body)
	if llvm.VerifyFunction(ff, llvm.PrintMessageAction) != nil {
		ff.EraseFromParentAsFunction()
		return ff, fmt.Errorf("function verification failed: %q", f.Name)
	}

	return ff, nil
}
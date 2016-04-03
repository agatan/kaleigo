package codegen

import (
	"fmt"
	"io"

	"github.com/agatan/kaleigo/ast"

	"llvm.org/llvm/bindings/go/llvm"
)

func init() {
	llvm.InitializeNativeTarget()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeNativeAsmPrinter()
}

// Generator holds all information for llvm code generation.
type Generator struct {
	ctx     llvm.Context
	mod     llvm.Module
	builder llvm.Builder
	values  map[string]llvm.Value
}

// New creates a new llvm code generator
func NewGenerator(name string) *Generator {
	return &Generator{
		ctx:     llvm.GlobalContext(),
		mod:     llvm.NewModule(name),
		builder: llvm.NewBuilder(),
		values:  make(map[string]llvm.Value),
	}
}

func (g *Generator) Dispose() {
	g.mod.Dispose()
	g.builder.Dispose()
}

func (g *Generator) Emit(fileast *ast.File, out io.Writer) error {
	for _, extern := range fileast.Externs {
		_, err := g.GenProto(extern)
		if err != nil {
			return err
		}
	}
	for _, def := range fileast.Defs {
		_, err := g.GenFun(def)
		if err != nil {
			return err
		}
	}
	_, err := g.GenFun(fileast.CreateMain())
	if err != nil {
		return err
	}

	target, err := llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())
	if err != nil {
		return err
	}
	m := target.CreateTargetMachine(llvm.DefaultTargetTriple(), "", "",
		llvm.CodeGenLevelNone, llvm.RelocDefault, llvm.CodeModelDefault)

	buf, err := m.EmitToMemoryBuffer(g.mod, llvm.ObjectFile)
	if err != nil {
		return err
	}
	defer buf.Dispose()

	_, err = out.Write(buf.Bytes())
	return err
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

	case *ast.BlockExpr:
		var last llvm.Value
		var err error
		for _, e := range e.Exprs {
			last, err = g.GenExpr(e)
			if err != nil {
				return last, err
			}
		}
		return last, nil

	case *ast.IfExpr:
		cond, err := g.GenExpr(e.Cond)
		if err != nil {
			return cond, err
		}
		// cond == 0.0 ??
		cond = g.builder.CreateFCmp(llvm.FloatONE, cond, llvm.ConstFloat(llvm.DoubleType(), 0.0), "ifcond")

		// create basic blocks for if jump
		parent := g.builder.GetInsertBlock().Parent()
		thenbb := llvm.AddBasicBlock(parent, "then")
		elsebb := llvm.AddBasicBlock(parent, "else")
		mergebb := llvm.AddBasicBlock(parent, "ifcont")

		g.builder.CreateCondBr(cond, thenbb, elsebb)

		// then block
		g.builder.SetInsertPointAtEnd(thenbb)
		then, err := g.GenExpr(e.Then)
		if err != nil {
			return then, err
		}
		g.builder.CreateBr(mergebb)
		thenbb = g.builder.GetInsertBlock()

		// else block
		g.builder.SetInsertPointAtEnd(elsebb)
		else_, err := g.GenExpr(e.Else)
		if err != nil {
			return else_, err
		}
		g.builder.CreateBr(mergebb)
		elsebb = g.builder.GetInsertBlock()

		g.builder.SetInsertPointAtEnd(mergebb)
		phi := g.builder.CreatePHI(llvm.DoubleType(), "iftmp")
		phi.AddIncoming([]llvm.Value{then}, []llvm.BasicBlock{thenbb})
		phi.AddIncoming([]llvm.Value{else_}, []llvm.BasicBlock{elsebb})
		return phi, nil

	case *ast.ForExpr:
		start, err := g.GenExpr(e.Start)
		if err != nil {
			return start, err
		}
		parent := g.builder.GetInsertBlock().Parent()
		preheaderBB := g.builder.GetInsertBlock()
		loopBB := llvm.AddBasicBlock(parent, "loop")

		g.builder.CreateBr(loopBB)

		g.builder.SetInsertPointAtEnd(loopBB)
		phi := g.builder.CreatePHI(llvm.DoubleType(), e.Var)
		phi.AddIncoming([]llvm.Value{start}, []llvm.BasicBlock{preheaderBB})

		oldVal, oldExists := g.values[e.Var]
		g.values[e.Var] = phi

		g.GenExpr(e.Body)

		var step llvm.Value
		if e.Step != nil {
			step, err = g.GenExpr(e.Step)
			if err != nil {
				return step, err
			}
		} else {
			step = llvm.ConstFloat(llvm.DoubleType(), 1.0)
		}

		next := g.builder.CreateFAdd(phi, step, "nextvar")

		end, err := g.GenExpr(e.End)
		if err != nil {
			return end, err
		}

		end = g.builder.CreateFCmp(llvm.FloatONE, end, llvm.ConstFloat(llvm.DoubleType(), 0.0), "loopcond")

		loopEndBB := g.builder.GetInsertBlock()
		afterBB := llvm.AddBasicBlock(parent, "afterloop")

		g.builder.CreateCondBr(end, loopBB, afterBB)
		g.builder.SetInsertPointAtEnd(afterBB)

		phi.AddIncoming([]llvm.Value{next}, []llvm.BasicBlock{loopEndBB})

		if oldExists {
			g.values[e.Var] = oldVal
		} else {
			delete(g.values, e.Var)
		}

		return llvm.ConstFloat(llvm.DoubleType(), 0.0), nil

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

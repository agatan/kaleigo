package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/agatan/kaleigo/ast"
	"github.com/agatan/kaleigo/codegen"
	"github.com/agatan/kaleigo/parse"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	g := codegen.New("my_jit")
	for {
		fmt.Print("ready> ")
		if !scanner.Scan() {
			fmt.Println("Goodbye.")
			return
		}
		parser := parse.New("<stdin>", scanner.Text())
		switch parser.ToplevelKind() {
		case parse.ToplevelDef:
			f := parser.ParseDefinition()
			fmt.Println("definition")
			HandleDefinition(g, f)
		case parse.ToplevelExtern:
			e := parser.ParseExtern()
			fmt.Println("extern")
			HandleExtern(g, e)
		default:
			e := parser.ParseTopLevelExpr()
			fmt.Println("toplevel expression")
			HandleToplevelExpression(g, e)
		}
	}
}

func HandleDefinition(g *codegen.Generator, f *ast.Function) {
	llf, err := g.GenFun(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	llf.Dump()
}

func HandleExtern(g *codegen.Generator, p *ast.Prototype) {
	f, err := g.GenProto(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	f.Dump()
}

func HandleToplevelExpression(g *codegen.Generator, expr *ast.Function) {
	f, err := g.GenFun(expr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	f.Dump()
}

package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/agatan/kaleigo/ast"
	"github.com/agatan/kaleigo/codegen"
	"github.com/agatan/kaleigo/parse"
)

// Compiler holds compile options and status
type Compiler struct {
	cc string
}

// NewCompiler creates a new compiler with the options.(currently option is none.)
func NewCompiler() *Compiler {
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "CC"
	}
	return &Compiler{
		cc: cc,
	}
}

func (c *Compiler) CompileFile(filename string, outname string) error {
	f, err := c.parseFile(filename)
	if err != nil {
		return err
	}

	g := codegen.NewGenerator("kaleigo")
	defer g.Dispose()

	base := filepath.Base(outname)
	obj := base + ".o"

	objh, err := os.Create(obj)
	if err != nil {
		return err
	}
	defer func(name string) {
		if err := os.Remove(name); err != nil {
			panic(err)
		}
	}(obj)
	defer func(objh io.Closer) {
		if err := objh.Close(); err != nil {
			panic(err)
		}
	}(objh)

	if err := g.Emit(f, objh); err != nil {
		return err
	}

	cmd := exec.Command("gcc", obj, "lib/runtime.c", "-o", outname)
	return cmd.Run()
}

func (c *Compiler) parseFile(filename string) (*ast.File, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(fh io.Closer) {
		err := fh.Close()
		if err != nil {
			panic(err)
		}
	}(fh)

	input, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, err
	}

	p := parse.New(filename, string(input))
	return p.Parse(), nil
}

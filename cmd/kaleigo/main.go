package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/agatan/kaleigo/ast"
	"github.com/agatan/kaleigo/codegen"
	"github.com/agatan/kaleigo/parse"
)

func compile(fileinfo *ast.File, outname string) error {
	base := filepath.Base(outname)
	bc := base + ".bc"

	bch, err := os.Create(bc)
	if err != nil {
		return err
	}
	defer func(bch io.Closer) {
		if err := bch.Close(); err != nil {
			panic(err)
		}
	}(bch)

	err = codegen.EmitBitCode(fileinfo, bch)
	if err != nil {
		return err
	}

	obj := base + ".o"
	cmd := exec.Command("llc-3.8", bc, "-filetype=obj", "-o", obj)
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("gcc", obj, "lib/runtime.c", "-o", outname)
	return cmd.Run()
}

func main() {
	file := "<stdin>"
	src := os.Stdin
	if len(os.Args) > 1 {
		var err error
		src, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		file = os.Args[1]
		defer func(src io.ReadCloser) {
			if err := src.Close(); err != nil {
				panic(err)
			}
		}(src)
	}

	input, err := ioutil.ReadAll(src)
	if err != nil {
		panic(err)
	}

	parser := parse.New(file, string(input))

	f := parser.Parse()
	fmt.Println("Parse: Done")

	err = compile(f, "a.out")
	if err != nil {
		panic(err)
	}
}

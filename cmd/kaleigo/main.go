package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "no file name given.")
		return
	}

	c := NewCompiler()
	err := c.CompileFile(os.Args[1], "a.out")
	if err != nil {
		log.Fatalln(err)
	}
}

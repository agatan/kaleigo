package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/agatan/kaleigo/parse"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("ready> ")
		if !scanner.Scan() {
			fmt.Println("Goodbye.")
			return
		}
		parser := parse.New("<stdin>", scanner.Text())
		switch parser.ToplevelKind() {
		case parse.ToplevelDef:
			_ = parser.ParseDefinition()
			fmt.Println("definition")
		case parse.ToplevelExtern:
			_ = parser.ParseExtern()
			fmt.Println("extern")
		default:
			_ = parser.ParseTopLevelExpr()
			fmt.Println("toplevel expression")
		}
	}
}

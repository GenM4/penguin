package main

import (
	"fmt"
	"os"

	"github.com/GenM4/penguin/pkg/parser"
	"github.com/GenM4/penguin/pkg/tokenizer"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	if os.Args[1] == "" {
		return
	}

	file := os.Args[1]
	dat, err := os.ReadFile(file)
	Check(err)

	fmt.Println("INPUT:")
	fmt.Println(string(dat))
	fmt.Println("")

	tokens := tokenizer.Tokenize(dat)

	fmt.Println("TOKENS:")
	for _, token := range tokens {
		if token.Data == "\n" {
			fmt.Print("\\n\\", "\n")
		} else {
			fmt.Print(token.Data, "\t")
		}
	}

	ASTRoot := parser.Parse(tokens)
	fmt.Println("Program: " + ASTRoot.Data)

	for _, stmt := range ASTRoot.Children {
		fmt.Println("Statement: " + stmt.Data)
		for _, expr := range stmt.Children {
			fmt.Println("Expression: ")
			fmt.Print("\t" + expr.Data + "\n")
			for _, term := range expr.Children {
				fmt.Print(term.Data + "\t\t")
			}
			fmt.Print("\n")
		}
	}

	return
}

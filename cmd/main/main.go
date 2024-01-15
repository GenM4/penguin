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
	//fmt.Println(tokens)

	for _, line := range dat {
		parser.Parse(string(line))
	}

	return
}

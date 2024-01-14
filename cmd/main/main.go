package main

import (
	"fmt"
	"github.com/GenM4/penguin/parser"
	"os"
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

	for _, line := range dat {
		parser.Parse(string(line))
		fmt.Println(line)
	}

	return
}

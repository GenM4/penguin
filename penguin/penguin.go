package main

import (
	"fmt"
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
		fmt.Println(line)
	}

	return
}

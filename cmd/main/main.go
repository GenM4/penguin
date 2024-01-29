package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/GenM4/penguin/pkg/generator"
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

	srcFilepath := os.Args[len(os.Args)-1]
	srcFilename := filepath.Base(srcFilepath)
	baseFilepath := filepath.Dir(srcFilepath)
	baseFilename := removeFileExtension(srcFilename)
	asmFilename := baseFilename + ".asm"
	asmFilepath := filepath.Join(baseFilepath, asmFilename)
	objFilename := baseFilename + ".o"
	objFilepath := filepath.Join(baseFilepath, objFilename)
	execFilepath := filepath.Join(baseFilepath, baseFilename)

	asmFile, err := os.OpenFile(asmFilepath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	log.Println("Opened file: " + asmFilepath)

	if err = os.Truncate(asmFilepath, 0); err != nil {
		panic(err)
	}

	generator.Generate(ASTRoot, asmFile)

	log.Println("Completed generating assembly to " + asmFilepath)

	assembleCmd := exec.Command("nasm", "-felf64", asmFilename)
	assembleCmd.Dir = baseFilepath
	if err = assembleCmd.Run(); err != nil {
		log.Println("Assembly failed")
		panic(err)
	}

	log.Println("Completed assembly to " + objFilepath)

	linkCmd := exec.Command("ld", objFilename, "-o", baseFilename)
	linkCmd.Dir = baseFilepath
	if err = linkCmd.Run(); err != nil {
		log.Print(err)
		panic(err)
	}

	log.Println("Completed linking to " + execFilepath)

	return
}

func removeFileExtension(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

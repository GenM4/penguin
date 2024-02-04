package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/GenM4/penguin/pkg/generator"
	"github.com/GenM4/penguin/pkg/parser"
	"github.com/GenM4/penguin/pkg/tokenizer"

	"github.com/m1gwings/treedrawer/tree"
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
	for _, token := range tokens.Tokens {
		if token.Data == "\n" {
			fmt.Print("\\n\\", "\n")
		} else {
			fmt.Print(token.Data, "\t")
		}
	}

	ASTRoot := parser.Parse(&tokens)
	ASTRoot.Data = file

	t := tree.NewTree(tree.NodeString("Prog: " + ASTRoot.Data))
	for _, stmt := range ASTRoot.Children {
		stmtNode := t.AddChild(tree.NodeString("Stmt: " + stmt.Data))
		for _, expr := range stmt.Children {
			exprNode := stmtNode.AddChild(tree.NodeString(expr.Kind.String() + ": " + expr.Data + "\n" + "Prec: " + strconv.Itoa(expr.Precedence)))
			for _, term := range expr.Children {
				var termNode *tree.Tree
				if term.Precedence != -1 {
					termNode = exprNode.AddChild(tree.NodeString(term.Kind.String() + ": " + term.Data + "\n" + "Prec: " + strconv.Itoa(term.Precedence)))
				} else {
					termNode = exprNode.AddChild(tree.NodeString(term.Data))
				}
				for _, term2 := range term.Children {
					var term2Node *tree.Tree
					if term2.Precedence != -1 {
						term2Node = termNode.AddChild(tree.NodeString(term2.Kind.String() + ": " + term2.Data + "\n" + "Prec: " + strconv.Itoa(term2.Precedence)))
					} else {
						term2Node = termNode.AddChild(tree.NodeString(term2.Data))
					}
					for _, term3 := range term2.Children {
						var term3Node *tree.Tree
						if term3.Precedence != -1 {
							term3Node = term2Node.AddChild(tree.NodeString(term3.Kind.String() + ": " + term3.Data + "\n" + "Prec: " + strconv.Itoa(term3.Precedence)))
						} else {
							term3Node = term2Node.AddChild(tree.NodeString(term3.Data))
						}
						for _, term4 := range term3.Children {
							if term4.Precedence != -1 {
								term3Node.AddChild(tree.NodeString(term4.Kind.String() + ": " + term4.Data + "\n" + "Prec: " + strconv.Itoa(term4.Precedence)))
							} else {
								term3Node.AddChild(tree.NodeString(term4.Data))
							}
						}
					}
				}
			}
		}
	}
	fmt.Println(t)

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

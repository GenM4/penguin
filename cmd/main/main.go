package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/GenM4/penguin/pkg/generator"
	"github.com/GenM4/penguin/pkg/parser"
	"github.com/GenM4/penguin/pkg/semantics"
	"github.com/GenM4/penguin/pkg/tokenizer"
	"github.com/GenM4/penguin/pkg/utils/files"

	"github.com/m1gwings/treedrawer/tree"
)

var PRINT_TOKEN_KINDS bool = false

func main() {

	dat := ReadSourceFile(os.Args[1])
	tokens := TokenizeFile(dat)

	vars := make(semantics.VarMap)

	ASTRoot := ParseTokens(tokens, &vars)

	printVarMap(&vars)

	fileData := files.GenerateFilepaths(os.Args)
	asmFile := files.OpenTargetFile(fileData.AsmFilepath)

	GenerateAssembly(ASTRoot, &vars, asmFile, fileData.AsmFilepath)

	Assemble(fileData)
	Link(fileData)

	return
}

func ReadSourceFile(filepath string) []byte {
	if filepath == "" {
		panic(fmt.Errorf("No source filepath provided"))
	}

	dat, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	fmt.Println("INPUT:")
	fmt.Println(string(dat))

	return dat
}

func TokenizeFile(srcData []byte) tokenizer.TokenStack {
	tokens := tokenizer.Tokenize(srcData)
	printTokens(tokens, PRINT_TOKEN_KINDS)

	return tokens
}

func ParseTokens(tokens tokenizer.TokenStack, vars *semantics.VarMap) *parser.ASTNode {
	ASTRoot := parser.Parse(&tokens, vars)
	ASTRoot.Data = os.Args[1]
	printAST(ASTRoot)

	return ASTRoot
}

func GenerateAssembly(root *parser.ASTNode, vars *semantics.VarMap, file *os.File, filepath string) {
	generator.Generate(root, vars, file)
	log.Println("Completed generating assembly to " + filepath)

}

func Assemble(fileData files.FileData) {
	assembleCmd := exec.Command("nasm", "-felf64", fileData.AsmFilename)
	assembleCmd.Dir = fileData.BaseFilepath
	if err := assembleCmd.Run(); err != nil {
		log.Println("Assembly Failed")
		panic(err)
	}

	log.Println("Completed assembly to " + fileData.ObjFilepath)
}

func Link(fileData files.FileData) {
	linkCmd := exec.Command("ld", fileData.ObjFilename, "-o", fileData.BaseFilename)
	linkCmd.Dir = fileData.BaseFilepath
	if err := linkCmd.Run(); err != nil {
		log.Print("Linking Failed")
		panic(err)
	}

	log.Println("Completed linking to " + fileData.ExecFilepath)
}

func printTokens(tokens tokenizer.TokenStack, printKinds bool) {
	fmt.Println("TOKENS:")
	for _, token := range tokens.Tokens {
		if token.Data == "\n" {
			fmt.Print("\\n\\", "\n")
		} else {
			var kind string
			if printKinds {
				kind = token.Kind.String() + ":"
			} else {
				kind = ""
			}
			fmt.Print(kind+token.Data, "\t")
		}
	}
}

func printAST(ASTRoot *parser.ASTNode) {
	t := tree.NewTree(tree.NodeString("Prog: " + ASTRoot.Data))
	for _, stmt := range ASTRoot.Children {
		stmtNode := t.AddChild(tree.NodeString("Stmt: " + stmt.Data))
		for _, expr := range stmt.Children {
			var exprNode *tree.Tree
			if expr.Kind == parser.Declaration {
				exprNode = stmtNode.AddChild(tree.NodeString(expr.Kind.String() + ": " + expr.Data + "\n" + "Mutable: " + strconv.FormatBool(expr.Mutable) + "\n" + "Type: " + expr.Type.String()))
			} else {
				exprNode = stmtNode.AddChild(tree.NodeString(expr.Kind.String() + ": " + expr.Data + "\n" + "Prec: " + strconv.Itoa(expr.Precedence)))
			}
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
}

func printVarMap(vars *semantics.VarMap) {
	fmt.Println("VARIABLE MAP: ")
	for k, v := range *vars {
		fmt.Print(k + ": ")
		fmt.Print(*v)
		fmt.Print("\t")
	}
	fmt.Print("\n")
}

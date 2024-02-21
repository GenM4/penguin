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

	vars, funcs := InitMaps()

	ASTRoot := ParseTokens(tokens, &vars, &funcs)

	printVarMap(&vars)

	fileData := files.GenerateFilepaths(os.Args)
	asmFile := files.OpenTargetFile(fileData.AsmFilepath)

	GenerateAssembly(ASTRoot, &vars, &funcs, asmFile, fileData.AsmFilepath)

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

func InitMaps() (semantics.VarMap, semantics.FuncMap) {
	vars := make(semantics.VarMap)

	funcs := make(semantics.FuncMap)
	funcs["exit"] = &semantics.Function{Mutable: false, Type: semantics.Int, NumArgs: 1}
	funcs["print"] = &semantics.Function{Mutable: false, Type: semantics.Char, NumArgs: 1}

	return vars, funcs
}

func ParseTokens(tokens tokenizer.TokenStack, vars *semantics.VarMap, funcs *semantics.FuncMap) *parser.ASTNode {
	ASTRoot := parser.Parse(&tokens, vars, funcs)
	ASTRoot.Data = os.Args[1]
	printAST(ASTRoot)

	return ASTRoot
}

func GenerateAssembly(root *parser.ASTNode, vars *semantics.VarMap, funcs *semantics.FuncMap, file *os.File, filepath string) {
	generator.Generate(root, vars, funcs, file)
	log.Println("Completed generating assembly to " + filepath)

}

func Assemble(fileData files.FileData) {
	assembleCmd := exec.Command("nasm", "-felf64", fileData.AsmFilename)
	assembleCmd.Dir = fileData.BaseFilepath
	if err := assembleCmd.Run(); err != nil {
		log.Println("Assembly Failed")
		panic(err)
	}

	log.Println("Completed assembling to " + fileData.ObjFilepath)
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
	t := tree.NewTree(tree.NodeString(formatASTNode(*ASTRoot)))
	for _, l1Node := range ASTRoot.Children {
		l1TreeNode := t.AddChild(tree.NodeString(formatASTNode(l1Node)))
		for _, l2Node := range l1Node.Children {
			l2TreeNode := l1TreeNode.AddChild(tree.NodeString(formatASTNode(l2Node)))
			for _, l3Node := range l2Node.Children {
				l3TreeNode := l2TreeNode.AddChild(tree.NodeString(formatASTNode(l3Node)))
				for _, l4Node := range l3Node.Children {
					l4TreeNode := l3TreeNode.AddChild(tree.NodeString(formatASTNode(l4Node)))
					for _, l5Node := range l4Node.Children {
						l4TreeNode.AddChild(tree.NodeString(formatASTNode(l5Node)))
					}
				}
			}
		}
	}
	fmt.Println(t)
}

func formatASTNode(node parser.ASTNode) string {
	switch node.Kind {
	case parser.Program:
		return node.Kind.String() + ": " + node.Data
	case parser.Declaration:
		return node.Kind.String() + ": " + node.Data + "\n" + "Mutable: " + strconv.FormatBool(node.Mutable) + "\n" + "Type: " + node.Type.String()
	case parser.Scope:
		return node.Kind.String() + ": " + "'" + node.Parent.Data + "'"
	case parser.Statement:
		return node.Kind.String() + ": " + node.Data
	case parser.Expression:
		return node.Kind.String() + ": " + node.Data + "\n" + "Prec: " + strconv.Itoa(node.Precedence)
	case parser.Identifier:
		return node.Kind.String() + ": " + node.Data + "\n" + "Type: " + node.Type.String() + "\n" + "Prec: " + strconv.Itoa(node.Precedence)
	case parser.Term:
		return node.Kind.String() + ": " + node.Data + "\n" + "Type: " + node.Type.String() + "\n" + "Prec: " + strconv.Itoa(node.Precedence)
	default:
		return node.Kind.String() + ": " + node.Data + "\n" + "Type: " + node.Type.String() + "\n"
	}
}

func printVarMap(vars *semantics.VarMap) {
	fmt.Println("VARIABLE MAP: ")
	for k, v := range *vars {
		variable := *v
		fmt.Print(k + ": ")
		fmt.Printf("%v ", variable.Mutable)
		fmt.Printf("%v ", variable.Type.String())
		fmt.Printf("%v ", variable.StackLocation)
		fmt.Print("\t")
	}
	fmt.Print("\n\n")
}

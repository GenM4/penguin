package generator

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/GenM4/penguin/pkg/parser"
)

func Generate(root *parser.ASTNode, asmFile *os.File) {

	_, err := asmFile.WriteString("global _start\n")
	_, err = asmFile.WriteString("_start:\n")
	if err != nil {
		panic(err)
	}

	err = traverseAST(*root, asmFile)
	if err != nil {
		asmFile.Close()
		panic(err)
	}

	genDefaultExit(asmFile)

	return
}

func traverseAST(node parser.ASTNode, asmFile *os.File) error {
	for _, child := range node.Children {
		if child.Kind == parser.Statement {
			log.Println("Generating assembly for statement: " + child.Data)
			genStatement(child, asmFile)
		} else {
			return errors.New("AST Node " + node.Kind.String() + "not implemented")
		}
	}
	return nil
}

func genStatement(node parser.ASTNode, asmFile *os.File) error {
	if node.Data == "exit" {
		genAtom(node.Children[0], asmFile)

		asmFile.WriteString("\tmov rax, 60\n")
		asmFile.WriteString("\tpop rdi\n")
		asmFile.WriteString("\tsyscall\n")
	}
	return nil
}

func genAtom(node parser.ASTNode, asmFile *os.File) error {
	if node.Kind == parser.Expression && len(node.Children) == 2 {
		genExpression(node, asmFile)
	} else if node.Kind == parser.Term {
		genTerm(node, asmFile)
	}

	return nil
}

func genExpression(node parser.ASTNode, asmFile *os.File) error {
	var parsedLHS, parsedRHS bool = false, false
	if node.Kind == parser.Expression && len(node.Children) == 2 {
		if node.Precedence <= node.Children[0].Precedence {
			fmt.Println("LHS: " + node.Children[0].Data)
			err := genExpression(node.Children[0], asmFile)
			if err != nil {
				return err
			}
			parsedLHS = true
		}
		if node.Precedence <= node.Children[1].Precedence {
			fmt.Println("RHS: " + node.Children[1].Data)
			err := genExpression(node.Children[1], asmFile)
			if err != nil {
				return err
			}
			parsedRHS = true
		}

		genBinaryExpression(node, asmFile)

		if !parsedLHS {
			err := genExpression(node.Children[0], asmFile)
			if err != nil {
				return err
			}
			parsedLHS = true
		}
		if !parsedRHS {
			err := genExpression(node.Children[1], asmFile)
			if err != nil {
				return err
			}
			parsedRHS = true
		}
	}

	return nil
}

func genBinaryExpression(node parser.ASTNode, asmFile *os.File) error {
	switch {
	case node.Data == "+":
		err := prepBinaryExpressionCall(node, asmFile)
		_, err = asmFile.WriteString("\tadd rax, rbx" + "\n")
		_, err = asmFile.WriteString("\tpush rax" + "\n")

		if err != nil {
			return err
		}
	case node.Data == "-":
		err := prepBinaryExpressionCall(node, asmFile)
		_, err = asmFile.WriteString("\tsub rax, rbx" + "\n")
		_, err = asmFile.WriteString("\tpush rax" + "\n")

		if err != nil {
			return err
		}
	case node.Data == "*":
		err := prepBinaryExpressionCall(node, asmFile)
		_, err = asmFile.WriteString("\tmul rbx" + "\n")
		_, err = asmFile.WriteString("\tpush rax" + "\n")

		if err != nil {
			return err
		}
	case node.Data == "/":
		err := prepBinaryExpressionCall(node, asmFile)
		_, err = asmFile.WriteString("\tdiv rbx" + "\n")
		_, err = asmFile.WriteString("\tpush rax" + "\n")

		if err != nil {
			return err
		}
	default:
		return errors.New("Expression " + node.Data + "not implemented")
	}

	return nil
}

func genTerm(node parser.ASTNode, asmFile *os.File) error {
	_, err := asmFile.WriteString("\tmov rax, " + node.Data + "\n")
	_, err = asmFile.WriteString("\tpush rax" + "\n")
	if err != nil {
		return err
	}
	return nil
}

func genDefaultExit(asmFile *os.File) {
	asmFile.WriteString("\tmov rax, 60\n")
	asmFile.WriteString("\tmov rdi, 0\n")
	asmFile.WriteString("\tsyscall\n")

}

func prepBinaryExpressionCall(node parser.ASTNode, asmFile *os.File) error {
	var err error
	if node.Children[0].IsOperator() {
		_, err = asmFile.WriteString("\tpop rax, " + "\n")
	} else {
		_, err = asmFile.WriteString("\tmov rax, " + node.Children[0].Data + "\n")
	}

	if node.Children[1].IsOperator() {
		_, err = asmFile.WriteString("\tpop rbx, " + "\n")
	} else {
		_, err = asmFile.WriteString("\tmov rbx, " + node.Children[1].Data + "\n")
	}

	return err
}

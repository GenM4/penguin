package generator

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/GenM4/penguin/pkg/parser"
	"github.com/GenM4/penguin/pkg/semantics"
)

type GeneratorData struct {
	asmFile          *os.File
	stackPtrLocation int
	vars             *semantics.VarMap
}

func Generate(root *parser.ASTNode, vars *semantics.VarMap, out *os.File) {
	genData := GeneratorData{
		asmFile:          out,
		stackPtrLocation: 1,
		vars:             vars,
	}

	_, err := genData.asmFile.WriteString("global _start\n")
	_, err = genData.asmFile.WriteString("_start:\n")
	if err != nil {
		panic(err)
	}

	err = traverseAST(*root, &genData)
	if err != nil {
		genData.asmFile.Close()
		panic(err)
	}

	genDefaultExit(genData.asmFile)

	return
}

func traverseAST(node parser.ASTNode, genData *GeneratorData) error {
	for _, child := range node.Children {
		if child.Kind == parser.Statement {
			log.Println("Generating assembly for statement: " + child.Data)
			err := genStatement(child, genData)
			if err != nil {
				return err
			}
		} else {
			return errors.New("AST Node " + node.Kind.String() + "not implemented")
		}
	}
	return nil
}

func genStatement(node parser.ASTNode, genData *GeneratorData) error {
	if node.Data == "=" {
		if node.Children[0].Kind == parser.Declaration {
			if variable, ok := (*genData.vars)[node.Children[0].Data]; ok {
				node.Children[1].Type = node.Children[0].Type

				genAtom(node.Children[1], genData)

				push("rax", variable.Type, genData)
				variable.StackLocation = genData.stackPtrLocation

			} else {
				return errors.New("Variable: " + node.Children[0].Data + " already declared")
			}
		} else if node.Children[0].Kind == parser.Identifier && node.Children[0].Mutable == true {
			node.Children[1].Type = node.Children[0].Type
			genAtom(node.Children[1], genData)

			variable := (*genData.vars)[node.Children[0].Data]
			word, _ := bytesToWord(variable.Type.Size())
			genData.asmFile.WriteString("\tmov " + word + " [rsp + " + strconv.Itoa(8*(genData.stackPtrLocation-variable.StackLocation)) + "]" + ", rax\n")

		} else {
			return errors.New("Attempt to modify a const value: " + node.Children[0].Data)
		}

	} else if node.Data == "exit" {
		genAtom(node.Children[0], genData)

		genData.asmFile.WriteString("\tmov rax, 60\n")
		genData.asmFile.WriteString("\tpop rdi\n")
		genData.asmFile.WriteString("\tsyscall\n")
	}
	return nil
}

func genAtom(node parser.ASTNode, genData *GeneratorData) error {
	var err error
	if node.Kind == parser.Expression && len(node.Children) == 2 {
		err = genExpression(node, genData)
	} else if node.Kind == parser.Term {
		err = genTerm(node, genData)
	} else if node.Kind == parser.Identifier {
		err = genIdentifier(node, genData)
	}

	if err != nil {
		return err
	}

	return nil
}

func genExpression(node parser.ASTNode, genData *GeneratorData) error {
	var generatedLHS, generatedRHS bool = false, false
	if node.Kind == parser.Expression && len(node.Children) == 2 {
		if node.Precedence <= node.Children[0].Precedence {
			err := genExpression(node.Children[0], genData)
			if err != nil {
				return err
			}
			generatedLHS = true
		}
		if node.Precedence <= node.Children[1].Precedence {
			err := genExpression(node.Children[1], genData)
			if err != nil {
				return err
			}
			generatedRHS = true
		}

		genBinaryExpression(node, genData)

		if !generatedLHS {
			err := genExpression(node.Children[0], genData)
			if err != nil {
				return err
			}
			generatedLHS = true
		}
		if !generatedRHS {
			err := genExpression(node.Children[1], genData)
			if err != nil {
				return err
			}
			generatedRHS = true
		}
	}

	return nil
}

func genBinaryExpression(node parser.ASTNode, genData *GeneratorData) error {
	switch {
	case node.Data == "+":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tadd rax, rbx" + "\n")
		err = push("rax", node.Type, genData)

		if err != nil {
			return err
		}
	case node.Data == "-":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tsub rax, rbx" + "\n")
		err = push("rax", node.Type, genData)

		if err != nil {
			return err
		}
	case node.Data == "*":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tmul rbx" + "\n")
		err = push("rax", node.Type, genData)

		if err != nil {
			return err
		}
	case node.Data == "/":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tdiv rbx" + "\n")
		err = push("rax", node.Type, genData)

		if err != nil {
			return err
		}
	default:
		return errors.New("Expression " + node.Data + "not implemented")
	}

	return nil
}

func genTerm(node parser.ASTNode, genData *GeneratorData) error {
	_, err := genData.asmFile.WriteString("\tmov rax, " + node.Data + "\n")
	if err != nil {
		return err
	}
	return nil
}

func genIdentifier(node parser.ASTNode, genData *GeneratorData) error {
	if variable, ok := (*genData.vars)[node.Data]; ok {
		word, _ := bytesToWord(variable.Type.Size())
		err := push(word+" [rsp + "+strconv.Itoa(8*(genData.stackPtrLocation-variable.StackLocation))+"]", variable.Type, genData)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("Variable: " + node.Data + "not declared")
	}

}

func genDefaultExit(asmFile *os.File) {
	asmFile.WriteString("\tmov rax, 60\n")
	asmFile.WriteString("\tmov rdi, 0\n")
	asmFile.WriteString("\tsyscall\n")

}

func prepBinaryExpressionCall(node parser.ASTNode, genData *GeneratorData) error {
	var err error
	if node.Children[0].IsOperator() {
		err = pop("rax", node.Type, genData)
	} else if node.Children[0].Kind == parser.Term {
		_, err = genData.asmFile.WriteString("\tmov rax, " + node.Children[0].Data + "\n")
	} else if node.Children[0].Kind == parser.Identifier {
		err = genIdentifier(node.Children[0], genData)
		err = pop("rax", node.Children[0].Type, genData)
	}

	if node.Children[1].IsOperator() {
		err = pop("rbx", node.Type, genData)
	} else if node.Children[1].Kind == parser.Term {
		_, err = genData.asmFile.WriteString("\tmov rbx, " + node.Children[1].Data + "\n")
	} else if node.Children[1].Kind == parser.Identifier {
		err = genIdentifier(node.Children[1], genData)
		err = pop("rbx", node.Children[1].Type, genData)
	}

	return err
}

func push(register string, dataType semantics.Type, genData *GeneratorData) error {
	_, err := genData.asmFile.WriteString("\tpush " + register + "\t" + ";;" + strconv.Itoa(genData.stackPtrLocation) + "\n")
	if err != nil {
		return err
	}

	genData.stackPtrLocation += 1

	return nil
}

func pop(register string, dataType semantics.Type, genData *GeneratorData) error {
	_, err := genData.asmFile.WriteString("\tpop " + register + "\n")
	if err != nil {
		return err
	}

	genData.stackPtrLocation -= 1

	return nil
}

func bytesToWord(bytes int) (string, int) {
	switch {
	case bytes <= 2:
		return "WORD", 0
	default:
		return "QWORD", 0
	}
}

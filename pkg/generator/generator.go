package generator

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/GenM4/penguin/pkg/parser"
	"github.com/GenM4/penguin/pkg/semantics"
)

type OpCode int

func (oc OpCode) String() string {
	return strconv.Itoa(int(oc))
}

type IntLiteral int

func (i IntLiteral) String() string {
	return strconv.Itoa(int(i)) + "h"
}

type Register int

const (
	RAX Register = iota + 0
	RBX
	RCX
	RDX
	EDI
	RSP
	RSI
	RDI
)

func (reg Register) String() string {
	name := []string{
		"rax",
		"rbx",
		"rcx",
		"rdx",
		"edi",
		"rsp",
		"rsi",
		"rdi",
	}

	i := int(reg)
	switch {
	case i <= int(RDI):
		return name[i]
	default:
		return strconv.Itoa(i)
	}
}

type StackAddress struct {
	Register Register
	Offset   int
	Size     string
}

func (sa StackAddress) String() string {
	return sa.Size + " [" + sa.Register.String() + " + " + strconv.Itoa(8*sa.Offset) + "]"
}

type movable interface {
	IntLiteral | OpCode | Register | StackAddress
	String() string
}

type pushable interface {
	Register | StackAddress
	String() string
}

type popable interface {
	Register | StackAddress
	String() string
}

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

	err = genProgram(*root, &genData)
	if err != nil {
		genData.asmFile.Close()
		panic(err)
	}

	err = genDefaultExit(genData.asmFile, &genData)
	if err != nil {
		panic(err)
	}

	return
}

func genProgram(node parser.ASTNode, genData *GeneratorData) error {
	for _, child := range node.Children {
		if child.Kind == parser.Statement {
			log.Println("Generating assembly for statement: " + child.Data)
			err := genStatement(child, genData)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("AST Node '%v'", node.Kind.String())
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
				push(RAX, node.Children[0].Type, genData)
				variable.StackLocation = genData.stackPtrLocation
			} else {
				return fmt.Errorf("Variable: '%v' already declared", node.Children[0].Data)
			}
		} else if node.Children[0].Kind == parser.Identifier && node.Children[0].Mutable == true {
			node.Children[1].Type = node.Children[0].Type
			genAtom(node.Children[1], genData)
			pop(RAX, node.Children[0].Type, genData)
			reassign(node.Children[0], genData)
		} else {
			return fmt.Errorf("Attempt to modify a const value: '%v'", node.Children[0].Data)
		}

	} else if node.Data == "++" || node.Data == "--" {
		expr := node.Children[0]

		if expr.Children[0].Mutable == true {
			genAtom(expr, genData)
			pop(RAX, node.Children[0].Type, genData)
			reassign(expr.Children[0], genData)
		} else {
			return fmt.Errorf("Attempt to modify const value: '%v'", node.Children[0].Data)
		}
	} else if node.Data == "print" {
		genAtom(node.Children[0], genData)
		move(RAX, OpCode(1), genData)
		move(RDI, OpCode(1), genData)
		move(RSI, RSP, genData)
		move(RDX, OpCode(1), genData)
		genData.asmFile.WriteString("\tsyscall\n")
	} else if node.Data == "exit" {
		genAtom(node.Children[0], genData)

		move(RAX, OpCode(60), genData)
		pop(RDI, node.Children[0].Type, genData)
		genData.asmFile.WriteString("\tsyscall\n")
	}
	return nil
}

func genAtom(node parser.ASTNode, genData *GeneratorData) error {
	var err error
	if node.Kind == parser.Expression && len(node.Children) == 2 {
		err = genExpression(node, genData)
	} else if node.Kind == parser.Term {
		err = genTerm(RAX, node, genData)
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
		err = push(RAX, node.Type, genData)

		if err != nil {
			return err
		}
	case node.Data == "-":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tsub rax, rbx" + "\n")
		err = push(RAX, node.Type, genData)

		if err != nil {
			return err
		}
	case node.Data == "*":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tmul rbx" + "\n")
		err = push(RAX, node.Type, genData)

		if err != nil {
			return err
		}
	case node.Data == "/":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tdiv rbx" + "\n")
		err = push(RAX, node.Type, genData)

		if err != nil {
			return err
		}
	default:
		return errors.New("Expression " + node.Data + "not implemented")
	}

	return nil
}

func genTerm(register Register, node parser.ASTNode, genData *GeneratorData) error {
	data, err := strconv.Atoi(node.Data)
	if err != nil {
		return err
	}

	err = move(register, IntLiteral(data), genData)
	if err != nil {
		return err
	}

	return nil
}

func genIdentifier(node parser.ASTNode, genData *GeneratorData) error {
	if variable, ok := (*genData.vars)[node.Data]; ok {
		offset := genData.stackPtrLocation - variable.StackLocation
		var addr = StackAddress{
			Register: RSP,
			Offset:   offset,
			Size:     bytesToWord(variable.Type.Size()),
		}

		err := push(addr, variable.Type, genData)
		if err != nil {
			return err
		}
	}

	return fmt.Errorf("Variable: '%v' not declared", node.Data)
}

func genDefaultExit(asmFile *os.File, genData *GeneratorData) error {
	err := move(RAX, OpCode(60), genData)
	err = move(RDI, OpCode(0), genData)
	if err != nil {
		return err
	}

	asmFile.WriteString("\tsyscall\n")

	return nil
}

func prepBinaryExpressionCall(node parser.ASTNode, genData *GeneratorData) error {
	var err error
	if node.Children[0].IsOperator() {
		err = pop(RAX, node.Type, genData)
	} else if node.Children[0].Kind == parser.Term {
		err = genTerm(RAX, node.Children[0], genData)
	} else if node.Children[0].Kind == parser.Identifier {
		err = genIdentifier(node.Children[0], genData)
		err = pop(RAX, node.Children[0].Type, genData)
	}

	if node.Children[1].IsOperator() {
		err = pop(RBX, node.Type, genData)
	} else if node.Children[1].Kind == parser.Term {
		err = genTerm(RBX, node.Children[1], genData)
	} else if node.Children[1].Kind == parser.Identifier {
		err = genIdentifier(node.Children[1], genData)
		err = pop(RBX, node.Children[1].Type, genData)
	}

	return err
}

func move[T1 movable, T2 movable](to T1, from T2, genData *GeneratorData) error {
	_, err := genData.asmFile.WriteString("\tmov " + to.String() + ", " + from.String() + "\n")

	if err != nil {
		return err
	}

	return nil
}

func push[T pushable](val T, dataType semantics.Type, genData *GeneratorData) error {
	_, err := genData.asmFile.WriteString("\tpush " + val.String() + "\t\t\t" + ";; Stack position: " + strconv.Itoa(genData.stackPtrLocation) + "\n")

	if err != nil {
		return err
	}

	genData.stackPtrLocation += 1

	return nil
}

func pop(register Register, dataType semantics.Type, genData *GeneratorData) error {
	_, err := genData.asmFile.WriteString("\tpop " + register.String() + "\n")
	if err != nil {
		return err
	}

	genData.stackPtrLocation -= 1

	return nil
}

func reassign(ident parser.ASTNode, genData *GeneratorData) error {
	variable := (*genData.vars)[ident.Data]
	offset := genData.stackPtrLocation - variable.StackLocation
	var addr = StackAddress{
		Register: RSP,
		Offset:   offset,
		Size:     bytesToWord(variable.Type.Size()),
	}

	move(addr, RAX, genData)

	return nil
}

func bytesToWord(bytes int) string {
	switch {
	case bytes <= 2:
		return "WORD"
	default:
		return "QWORD"
	}
}

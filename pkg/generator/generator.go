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

type CharLiteral string

func (c CharLiteral) String() string {
	if c == "\\n" {
		return OpCode(10).String() // ascii value for CR
	}

	return "'" + string(c) + "'"
}

type Register int

const (
	RAX Register = iota + 0
	RBX
	RCX
	ECX
	RDX
	EDX
	RDI
	EDI
	RSI
	ESI
	RSP
	RBP
)

func (reg Register) String() string {
	name := []string{
		"rax",
		"rbx",
		"rcx",
		"ecx",
		"rdx",
		"edx",
		"rdi",
		"edi",
		"rsi",
		"esi",
		"rsp",
		"rbp",
	}

	i := int(reg)
	switch {
	case i <= int(RBP):
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
	IntLiteral | CharLiteral | OpCode | Register | StackAddress
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
	argRegisters     []Register
	stackPtrLocation int
	vars             *semantics.VarMap
	funcs            *semantics.FuncMap
}

func Generate(root *parser.ASTNode, vars *semantics.VarMap, funcs *semantics.FuncMap, out *os.File) {
	genData := GeneratorData{
		asmFile:          out,
		argRegisters:     []Register{RDI, RSI, RDX, RCX},
		stackPtrLocation: 1,
		vars:             vars,
		funcs:            funcs,
	}

	_, err := genData.asmFile.WriteString("global _start\n")

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
		if child.Kind == parser.Statement || child.Kind == parser.Call {
			log.Println("Generating assembly for statement: " + child.Data + " in global scope")
			err := genStatement(child, genData)
			if err != nil {
				return err
			}
		} else if child.Kind == parser.Declaration {
			if len(child.Children) > 0 {
				log.Println("Generating assembly for declaration: " + child.Data + "() in global scope")
				err := genDeclaration(child, genData)
				if err != nil {
					return err
				}
			} else {
				continue
			}
		} else {
			return fmt.Errorf("Unexpected %v in %v", child.Kind.String(), node.Kind.String())
		}
	}
	return nil
}

func genDeclaration(node parser.ASTNode, genData *GeneratorData) error {
	if node.Data == "main" {
		genData.asmFile.WriteString("_start:\n")
	} else {
		genData.asmFile.WriteString((*genData.funcs)[node.Data].Signature + ":" + "\n")
	}

	localStackLocation := genData.stackPtrLocation
	genData.stackPtrLocation = 1

	err := push(RBP, genData)
	if err != nil {
		return err
	}

	err = move(RBP, RSP, genData)
	if err != nil {
		return err
	}

	err = genArguments(node.Children[:len(node.Children)-1], genData)
	if err != nil {
		return err
	}

	err = genScope(node.Children[len(node.Children)-1], genData)
	if err != nil {
		return err
	}

	err = pop(RBP, genData)
	if err != nil {
		return err
	}

	genData.asmFile.WriteString("\tret\n")
	genData.stackPtrLocation = localStackLocation

	return nil
}

func genArguments(args []parser.ASTNode, genData *GeneratorData) error {
	if len(args) > len(genData.argRegisters) {
		return fmt.Errorf("Exceeded maximum number of arguments in function call")
	}

	for i, arg := range args {
		err := genArg(genData.argRegisters[i], arg, genData)
		if err != nil {
			return err
		}
	}

	return nil
}

func genScope(node parser.ASTNode, genData *GeneratorData) error {
	for _, child := range node.Children {
		if child.Kind == parser.Statement || child.Kind == parser.Call {
			log.Println("Generating assembly for statement: " + child.Data + " in " + node.Kind.String() + ": " + "'" + node.Parent.Data + "'")
			err := genStatement(child, genData)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Unexpected %v in %v: '%v'", child.Kind.String(), node.Kind.String(), node.Data)
		}
	}

	return nil

}

func genStatement(node parser.ASTNode, genData *GeneratorData) error {
	if node.Data == "=" {
		if node.Children[0].Kind == parser.Declaration {
			if variable, ok := (*genData.vars)[node.Children[0].Data]; ok {
				if node.Children[1].Type == semantics.Untyped {
					node.Children[1].Type = node.Children[0].Type
				}

				err := genAtom(RAX, node.Children[1], genData)
				if err != nil {
					return err
				}

				variable.StackLocation = genData.stackPtrLocation

				err = reassign(node.Children[0], genData)
				if err != nil {
					return err
				}

				genData.stackPtrLocation += 1
			} else {
				return fmt.Errorf("Variable: '%v' already declared", node.Children[0].Data)
			}
		} else if node.Children[0].Kind == parser.Identifier && node.Children[0].Mutable == true {
			if node.Children[1].Type == semantics.Untyped {
				node.Children[1].Type = node.Children[0].Type
			}

			err := genAtom(RAX, node.Children[1], genData)
			if err != nil {
				return err
			}

			err = reassign(node.Children[0], genData)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Attempt to modify a const value: '%v'", node.Children[0].Data)
		}

	} else if node.Data == "++" || node.Data == "--" {
		expr := node.Children[0]

		if expr.Children[0].Mutable == true {
			err := genAtom(RAX, expr, genData)
			if err != nil {
				return err
			}

			err = reassign(expr.Children[0], genData)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Attempt to modify const value: '%v'", expr.Children[0].Data)
		}
	} else if function, ok := (*genData.funcs)[node.Data]; ok {
		err := genCall(RAX, function, node, genData)
		if err != nil {
			return err
		}
	} else if node.Data == "return" {
		expr := node.Children[0]

		err := genAtom(RAX, expr, genData)
		if err != nil {
			return err
		}
	}
	return nil
}

func genAtom(to Register, node parser.ASTNode, genData *GeneratorData) error {
	var err error
	if node.Kind == parser.Expression && len(node.Children) == 2 {
		err = genExpression(node, genData)
		pop(to, genData)
	} else if node.Kind == parser.Term {
		err = genTerm(to, node, genData)
	} else if _, ok := (*genData.vars)[node.Data]; ok {
		err = genIdentifier(to, node, genData)
	} else if function, ok := (*genData.funcs)[node.Data]; ok {
		err = genCall(to, function, node, genData)
	}

	if err != nil {
		return err
	}

	return nil
}

func genCall(to Register, function *semantics.Function, node parser.ASTNode, genData *GeneratorData) error {
	if node.Data == "print" {
		if node.Children[0].Type != semantics.Char {
			return fmt.Errorf("print only implemented for char, attempted call with type %v", node.Type.String())
		}

		genAtom(RAX, node.Children[0], genData)
		push(RAX, genData)

		move(RAX, OpCode(1), genData)
		move(RDI, OpCode(1), genData)
		move(RSI, RSP, genData)
		move(RDX, OpCode(1), genData)
		genData.asmFile.WriteString("\tsyscall\n")
		pop(RAX, genData)
	} else if node.Data == "exit" {
		genAtom(RDI, node.Children[0], genData)
		move(RAX, OpCode(60), genData)
		genData.asmFile.WriteString("\tsyscall\n")
	} else {
		for i := 0; i < function.NumArgs; i++ {
			err := genAtom(RAX, node.Children[i], genData)
			if err != nil {
				return err
			}

			err = move(genData.argRegisters[i], RAX, genData)
			if err != nil {
				return err
			}
		}

		genData.asmFile.WriteString("\tcall " + function.Signature + "\n")
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
		err = push(RAX, genData)

		if err != nil {
			return err
		}
	case node.Data == "-":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tsub rax, rbx" + "\n")
		err = push(RAX, genData)

		if err != nil {
			return err
		}
	case node.Data == "*":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tmul rbx" + "\n")
		err = push(RAX, genData)

		if err != nil {
			return err
		}
	case node.Data == "/":
		err := prepBinaryExpressionCall(node, genData)
		_, err = genData.asmFile.WriteString("\tdiv rbx" + "\n")
		err = push(RAX, genData)

		if err != nil {
			return err
		}
	default:
		return errors.New("Expression " + node.Data + "not implemented")
	}

	return nil
}

func genTerm(register Register, node parser.ASTNode, genData *GeneratorData) error {
	var err error

	switch node.Type {
	case semantics.Int:
		err = genIntLiteral(register, node, genData)
		break
	case semantics.Char:
		err = genCharLiteral(register, node, genData)
		break
	}

	return err
}

func genIntLiteral(register Register, node parser.ASTNode, genData *GeneratorData) error {
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

func genCharLiteral(register Register, node parser.ASTNode, genData *GeneratorData) error {
	return move(register, CharLiteral(node.Data[1:len(node.Data)-1]), genData)
}

func genIdentifier(to Register, node parser.ASTNode, genData *GeneratorData) error {
	if variable, ok := (*genData.vars)[node.Data]; ok {
		offset := variable.StackLocation
		var addr = StackAddress{
			Offset:   offset,
			Register: RBP,
			Size:     bytesToWord(variable.Type.Size()),
		}

		return move(to, addr, genData)
	}

	return fmt.Errorf("Variable: '%v' not declared", node.Data)
}

func genArg(from Register, node parser.ASTNode, genData *GeneratorData) error {
	if variable, ok := (*genData.vars)[node.Data]; ok {
		variable.StackLocation = genData.stackPtrLocation
		offset := variable.StackLocation
		var addr = StackAddress{
			Offset:   offset,
			Register: RBP,
			Size:     bytesToWord(variable.Type.Size()),
		}

		err := move(addr, from, genData)
		genData.stackPtrLocation += 1

		return err
	}

	return fmt.Errorf("Variable: '%v' already declared in outer scope", node.Data)
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
		err = pop(RAX, genData)
	} else if node.Children[0].Kind == parser.Term {
		err = genTerm(RAX, node.Children[0], genData)
	} else if node.Children[0].Kind == parser.Identifier {
		err = genIdentifier(RAX, node.Children[0], genData)
	}

	if node.Children[1].IsOperator() {
		err = pop(RBX, genData)
	} else if node.Children[1].Kind == parser.Term {
		err = genTerm(RBX, node.Children[1], genData)
	} else if node.Children[1].Kind == parser.Identifier {
		err = genIdentifier(RBX, node.Children[1], genData)
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

func push[T pushable](val T, genData *GeneratorData) error {
	_, err := genData.asmFile.WriteString("\tpush " + val.String() + "\t\t\t" + ";; Local Stack position: " + strconv.Itoa(genData.stackPtrLocation) + "\n")
	if err != nil {
		return err
	}
	genData.stackPtrLocation += 1

	return nil
}

func pop(register Register, genData *GeneratorData) error {
	genData.stackPtrLocation -= 1

	_, err := genData.asmFile.WriteString("\tpop " + register.String() + "\n")
	if err != nil {
		return err
	}

	return nil
}

func reassign(ident parser.ASTNode, genData *GeneratorData) error {
	variable := (*genData.vars)[ident.Data]
	offset := variable.StackLocation
	var addr = StackAddress{
		Register: RBP,
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

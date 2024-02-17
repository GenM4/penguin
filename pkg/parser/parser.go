package parser

import (
	"fmt"
	"strconv"

	"github.com/GenM4/penguin/pkg/semantics"
	"github.com/GenM4/penguin/pkg/tokenizer"
)

type ASTNodeType int

const (
	Program ASTNodeType = iota + 0
	Statement
	Declaration
	Expression
	Identifier
	Term
)

type ASTNode struct {
	Kind       ASTNodeType
	Data       string
	Precedence int
	Type       semantics.Type
	Mutable    bool
	Parent     *ASTNode
	Children   []ASTNode
}

func (nodeType ASTNodeType) String() string {
	name := []string{
		"Program",
		"Statement",
		"Declaration",
		"Expression",
		"Identifier",
		"Term",
	}

	i := int(nodeType)
	switch {
	case i <= int(Term):
		return name[i]
	default:
		return strconv.Itoa(i)
	}

}

func (node ASTNode) IsOperator() bool {
	if node.Kind == Expression && (node.Data == "+" || node.Data == "-" || node.Data == "*" || node.Data == "/") {
		return true
	}
	return false
}

func Parse(tokens *tokenizer.TokenStack, vars *semantics.VarMap) *ASTNode {
	root, err := parseProgram(tokens, vars)
	if err != nil {
		panic(err)
	}

	return root
}

func parseProgram(tokens *tokenizer.TokenStack, vars *semantics.VarMap) (*ASTNode, error) {
	var prog = ASTNode{
		Kind: Program,
	}

	for tokens.Len() > 1 {
		if tokens.Top().Kind == tokenizer.CR {
			tokens.Next()
		} else {

			stmt, err := parseStatement(tokens, vars)
			if err != nil {
				return nil, err
			}

			prog.Children = append(prog.Children, stmt)
		}
	}

	return &prog, nil
}

func parseStatement(tokens *tokenizer.TokenStack, vars *semantics.VarMap) (ASTNode, error) {
	stmt := ASTNode{
		Kind: Statement,
	}

	if tokens.Top().Kind == tokenizer.Mutable && tokens.Peek(1).Kind == tokenizer.Type {
		if tokens.Peek(3).Kind == tokenizer.SingleEqual {
			stmt, err := parseAssignment(true, false, tokens, vars)

			return *stmt, err
		}
	} else if tokens.Top().Kind == tokenizer.Type {
		if tokens.Peek(2).Kind == tokenizer.SingleEqual {
			stmt, err := parseAssignment(false, false, tokens, vars)

			return *stmt, err
		}

	} else if tokens.Top().Kind == tokenizer.Identifier {
		if tokens.Peek(1).Kind == tokenizer.SingleEqual {
			stmt, err := parseAssignment(true, true, tokens, vars)

			return *stmt, err

		} else if tokens.Peek(1).Kind == tokenizer.Operator_plusplus || tokens.Peek(1).Kind == tokenizer.Operator_minusminus {
			stmt.Data = tokens.Peek(1).Data

			expr, err := parseIncrement(tokens, vars)
			if err != nil {
				return ASTNode{}, err
			}

			stmt.Children = append(stmt.Children, expr)

			return stmt, nil
		} else {
			return ASTNode{}, fmt.Errorf("Unrecognized operator after identifier '%v'", tokens.Top().Data)
		}

	} else if tokens.Top().Kind == tokenizer.Exit || tokens.Top().Kind == tokenizer.Print {
		stmt.Data = tokens.Top().Data

		tokens.Next()
		if tokens.Top().Kind == tokenizer.Open_paren {
			tokens.Next()
		} else {
			return ASTNode{}, fmt.Errorf("Expected '(' before '%v'", tokens.Top().Data)
		}

		expr, err := parseExpression(tokens, 0, vars)
		if err != nil {
			return ASTNode{}, err
		}

		if tokens.Top().Kind == tokenizer.Close_paren {
			tokens.Next()
		} else if tokens.Top().Kind == tokenizer.CR {
		} else {
			return ASTNode{}, fmt.Errorf("Mismatched parens, expected ')' before '%v'", tokens.Top().Data)
		}

		stmt.Children = append(stmt.Children, *expr)

		return stmt, nil
	}

	return ASTNode{}, fmt.Errorf("Unrecognized token: '%v' before: '%v'", tokens.Top().Data, tokens.Peek(1).Data)
}

func parseAssignment(hasMutable bool, isDeclared bool, tokens *tokenizer.TokenStack, vars *semantics.VarMap) (*ASTNode, error) {
	assignment := &ASTNode{
		Data: "=",
		Kind: Statement,
	}

	var err error
	var lhs *ASTNode
	if isDeclared {
		lhs, err = parseTerm(tokens, vars)
		if lhs.Mutable != true {
			return &ASTNode{}, fmt.Errorf("Attempt to write to immutable value '%v'", lhs.Data)
		}
	} else {
		lhs, err = parseDeclaration(tokens, vars, hasMutable)
	}

	tokens.Next()
	tokens.Next()

	expr, err := parseExpression(tokens, 0, vars)
	if err != nil {
		return &ASTNode{}, err
	}

	if lhs.Type != expr.Type {
		return &ASTNode{}, fmt.Errorf("Attempted to assign expression (type: %v) to '%v' (type: %v)", expr.Type.String(), lhs.Data, lhs.Type.String())
	}

	tokens.Next()

	assignment.Children = append(assignment.Children, *lhs)
	assignment.Children = append(assignment.Children, *expr)

	return assignment, nil
}

func parseIncrement(tokens *tokenizer.TokenStack, vars *semantics.VarMap) (ASTNode, error) {
	lhs, err := parseTerm(tokens, vars)
	if err != nil {
		return ASTNode{}, err
	}

	if lhs.Type != semantics.Int {
		return ASTNode{}, fmt.Errorf("Increment/Decrement operator not implemented for type %v", lhs.Type.String())
	}

	rhs := &ASTNode{
		Kind: Term,
		Data: "1",
	}

	tokens.Next()

	expr := &ASTNode{
		Kind:       Expression,
		Precedence: 2,
		Data:       string(tokens.Top().Data[1]),
	}

	expr.Children = append(expr.Children, *lhs)
	expr.Children = append(expr.Children, *rhs)

	tokens.Next()

	return *expr, nil
}

func parseDeclaration(tokens *tokenizer.TokenStack, vars *semantics.VarMap, hasMutable bool) (*ASTNode, error) {
	decl := &ASTNode{
		Kind: Declaration,
	}

	if hasMutable && tokens.Top().Data == "mut" {
		tokens.Next()
		decl.Mutable = true
	} else if hasMutable && tokens.Top().Data == "const" {
		tokens.Next()
		decl.Mutable = false
	} else if hasMutable {
		return &ASTNode{}, fmt.Errorf("Unrecognized keyword: '%v' before '%v'", tokens.Top().Data, tokens.Next().Data)
	} else {
		decl.Mutable = false
	}

	var err error
	decl.Type, err = semantics.MatchType(tokens.Top().Data)
	if err != nil {
		return &ASTNode{}, err
	}

	tokens.Next()

	decl.Data = tokens.Top().Data
	(*vars)[decl.Data] = &semantics.Variable{Mutable: decl.Mutable, Type: decl.Type, StackLocation: 0}

	return decl, nil
}

func parseExpression(tokens *tokenizer.TokenStack, minPrec int, vars *semantics.VarMap) (*ASTNode, error) {
	lhs, err := parseTerm(tokens, vars)
	if err != nil {
		return &ASTNode{}, err
	}
	lookahead := tokens.Peek(1)

	var expr *ASTNode
	var nextMinPrec int

	for tokenizer.IsOperator(lookahead) && tokenizer.OperatorPrecedence(lookahead) >= minPrec {
		tokens.Next()
		nextMinPrec = tokenizer.OperatorPrecedence(tokens.Top())

		expr = new(ASTNode)
		expr.Kind = Expression
		expr.Data = tokens.Top().Data
		expr.Precedence = tokenizer.OperatorPrecedence(tokens.Top()) + 1
		expr.Children = []ASTNode{{}, {}}

		tokens.Next()

		rhs, err := parseTerm(tokens, vars)
		if err != nil {
			return &ASTNode{}, err
		}

		lookahead = tokens.Peek(1)
		for tokenizer.IsOperator(lookahead) && tokenizer.OperatorPrecedence(lookahead) > nextMinPrec {
			rhs, err = parseExpression(tokens, nextMinPrec, vars)
			if err != nil {
				return &ASTNode{}, err
			}

			if tokens.Top().Kind != tokenizer.Close_paren {
				tokens.Next()
			}
			lookahead = tokens.Peek(1)

		}
		expr.Children[0] = *lhs
		expr.Children[1] = *rhs
		expr.Type = lhs.Type

		lookahead = tokens.Peek(1)

		lhs = expr
	}
	if tokens.Peek(1).Kind == tokenizer.Close_paren {
		tokens.Next()
	}
	return lhs, nil
}

func parseTerm(tokens *tokenizer.TokenStack, vars *semantics.VarMap) (*ASTNode, error) {
	if tokens.Top().Kind == tokenizer.Int_literal {
		return &ASTNode{
			Kind: Term,
			Data: tokens.Top().Data,
			Type: semantics.Int,
		}, nil
	} else if tokens.Top().Kind == tokenizer.Char_literal {
		return &ASTNode{
			Kind: Term,
			Data: tokens.Top().Data,
			Type: semantics.Char,
		}, nil
	} else if tokens.Top().Kind == tokenizer.Identifier {
		variable, ok := (*vars)[tokens.Top().Data]
		if !ok {
			return &ASTNode{}, fmt.Errorf("Variable: '%v' not declared", tokens.Top().Data)
		}

		return &ASTNode{
			Kind:    Identifier,
			Data:    tokens.Top().Data,
			Type:    variable.Type,
			Mutable: variable.Mutable,
		}, nil
	}

	return &ASTNode{}, fmt.Errorf("Unrecognized term: '%v'", tokens.Top())
}

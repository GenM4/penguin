package parser

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/GenM4/penguin/pkg/tokenizer"
)

type ASTNodeType int

const (
	Program ASTNodeType = iota + 0
	Statement
	Expression
	Term
)

type ASTNode struct {
	Kind       ASTNodeType
	Data       string
	Precedence int
	Parent     *ASTNode
	Children   []ASTNode
}

func (nodeType ASTNodeType) String() string {
	name := []string{
		"Program",
		"Statement",
		"Expression",
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

func Parse(tokens *tokenizer.TokenStack) *ASTNode {
	root, err := parseProgram(tokens)
	if err != nil {
		panic(err)
	}

	return root
}

func parseProgram(tokens *tokenizer.TokenStack) (*ASTNode, error) {
	var node_Prog = ASTNode{
		Kind: Program,
	}

	for tokens.Len() > 1 {
		if tokens.Top().Kind == tokenizer.CR {
			tokens.Next()
		}

		stmt, err := parseStatement(tokens)
		if err != nil {
			return nil, err
		}

		node_Prog.Children = append(node_Prog.Children, stmt)

	}

	return &node_Prog, nil
}

func parseStatement(tokens *tokenizer.TokenStack) (ASTNode, error) {
	if tokens.Top().Kind == tokenizer.Exit {
		tokExitData := tokens.Top().Data

		tokens.Next()
		if tokens.Top().Kind == tokenizer.Open_paren {
			tokens.Next()
		} else {
			return ASTNode{}, fmt.Errorf("Expected '(' before %v", tokens.Top().Data)
		}

		expr, err := parseExpression(tokens, 0)
		if err != nil {
			return ASTNode{}, err
		}

		if tokens.Top().Kind == tokenizer.Close_paren {
			tokens.Next()
		} else if tokens.Top().Kind == tokenizer.CR {
		} else {
			return ASTNode{}, fmt.Errorf("Mismatched parens, expected ')' before %v", tokens.Top().Data)
		}

		var stmt = ASTNode{
			Kind: Statement,
			Data: tokExitData,
		}

		stmt.Children = append(stmt.Children, *expr)

		return stmt, nil
	}

	return ASTNode{}, errors.New("Unrecognized token: " + tokens.Top().Data)
}

func parseExpression(tokens *tokenizer.TokenStack, minPrec int) (*ASTNode, error) {
	lhs, err := parseTerm(tokens)
	if err != nil {
		return nil, err
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

		rhs, err := parseTerm(tokens)
		if err != nil {
			return nil, err
		}

		lookahead = tokens.Peek(1)
		for tokenizer.IsOperator(lookahead) && tokenizer.OperatorPrecedence(lookahead) > nextMinPrec {
			rhs, err = parseExpression(tokens, nextMinPrec)
			if err != nil {
				return nil, err
			}

			if tokens.Top().Kind != tokenizer.Close_paren {
				tokens.Next()
			}
			lookahead = tokens.Peek(1)

		}
		expr.Children[0] = *lhs
		expr.Children[1] = *rhs

		lookahead = tokens.Peek(1)

		lhs = expr
	}
	if tokens.Peek(1).Kind == tokenizer.Close_paren {
		tokens.Next()
	}
	return lhs, nil
}

func parseTerm(tokens *tokenizer.TokenStack) (*ASTNode, error) {
	if tokens.Top().Kind == tokenizer.Int_literal {
		return &ASTNode{
			Kind: Term,
			Data: tokens.Top().Data,
		}, nil
	}

	return nil, fmt.Errorf("Unrecognized term: %v", tokens.Top())
}

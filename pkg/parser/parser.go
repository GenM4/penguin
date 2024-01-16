package parser

import (
	"errors"

	"github.com/GenM4/penguin/pkg/tokenizer"
)

type ASTNodeType int

const (
	program ASTNodeType = iota + 0
	statement
	expression
	operator
	term
)

type ASTNode struct {
	Kind     ASTNodeType
	Data     string
	Parent   *ASTNode
	Children []ASTNode
}

func Parse(tokens []tokenizer.Token) *ASTNode {
	root, err := parseProgram(tokens)
	if err != nil {
		panic(err)
	}

	return root
}

func parseProgram(tokens []tokenizer.Token) (*ASTNode, error) {
	var node_Prog = ASTNode{
		Kind: program,
	}

	for i := 0; i < len(tokens); i++ {
		if tokens[i].Kind == tokenizer.CR {
			continue
		}
		stmt, numParsed, err := parseStatement(tokens[i:])

		if err != nil {
			return nil, err
		}

		node_Prog.Children = append(node_Prog.Children, stmt)

		i += numParsed
	}

	return &node_Prog, nil
}

func parseStatement(tokens []tokenizer.Token) (ASTNode, int, error) {
	curr := tokens[0]
	if curr.Kind == tokenizer.Exit {

		expr, numParsed, err := parseExpression(tokens[1:])

		if err != nil {
			return ASTNode{}, 0, err
		}

		var stmt = ASTNode{
			Kind: statement,
			Data: curr.Data,
		}

		stmt.Children = append(stmt.Children, expr)

		return stmt, numParsed + 1, nil
	}

	return ASTNode{}, 0, errors.New("Unrecognized token: " + curr.Data)
}

func parseExpression(tokens []tokenizer.Token) (ASTNode, int, error) {
	start := 0
	if tokens[0].Kind == tokenizer.Open_paren {
		start = 1
	}

	if tokens[start].Kind == tokenizer.Int_literal {
		term1 := ASTNode{
			Kind: term,
			Data: tokens[start].Data,
		}

		tok := view(tokens[start+1:], 0)
		if tok.Kind == tokenizer.Operator_plus || tok.Kind == tokenizer.Operator_minus || tok.Kind == tokenizer.Operator_star || tok.Kind == tokenizer.Operator_slash {
			op := ASTNode{
				Kind: operator,
				Data: tok.Data,
			}

			op.Children = append(op.Children, term1)

			view(tokens[start+2:], 0)
			term2, numParsed, err := parseExpression(tokens[start+2:])
			if err != nil {
				return ASTNode{}, start + 2 + numParsed, err
			}

			op.Children = append(op.Children, term2)

			return op, start + 2 + numParsed, nil
		} else if tok.Kind == tokenizer.Close_paren {
			return term1, start + 1, nil
		}
	}

	return ASTNode{}, start, errors.New("Unidentified expression: " + tokens[start].Data)
}

func view(tokens []tokenizer.Token, pos int) tokenizer.Token {
	if pos > len(tokens) {
		panic(errors.New("Attempted view outside of available tokens"))
	}

	return tokens[pos]
}

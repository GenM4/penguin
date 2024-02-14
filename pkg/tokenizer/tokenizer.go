package tokenizer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	Exit TokenType = iota + 0
	Print
	Open_paren
	Close_paren
	CR
	Operator_plus
	Operator_minus
	Operator_star
	Operator_slash
	Operator_plusplus
	Operator_minusminus
	Int_literal
	Char_literal
	Mutable
	Type
	SingleEqual
	Identifier
)

func (tokenType TokenType) String() string {
	name := []string{
		"Exit",
		"Print",
		"Open_paren",
		"Close_paren",
		"CR",
		"Plus",
		"Minus",
		"Star",
		"Slash",
		"MinusMinus",
		"PlusPlus",
		"Int_Literal",
		"Char_Literal",
		"Mutable",
		"Type",
		"Equal",
		"Identifier",
	}

	i := int(tokenType)
	switch {
	case i <= int(Identifier):
		return name[i]
	default:
		return strconv.Itoa(i)
	}

}

var TokenDict = map[string]TokenType{
	"exit":  Exit,
	"print": Print,
	"(":     Open_paren,
	")":     Close_paren,
	"\n":    CR,
	"+":     Operator_plus,
	"-":     Operator_minus,
	"*":     Operator_star,
	"/":     Operator_slash,
	"++":    Operator_plusplus,
	"--":    Operator_minusminus,
	"mut":   Mutable,
	"const": Mutable,
	"int":   Type,
	"char":  Type,
	"=":     SingleEqual,
}

type Token struct {
	Data string
	Kind TokenType
}

type TokenStack struct {
	Tokens []Token
	index  int
}

func (toks TokenStack) Append(buf string) TokenStack {
	if buf == "" {
		return toks
	}

	kind, err := matchToken(buf)
	if err != nil {
		panic(err)
	}

	tok := Token{
		Data: buf,
		Kind: kind,
	}
	toks.Tokens = append(toks.Tokens, tok)

	return toks
}

func (toks TokenStack) Top() Token {
	return toks.Tokens[toks.index]
}

func (toks *TokenStack) Next() Token {
	toks.index++
	if toks.index > len(toks.Tokens) {
		panic(fmt.Errorf("Attempt to access outside bounds of TokenStack at index %v", toks.index))
	}
	return toks.Top()
}

func (toks TokenStack) Peek(offset int) Token {
	return toks.Tokens[toks.index+offset]
}

func (toks TokenStack) Len() int {
	return len(toks.Tokens[toks.index:])
}

func OperatorPrecedence(tok Token) int {
	switch {
	case tok.Kind == Operator_plus || tok.Kind == Operator_minus:
		return 1
	case tok.Kind == Operator_star || tok.Kind == Operator_slash:
		return 2
	default:
		return -1
	}
}

func IsOperator(tok Token) bool {
	if tok.Kind == Operator_plus ||
		tok.Kind == Operator_minus ||
		tok.Kind == Operator_star ||
		tok.Kind == Operator_slash ||
		tok.Kind == Operator_plusplus ||
		tok.Kind == Operator_minusminus {
		return true
	}

	return false
}

func Tokenize(raw []byte) TokenStack {
	fileContents := string(raw)

	var result TokenStack
	result.index = 0

	last := 0
	for i := 0; i < len(fileContents); i++ {
		buf := fileContents[last:i]

		curr := view(fileContents, i)
		if curr == '/' && view(fileContents, i+1) == '/' {
			i += strings.Index(fileContents[i:], "\n")
			last = i + 1

			if i == 0 { // EOF
				break
			}
		} else if curr == '+' && view(fileContents, i+1) == '+' {
			result = result.Append(buf)
			result = result.Append("++")
			last = i + 2
			i = i + 1
		} else if curr == '-' && view(fileContents, i+1) == '-' {
			result = result.Append(buf)
			result = result.Append("--")
			last = i + 2
			i = i + 1
		} else if curr == '\'' {
			result = result.Append(buf)
			i++
			curr = view(fileContents, i)
			for curr != '\'' {
				i++
				curr = view(fileContents, i)
			}
		} else if curr == '(' {
			result = result.Append(buf)
			result = result.Append("(")
			last = i + 1
		} else if curr == ')' {
			result = result.Append(buf)
			result = result.Append(")")
			last = i + 1
		} else if curr == '\n' {
			result = result.Append(buf)
			result = result.Append("\n")
			last = i + 1
		} else if curr == ' ' {
			result = result.Append(buf)
			last = i + 1
		}
	}

	return result
}

func matchToken(tokenAsString string) (TokenType, error) {
	if result, found := TokenDict[tokenAsString]; found {
		return result, nil
	} else if tokenAsString != "" && unicode.IsDigit(rune(tokenAsString[0])) {
		_, err := strconv.Atoi(tokenAsString)
		if err != nil {
			panic(err)
		}
		return Int_literal, nil
	} else if tokenAsString != "" && isCharConstant(tokenAsString) {
		return Char_literal, nil
	} else if isAlphabetic(tokenAsString) {
		return Identifier, nil
	} else {
		return -1, fmt.Errorf("Token Not Recognized: %v", tokenAsString)
	}
}

func isAlphabetic(str string) bool {
	for _, r := range str {
		if !unicode.IsLetter(r) {
			return false
		}
	}

	return true
}

func isCharConstant(str string) bool {
	if str[0] == '\'' && str[len(str)-1] == '\'' {
		return true
	}

	return false
}

func view(str string, pos int) rune {
	if pos > len(str) {
		panic(errors.New("Attempted view outside string"))
	}

	return rune(str[pos])
}

package tokenizer

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

type TokenType int

const (
	Exit TokenType = iota + 0
	Open_paren
	Close_paren
	CR
	Operator_plus
	Operator_minus
	Operator_star
	Operator_slash
	Int_literal
)

var TokenDict = map[string]TokenType{
	"exit": Exit,
	"(":    Open_paren,
	")":    Close_paren,
	"\n":   CR,
	"+":    Operator_plus,
	"-":    Operator_minus,
	"*":    Operator_star,
	"/":    Operator_slash,
}

type Token struct {
	Data string
	Kind TokenType
}

func matchToken(tokenAsString string) (TokenType, error) {
	result, found := TokenDict[tokenAsString]
	if found {
		return result, nil
	} else if tokenAsString != "" && unicode.IsDigit(rune(tokenAsString[0])) {
		_, err := strconv.Atoi(tokenAsString)
		Check(err)
		return Int_literal, nil
	} else {
		return -1, fmt.Errorf("Token Not Recognized: %v", tokenAsString)
	}
}

func Tokenize(raw []byte) []Token {
	fileContents := string(raw)

	var result []Token

	last := 0
	for i := 0; i < len(fileContents); i++ {
		buf := fileContents[last:i]

		curr := view(fileContents, i)
		if curr == '(' {
			result = pushToken(result, buf)
			result = pushToken(result, "(")
			last = i + 1
		} else if curr == ')' {
			result = pushToken(result, buf)
			result = pushToken(result, ")")
			last = i + 1
		} else if curr == '\n' {
			result = pushToken(result, buf)
			result = pushToken(result, "\n")
			last = i + 1
		} else if curr == ' ' {
			result = pushToken(result, buf)
			last = i + 1
		}

	}

	return result
}

func pushToken(result []Token, buf string) []Token {
	if buf == "" {
		return result
	}

	kind, err := matchToken(buf)
	Check(err)
	tok := Token{
		Data: buf,
		Kind: kind,
	}
	result = append(result, tok)

	return result
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func view(str string, pos int) rune {
	if pos > len(str) {
		panic(errors.New("Attempted view outside string"))
	}

	return rune(str[pos])
}

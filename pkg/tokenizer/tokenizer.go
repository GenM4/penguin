package tokenizer

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

type TokenType int

const (
	exit TokenType = iota + 0
	open_paren
	close_paren
	cr
	operator
	int_literal
)

type Token struct {
	Data string
	Kind TokenType
}

func matchToken(tokenAsString string) (TokenType, error) {
	if tokenAsString == "exit" {
		return exit, nil
	} else if tokenAsString == "(" {
		return open_paren, nil
	} else if tokenAsString == ")" {
		return close_paren, nil
	} else if tokenAsString == "\n" {
		return cr, nil
	} else if tokenAsString != "" && unicode.IsDigit(rune(tokenAsString[0])) {
		_, err := strconv.Atoi(tokenAsString)
		Check(err)
		return int_literal, nil
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

		curr, err := View(fileContents, i)
		Check(err)

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

func View(str string, pos int) (rune, error) {
	if pos > len(str) {
		return -1, errors.New("EOS")
	}

	return rune(str[pos]), nil
}

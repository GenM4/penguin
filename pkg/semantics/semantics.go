package semantics

import (
	"fmt"
	"strconv"
)

type Type int

const (
	Bool Type = iota + 0
	Int
	Char
	Float
)

func (typ Type) String() string {
	name := []string{
		"Bool",
		"Int",
		"Char",
		"Float",
	}

	i := int(typ)
	switch {
	case i <= int(Float):
		return name[i]
	default:
		return strconv.Itoa(i)
	}

}

type Variable struct {
	Mutable bool
	Type    Type
}

type VarMap map[string]Variable

func MatchType(str string) (Type, error) {
	switch {
	case str == "int":
		return Int, nil
	default:
		return -1, fmt.Errorf("Type %v not implemented", str)
	}
}

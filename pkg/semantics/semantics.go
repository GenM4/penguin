package semantics

import (
	"fmt"
	"strconv"
)

type Type int

const (
	Untyped Type = iota + 0
	Byte
	Bool
	Int
	Char
	Float
)

func (typ Type) String() string {
	name := []string{
		"Untyped",
		"Byte",
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

func (typ Type) Size() int {
	// returns size in bytes
	switch typ {
	case Bool:
		return 2
	case Int:
		return 4
	case Char:
		return 4
	case Float:
		return 8
	default:
		return -1
	}
}

type Variable struct {
	Mutable       bool
	Type          Type
	StackLocation int
	IsGlobal      bool
}

type VarMap map[string]*Variable

type Function struct {
	Mutable   bool
	Type      Type
	Signature string
	NumArgs   int
}

type FuncMap map[string]*Function

func MatchType(str string) (Type, error) {
	switch {
	case str == "int":
		return Int, nil
	case str == "char":
		return Char, nil
	default:
		return -1, fmt.Errorf("Type %v not implemented", str)
	}
}

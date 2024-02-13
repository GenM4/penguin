package semantics

import (
	"fmt"
	"strconv"
)

type Type int

const (
	Byte Type = iota + 0
	Bool
	Int
	Char
	Float
)

func (typ Type) String() string {
	name := []string{
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
}

type VarMap map[string]*Variable

func MatchType(str string) (Type, error) {
	switch {
	case str == "int":
		return Int, nil
	default:
		return -1, fmt.Errorf("Type %v not implemented", str)
	}
}

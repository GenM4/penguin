statement -> declaration
statement -> [...type]...identifier = identifier([...type]...atom)
statement -> identifier = atom
statement -> identifier idOp

declaration -> mutable type identifier(...) scope
declaration -> mutable type identifier

expr -> atom operator atom

atom -> {identifer, expr, term}

term -> literal

type -> {int, char}
mutable -> {mut, const}
operator -> {+, -, *, /}
idOp -> {++, --}


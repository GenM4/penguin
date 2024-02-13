statement -> identifier = atom
statement -> identifier idOp
statement -> exit(atom)

atom -> { identifer, expr, term}

declaration -> mutable type identifier

expr -> term operator term

term -> literal

type -> {int}
mutable -> {mut, const}
operator -> {+, -, *, /, idOp}
idOp -> {++, --}


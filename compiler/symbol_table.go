package compiler

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	// outer symbol table (symbol table we are extending)
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Outer: nil,
		store: make(map[string]Symbol),
	}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := &SymbolTable{
		store: make(map[string]Symbol),
	}
	s.Outer = outer

	return s
}

func (st *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Index: st.numDefinitions,
	}

	// check wether the symbol table is contained within another one
	// and then set the scope accordingly
	if st.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	st.store[name] = symbol
	st.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := st.store[name]

	// if its not on the local symbol table
	// check recursively on the outer ones
	if !ok && st.Outer != nil {
		symbol, ok := st.Outer.Resolve(name)
		return symbol, ok
	}

	return symbol, ok
}

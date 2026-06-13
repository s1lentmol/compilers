package semantic

import "compilers/parser"

type SymbolKind int

const (
	KindVar SymbolKind = iota
	KindFunc
	KindParam
)

type Symbol struct {
	Name     string
	Kind     SymbolKind
	Type     Type
	Arity    int
	Position parser.Position
	Used     bool
}

type Environment struct {
	parent  *Environment
	symbols map[string]*Symbol
	order   []string
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:  parent,
		symbols: make(map[string]*Symbol),
	}
}

func (e *Environment) Define(symbol *Symbol) bool {
	if _, exists := e.symbols[symbol.Name]; exists {
		return false
	}
	e.symbols[symbol.Name] = symbol
	e.order = append(e.order, symbol.Name)
	return true
}

func (e *Environment) Resolve(name string) *Symbol {
	if symbol, exists := e.symbols[name]; exists {
		return symbol
	}
	if e.parent == nil {
		return nil
	}
	return e.parent.Resolve(name)
}

func (e *Environment) IsDefined(name string) bool {
	return e.Resolve(name) != nil
}

func (e *Environment) LocalSymbols() []*Symbol {
	result := make([]*Symbol, 0, len(e.order))
	for _, name := range e.order {
		result = append(result, e.symbols[name])
	}
	return result
}

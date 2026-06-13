package semantic

type Environment struct {
	parent    *Environment
	variables map[string]struct{}
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:    parent,
		variables: make(map[string]struct{}),
	}
}

func (e *Environment) Define(name string) bool {
	if _, exists := e.variables[name]; exists {
		return false
	}

	e.variables[name] = struct{}{}
	return true
}

func (e *Environment) IsDefined(name string) bool {
	if _, exists := e.variables[name]; exists {
		return true
	}

	if e.parent == nil {
		return false
	}

	return e.parent.IsDefined(name)
}

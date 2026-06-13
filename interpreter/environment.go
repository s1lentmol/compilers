package interpreter

type Environment struct {
	parent *Environment
	values map[string]Value
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent: parent,
		values: make(map[string]Value),
	}
}

func (e *Environment) Define(name string, value Value) {
	e.values[name] = value
}

func (e *Environment) Get(name string) (Value, bool) {
	if value, exists := e.values[name]; exists {
		return value, true
	}
	if e.parent == nil {
		return nil, false
	}
	return e.parent.Get(name)
}

func (e *Environment) Assign(name string, value Value) bool {
	if _, exists := e.values[name]; exists {
		e.values[name] = value
		return true
	}
	if e.parent == nil {
		return false
	}
	return e.parent.Assign(name, value)
}

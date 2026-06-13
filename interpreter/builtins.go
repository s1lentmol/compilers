package interpreter

func registerBuiltins(env *Environment) {
	builtins := []*Builtin{
		{
			Name:  "len",
			Arity: 1,
			Fn: func(args []Value) Value {
				switch v := args[0].(type) {
				case *Array:
					return float64(len(v.Elements))
				case string:
					return float64(len([]rune(v)))
				default:
					panic(&RuntimeError{Message: "len() requires array or string, got " + typeName(args[0])})
				}
			},
		},
		{
			Name:  "append",
			Arity: 2,
			Fn: func(args []Value) Value {
				array, ok := args[0].(*Array)
				if !ok {
					panic(&RuntimeError{Message: "append() requires array as first argument, got " + typeName(args[0])})
				}
				elements := make([]Value, len(array.Elements), len(array.Elements)+1)
				copy(elements, array.Elements)
				elements = append(elements, args[1])
				return &Array{Elements: elements}
			},
		},
	}

	for _, builtin := range builtins {
		env.Define(builtin.Name, builtin)
	}
}

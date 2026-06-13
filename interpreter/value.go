package interpreter

import (
	"compilers/parser"
	"strconv"
	"strings"
)

type Value any

type Array struct {
	Elements []Value
}

type Function struct {
	Declaration *parser.FunctionStatement
	Closure     *Environment
}

type Builtin struct {
	Name  string
	Arity int
	Fn    func(args []Value) Value
}

type RuntimeError struct {
	Position parser.Position
	Message  string
}

func (e *RuntimeError) Error() string {
	if e.Position.Line > 0 {
		return "[Runtime Error] line " + strconv.Itoa(e.Position.Line) +
			", column " + strconv.Itoa(e.Position.Column) + ": " + e.Message
	}
	return "[Runtime Error] " + e.Message
}

type returnSignal struct {
	value Value
}

func isTruthy(value Value) bool {
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	case *Array:
		return len(v.Elements) > 0
	default:
		return true
	}
}

func valuesEqual(a, b Value) bool {
	switch av := a.(type) {
	case nil:
		return b == nil
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case *Array:
		bv, ok := b.(*Array)
		return ok && av == bv
	default:
		return false
	}
}

func Stringify(value Value) string {
	switch v := value.(type) {
	case nil:
		return "nil"
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case *Array:
		parts := make([]string, len(v.Elements))
		for i, element := range v.Elements {
			parts[i] = Stringify(element)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *Function:
		return "<func " + v.Declaration.Name + ">"
	case *Builtin:
		return "<builtin " + v.Name + ">"
	default:
		return "<unknown>"
	}
}

func typeName(value Value) string {
	switch value.(type) {
	case nil:
		return "nil"
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "bool"
	case *Array:
		return "array"
	case *Function, *Builtin:
		return "function"
	default:
		return "unknown"
	}
}

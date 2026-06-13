package interpreter

import (
	"compilers/parser"
	"fmt"
	"io"
	"math"
	"os"
)

type Interpreter struct {
	global *Environment
	out    io.Writer
}

func New() *Interpreter {
	global := NewEnvironment(nil)
	registerBuiltins(global)
	return &Interpreter{
		global: global,
		out:    os.Stdout,
	}
}

func (i *Interpreter) SetOutput(w io.Writer) {
	i.out = w
}

func (i *Interpreter) Run(statements []parser.Statement) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if runtimeErr, ok := r.(*RuntimeError); ok {
				err = runtimeErr
				return
			}
			panic(r)
		}
	}()

	i.executeBlock(statements, i.global)
	return nil
}

func (i *Interpreter) executeBlock(statements []parser.Statement, env *Environment) {
	for _, statement := range statements {
		if fn, ok := statement.(*parser.FunctionStatement); ok {
			env.Define(fn.Name, &Function{Declaration: fn, Closure: env})
		}
	}
	for _, statement := range statements {
		i.execute(statement, env)
	}
}

func (i *Interpreter) execute(statement parser.Statement, env *Environment) {
	switch stmt := statement.(type) {
	case *parser.VarStatement:
		var value Value
		if stmt.Initializer != nil {
			value = i.evaluate(stmt.Initializer, env)
		}
		env.Define(stmt.Name, value)

	case *parser.PrintStatement:
		value := i.evaluate(stmt.Value, env)
		fmt.Fprintln(i.out, Stringify(value))

	case *parser.ExpressionStatement:
		i.evaluate(stmt.Expression, env)

	case *parser.BlockStatement:
		i.executeBlock(stmt.Statements, NewEnvironment(env))

	case *parser.IfStatement:
		if isTruthy(i.evaluate(stmt.Condition, env)) {
			i.execute(stmt.ThenBranch, env)
		} else if stmt.ElseBranch != nil {
			i.execute(stmt.ElseBranch, env)
		}

	case *parser.WhileStatement:
		for isTruthy(i.evaluate(stmt.Condition, env)) {
			i.execute(stmt.Body, env)
		}

	case *parser.FunctionStatement:
		env.Define(stmt.Name, &Function{Declaration: stmt, Closure: env})

	case *parser.ReturnStatement:
		var value Value
		if stmt.Value != nil {
			value = i.evaluate(stmt.Value, env)
		}
		panic(returnSignal{value: value})

	default:
		panic(&RuntimeError{Message: fmt.Sprintf("unsupported statement: %T", statement)})
	}
}

func (i *Interpreter) evaluate(expression parser.Expression, env *Environment) Value {
	switch expr := expression.(type) {
	case *parser.NumberExpression:
		return expr.Value

	case *parser.StringExpression:
		return expr.Value

	case *parser.BooleanExpression:
		return expr.Value

	case *parser.VariableExpression:
		value, ok := env.Get(expr.Name)
		if !ok {
			panic(&RuntimeError{Position: expr.Position, Message: fmt.Sprintf("undefined variable %q", expr.Name)})
		}
		return value

	case *parser.AssignExpression:
		value := i.evaluate(expr.Value, env)
		if !env.Assign(expr.Name, value) {
			panic(&RuntimeError{Position: expr.Position, Message: fmt.Sprintf("undefined variable %q", expr.Name)})
		}
		return value

	case *parser.UnaryExpression:
		return i.evaluateUnary(expr, env)

	case *parser.BinaryExpression:
		return i.evaluateBinary(expr, env)

	case *parser.CallExpression:
		return i.evaluateCall(expr, env)

	case *parser.ArrayExpression:
		elements := make([]Value, len(expr.Elements))
		for idx, element := range expr.Elements {
			elements[idx] = i.evaluate(element, env)
		}
		return &Array{Elements: elements}

	case *parser.IndexExpression:
		return i.evaluateIndex(expr, env)

	case *parser.IndexAssignExpression:
		return i.evaluateIndexAssign(expr, env)

	default:
		panic(&RuntimeError{Message: fmt.Sprintf("unsupported expression: %T", expression)})
	}
}

func (i *Interpreter) evaluateUnary(expr *parser.UnaryExpression, env *Environment) Value {
	operand := i.evaluate(expr.Right, env)
	switch expr.Operator {
	case "-":
		number, ok := operand.(float64)
		if !ok {
			panic(&RuntimeError{Position: expr.Position, Message: fmt.Sprintf("operator '-' requires number, got %s", typeName(operand))})
		}
		return -number
	case "!":
		return !isTruthy(operand)
	default:
		panic(&RuntimeError{Position: expr.Position, Message: "unknown unary operator " + expr.Operator})
	}
}

func (i *Interpreter) evaluateBinary(expr *parser.BinaryExpression, env *Environment) Value {
	if expr.Operator == "&&" {
		left := i.evaluate(expr.Left, env)
		if !isTruthy(left) {
			return false
		}
		return isTruthy(i.evaluate(expr.Right, env))
	}
	if expr.Operator == "||" {
		left := i.evaluate(expr.Left, env)
		if isTruthy(left) {
			return true
		}
		return isTruthy(i.evaluate(expr.Right, env))
	}

	left := i.evaluate(expr.Left, env)
	right := i.evaluate(expr.Right, env)

	switch expr.Operator {
	case "==":
		return valuesEqual(left, right)
	case "!=":
		return !valuesEqual(left, right)
	case "+":
		return i.evalPlus(left, right, expr.Position)
	case "<", "<=", ">", ">=":
		return i.evalComparison(expr.Operator, left, right, expr.Position)
	}

	leftNum, rightNum := i.numericOperands(expr.Operator, left, right, expr.Position)
	switch expr.Operator {
	case "-":
		return leftNum - rightNum
	case "*":
		return leftNum * rightNum
	case "/":
		if rightNum == 0 {
			panic(&RuntimeError{Position: expr.Position, Message: "division by zero"})
		}
		return leftNum / rightNum
	case "%":
		if rightNum == 0 {
			panic(&RuntimeError{Position: expr.Position, Message: "modulo by zero"})
		}
		return math.Mod(leftNum, rightNum)
	default:
		panic(&RuntimeError{Position: expr.Position, Message: "unknown operator " + expr.Operator})
	}
}

func (i *Interpreter) evalPlus(left, right Value, position parser.Position) Value {
	if leftNum, ok := left.(float64); ok {
		if rightNum, ok := right.(float64); ok {
			return leftNum + rightNum
		}
	}
	if leftStr, ok := left.(string); ok {
		if rightStr, ok := right.(string); ok {
			return leftStr + rightStr
		}
	}
	panic(&RuntimeError{Position: position, Message: fmt.Sprintf("operator '+' cannot be applied to %s and %s", typeName(left), typeName(right))})
}

func (i *Interpreter) evalComparison(operator string, left, right Value, position parser.Position) Value {
	if leftNum, ok := left.(float64); ok {
		if rightNum, ok := right.(float64); ok {
			switch operator {
			case "<":
				return leftNum < rightNum
			case "<=":
				return leftNum <= rightNum
			case ">":
				return leftNum > rightNum
			case ">=":
				return leftNum >= rightNum
			}
		}
	}
	if leftStr, ok := left.(string); ok {
		if rightStr, ok := right.(string); ok {
			switch operator {
			case "<":
				return leftStr < rightStr
			case "<=":
				return leftStr <= rightStr
			case ">":
				return leftStr > rightStr
			case ">=":
				return leftStr >= rightStr
			}
		}
	}
	panic(&RuntimeError{Position: position, Message: fmt.Sprintf("operator '%s' cannot compare %s and %s", operator, typeName(left), typeName(right))})
}

func (i *Interpreter) numericOperands(operator string, left, right Value, position parser.Position) (float64, float64) {
	leftNum, ok := left.(float64)
	if !ok {
		panic(&RuntimeError{Position: position, Message: fmt.Sprintf("operator '%s' requires numbers, left operand is %s", operator, typeName(left))})
	}
	rightNum, ok := right.(float64)
	if !ok {
		panic(&RuntimeError{Position: position, Message: fmt.Sprintf("operator '%s' requires numbers, right operand is %s", operator, typeName(right))})
	}
	return leftNum, rightNum
}

func (i *Interpreter) evaluateCall(expr *parser.CallExpression, env *Environment) Value {
	callee := i.evaluate(expr.Callee, env)

	arguments := make([]Value, len(expr.Arguments))
	for idx, argument := range expr.Arguments {
		arguments[idx] = i.evaluate(argument, env)
	}

	switch fn := callee.(type) {
	case *Function:
		return i.callFunction(fn, arguments, expr.Position)
	case *Builtin:
		if fn.Arity >= 0 && len(arguments) != fn.Arity {
			panic(&RuntimeError{Position: expr.Position, Message: fmt.Sprintf("%s expects %d argument(s), got %d", fn.Name, fn.Arity, len(arguments))})
		}
		return fn.Fn(arguments)
	default:
		panic(&RuntimeError{Position: expr.Position, Message: fmt.Sprintf("%s is not callable", typeName(callee))})
	}
}

func (i *Interpreter) callFunction(fn *Function, arguments []Value, position parser.Position) (result Value) {
	if len(arguments) != len(fn.Declaration.Parameters) {
		panic(&RuntimeError{Position: position, Message: fmt.Sprintf("function %q expects %d argument(s), got %d",
			fn.Declaration.Name, len(fn.Declaration.Parameters), len(arguments))})
	}

	frame := NewEnvironment(fn.Closure)
	for idx, param := range fn.Declaration.Parameters {
		frame.Define(param, arguments[idx])
	}

	defer func() {
		if r := recover(); r != nil {
			if signal, ok := r.(returnSignal); ok {
				result = signal.value
				return
			}
			panic(r)
		}
	}()

	i.executeBlock(fn.Declaration.Body.Statements, frame)
	return nil
}

func (i *Interpreter) evaluateIndex(expr *parser.IndexExpression, env *Environment) Value {
	collection := i.evaluate(expr.Collection, env)
	index := i.evaluate(expr.Index, env)

	switch coll := collection.(type) {
	case *Array:
		idx := i.indexValue(index, len(coll.Elements), expr.Position)
		return coll.Elements[idx]
	case string:
		runes := []rune(coll)
		idx := i.indexValue(index, len(runes), expr.Position)
		return string(runes[idx])
	default:
		panic(&RuntimeError{Position: expr.Position, Message: fmt.Sprintf("type %s is not indexable", typeName(collection))})
	}
}

func (i *Interpreter) evaluateIndexAssign(expr *parser.IndexAssignExpression, env *Environment) Value {
	collection := i.evaluate(expr.Collection, env)
	index := i.evaluate(expr.Index, env)
	value := i.evaluate(expr.Value, env)

	array, ok := collection.(*Array)
	if !ok {
		panic(&RuntimeError{Position: expr.Position, Message: fmt.Sprintf("cannot assign by index to %s", typeName(collection))})
	}

	idx := i.indexValue(index, len(array.Elements), expr.Position)
	array.Elements[idx] = value
	return value
}

func (i *Interpreter) indexValue(index Value, length int, position parser.Position) int {
	number, ok := index.(float64)
	if !ok {
		panic(&RuntimeError{Position: position, Message: fmt.Sprintf("index must be a number, got %s", typeName(index))})
	}
	idx := int(number)
	if float64(idx) != number {
		panic(&RuntimeError{Position: position, Message: "index must be an integer"})
	}
	if idx < 0 || idx >= length {
		panic(&RuntimeError{Position: position, Message: fmt.Sprintf("index %d out of bounds [0, %d)", idx, length)})
	}
	return idx
}

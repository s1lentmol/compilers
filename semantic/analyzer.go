package semantic

import (
	"compilers/parser"
	"fmt"
)

type Type int

const (
	TypeUnknown Type = iota
	TypeNumber
	TypeString
	TypeBool
	TypeArray
	TypeFunction
	TypeNil
)

func (t Type) String() string {
	switch t {
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeArray:
		return "array"
	case TypeFunction:
		return "function"
	case TypeNil:
		return "nil"
	default:
		return "unknown"
	}
}

type Result struct {
	Errors   []error
	Warnings []string
}

type Analyzer struct {
	environment *Environment
	result      Result
}

func Analyze(statements []parser.Statement) Result {
	analyzer := &Analyzer{
		environment: NewEnvironment(nil),
	}
	analyzer.registerBuiltins()
	analyzer.analyzeBlock(statements)
	analyzer.reportUnused(analyzer.environment)
	return analyzer.result
}

func (a *Analyzer) analyzeBlock(statements []parser.Statement) {
	for _, statement := range statements {
		if fn, ok := statement.(*parser.FunctionStatement); ok {
			symbol := &Symbol{
				Name:     fn.Name,
				Kind:     KindFunc,
				Type:     TypeFunction,
				Arity:    len(fn.Parameters),
				Position: fn.Position,
			}
			if !a.environment.Define(symbol) {
				a.addError(fn.Position, "function %q is already defined in this scope", fn.Name)
			}
		}
	}

	for _, statement := range statements {
		a.visitStatement(statement)
	}
}

func (a *Analyzer) visitStatement(statement parser.Statement) {
	switch stmt := statement.(type) {
	case *parser.VarStatement:
		var inferred Type = TypeUnknown
		if stmt.Initializer != nil {
			inferred = a.visitExpression(stmt.Initializer)
		}
		symbol := &Symbol{
			Name:     stmt.Name,
			Kind:     KindVar,
			Type:     inferred,
			Position: stmt.Position,
		}
		if !a.environment.Define(symbol) {
			a.addError(stmt.Position, "variable %q is already defined in this scope", stmt.Name)
		}

	case *parser.PrintStatement:
		a.visitExpression(stmt.Value)

	case *parser.ExpressionStatement:
		a.visitExpression(stmt.Expression)

	case *parser.BlockStatement:
		a.withScope(func() {
			a.analyzeBlock(stmt.Statements)
		})

	case *parser.IfStatement:
		a.visitExpression(stmt.Condition)
		a.visitStatement(stmt.ThenBranch)
		if stmt.ElseBranch != nil {
			a.visitStatement(stmt.ElseBranch)
		}

	case *parser.WhileStatement:
		a.visitExpression(stmt.Condition)
		a.visitStatement(stmt.Body)

	case *parser.FunctionStatement:
		a.withScope(func() {
			for _, param := range stmt.Parameters {
				a.environment.Define(&Symbol{
					Name:     param,
					Kind:     KindParam,
					Type:     TypeUnknown,
					Position: stmt.Position,
				})
			}
			a.analyzeBlock(stmt.Body.Statements)
		})

	case *parser.ReturnStatement:
		if stmt.Value != nil {
			a.visitExpression(stmt.Value)
		}

	default:
		a.result.Errors = append(a.result.Errors, fmt.Errorf("unsupported statement type: %T", statement))
	}
}

func (a *Analyzer) visitExpression(expression parser.Expression) Type {
	switch expr := expression.(type) {
	case *parser.NumberExpression:
		return TypeNumber

	case *parser.StringExpression:
		return TypeString

	case *parser.BooleanExpression:
		return TypeBool

	case *parser.VariableExpression:
		symbol := a.environment.Resolve(expr.Name)
		if symbol == nil {
			a.addError(expr.Position, "variable %q is not defined", expr.Name)
			return TypeUnknown
		}
		symbol.Used = true
		return symbol.Type

	case *parser.AssignExpression:
		valueType := a.visitExpression(expr.Value)
		symbol := a.environment.Resolve(expr.Name)
		if symbol == nil {
			a.addError(expr.Position, "variable %q is not defined", expr.Name)
			return TypeUnknown
		}
		symbol.Used = true
		symbol.Type = valueType
		return valueType

	case *parser.BinaryExpression:
		left := a.visitExpression(expr.Left)
		right := a.visitExpression(expr.Right)
		return a.checkBinary(expr.Operator, left, right, expr.Position)

	case *parser.UnaryExpression:
		operand := a.visitExpression(expr.Right)
		if expr.Operator == "-" {
			if operand != TypeUnknown && operand != TypeNumber {
				a.addError(expr.Position, "unary operator '-' requires number, got %s", operand)
			}
			return TypeNumber
		}
		return TypeBool

	case *parser.CallExpression:
		return a.visitCall(expr)

	case *parser.ArrayExpression:
		for _, element := range expr.Elements {
			a.visitExpression(element)
		}
		return TypeArray

	case *parser.IndexExpression:
		collection := a.visitExpression(expr.Collection)
		index := a.visitExpression(expr.Index)
		a.checkIndex(collection, index, expr.Position)
		return TypeUnknown

	case *parser.IndexAssignExpression:
		collection := a.visitExpression(expr.Collection)
		index := a.visitExpression(expr.Index)
		a.checkIndex(collection, index, expr.Position)
		return a.visitExpression(expr.Value)

	default:
		a.result.Errors = append(a.result.Errors, fmt.Errorf("unsupported expression type: %T", expression))
		return TypeUnknown
	}
}

func (a *Analyzer) visitCall(expr *parser.CallExpression) Type {
	for _, argument := range expr.Arguments {
		a.visitExpression(argument)
	}

	if callee, ok := expr.Callee.(*parser.VariableExpression); ok {
		symbol := a.environment.Resolve(callee.Name)
		if symbol == nil {
			a.addError(callee.Position, "function %q is not defined", callee.Name)
			return TypeUnknown
		}
		symbol.Used = true
		if symbol.Kind != KindFunc && symbol.Type != TypeUnknown {
			a.addError(expr.Position, "%q is not a function", callee.Name)
			return TypeUnknown
		}
		if symbol.Kind == KindFunc && len(expr.Arguments) != symbol.Arity {
			a.addError(expr.Position, "function %q expects %d argument(s), got %d", callee.Name, symbol.Arity, len(expr.Arguments))
		}
		return TypeUnknown
	}

	a.visitExpression(expr.Callee)
	return TypeUnknown
}

func (a *Analyzer) checkBinary(operator string, left, right Type, position parser.Position) Type {
	switch operator {
	case "+":
		if left == TypeNumber && right == TypeNumber {
			return TypeNumber
		}
		if left == TypeString && right == TypeString {
			return TypeString
		}
		if left != TypeUnknown && right != TypeUnknown {
			a.addError(position, "operator '+' cannot be applied to types %s and %s", left, right)
		}
		if left == TypeString || right == TypeString {
			return TypeString
		}
		return TypeUnknown

	case "-", "*", "/", "%":
		if left != TypeUnknown && left != TypeNumber {
			a.addError(position, "operator '%s' requires numbers, left operand is %s", operator, left)
		}
		if right != TypeUnknown && right != TypeNumber {
			a.addError(position, "operator '%s' requires numbers, right operand is %s", operator, right)
		}
		return TypeNumber

	case "<", "<=", ">", ">=":
		a.checkOrdered(operator, left, position)
		a.checkOrdered(operator, right, position)
		if left != TypeUnknown && right != TypeUnknown && left != right {
			a.addError(position, "operator '%s' cannot compare %s and %s", operator, left, right)
		}
		return TypeBool

	case "==", "!=":
		return TypeBool

	case "&&", "||":
		return TypeBool

	default:
		return TypeUnknown
	}
}

func (a *Analyzer) checkOrdered(operator string, t Type, position parser.Position) {
	if t != TypeUnknown && t != TypeNumber && t != TypeString {
		a.addError(position, "operator '%s' cannot be applied to type %s", operator, t)
	}
}

func (a *Analyzer) checkIndex(collection, index Type, position parser.Position) {
	if collection != TypeUnknown && collection != TypeArray && collection != TypeString {
		a.addError(position, "type %s is not indexable", collection)
	}
	if index != TypeUnknown && index != TypeNumber {
		a.addError(position, "index must be a number, got %s", index)
	}
}

func (a *Analyzer) withScope(fn func()) {
	previous := a.environment
	a.environment = NewEnvironment(previous)
	fn()
	a.reportUnused(a.environment)
	a.environment = previous
}

func (a *Analyzer) reportUnused(env *Environment) {
	for _, symbol := range env.LocalSymbols() {
		if symbol.Used || symbol.Kind == KindParam {
			continue
		}
		kind := "variable"
		if symbol.Kind == KindFunc {
			kind = "function"
		}
		warning := fmt.Sprintf("%s %q is declared but never used", kind, symbol.Name)
		if symbol.Position.Line > 0 {
			warning = fmt.Sprintf("line %d, column %d: %s", symbol.Position.Line, symbol.Position.Column, warning)
		}
		a.result.Warnings = append(a.result.Warnings, warning)
	}
}

func (a *Analyzer) registerBuiltins() {
	builtins := []struct {
		name  string
		arity int
	}{
		{"len", 1},
		{"append", 2},
	}
	for _, builtin := range builtins {
		a.environment.Define(&Symbol{
			Name:  builtin.name,
			Kind:  KindFunc,
			Type:  TypeFunction,
			Arity: builtin.arity,
			Used:  true,
		})
	}
}

func (a *Analyzer) addError(position parser.Position, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if position.Line > 0 {
		message = fmt.Sprintf("line %d, column %d: %s", position.Line, position.Column, message)
	}
	a.result.Errors = append(a.result.Errors, fmt.Errorf("%s", message))
}

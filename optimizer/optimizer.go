package optimizer

import (
	"compilers/parser"
	"math"
)

func Optimize(statements []parser.Statement) []parser.Statement {
	return optimizeStatements(statements)
}

func optimizeStatements(statements []parser.Statement) []parser.Statement {
	result := make([]parser.Statement, 0, len(statements))
	for _, statement := range statements {
		optimized := optimizeStatement(statement)
		if optimized == nil {
			continue
		}
		result = append(result, optimized)
		if _, isReturn := optimized.(*parser.ReturnStatement); isReturn {
			break
		}
	}
	return result
}

func optimizeStatement(statement parser.Statement) parser.Statement {
	switch stmt := statement.(type) {
	case *parser.VarStatement:
		if stmt.Initializer != nil {
			stmt.Initializer = optimizeExpression(stmt.Initializer)
		}
		return stmt

	case *parser.PrintStatement:
		stmt.Value = optimizeExpression(stmt.Value)
		return stmt

	case *parser.ExpressionStatement:
		stmt.Expression = optimizeExpression(stmt.Expression)
		return stmt

	case *parser.ReturnStatement:
		if stmt.Value != nil {
			stmt.Value = optimizeExpression(stmt.Value)
		}
		return stmt

	case *parser.BlockStatement:
		stmt.Statements = optimizeStatements(stmt.Statements)
		return stmt

	case *parser.FunctionStatement:
		stmt.Body.Statements = optimizeStatements(stmt.Body.Statements)
		return stmt

	case *parser.IfStatement:
		return optimizeIf(stmt)

	case *parser.WhileStatement:
		return optimizeWhile(stmt)

	default:
		return statement
	}
}

func optimizeIf(stmt *parser.IfStatement) parser.Statement {
	stmt.Condition = optimizeExpression(stmt.Condition)

	if value, ok := literalValue(stmt.Condition); ok {
		if isTruthy(value) {
			return optimizeStatement(stmt.ThenBranch)
		}
		if stmt.ElseBranch != nil {
			return optimizeStatement(stmt.ElseBranch)
		}
		return nil
	}

	stmt.ThenBranch = optimizeStatement(stmt.ThenBranch)
	if stmt.ThenBranch == nil {
		stmt.ThenBranch = &parser.BlockStatement{Position: stmt.Position}
	}
	if stmt.ElseBranch != nil {
		stmt.ElseBranch = optimizeStatement(stmt.ElseBranch)
	}
	return stmt
}

func optimizeWhile(stmt *parser.WhileStatement) parser.Statement {
	stmt.Condition = optimizeExpression(stmt.Condition)

	if value, ok := literalValue(stmt.Condition); ok && !isTruthy(value) {
		return nil
	}

	stmt.Body = optimizeStatement(stmt.Body)
	if stmt.Body == nil {
		stmt.Body = &parser.BlockStatement{Position: stmt.Position}
	}
	return stmt
}

func optimizeExpression(expression parser.Expression) parser.Expression {
	switch expr := expression.(type) {
	case *parser.UnaryExpression:
		expr.Right = optimizeExpression(expr.Right)
		if folded, ok := foldUnary(expr); ok {
			return folded
		}
		return expr

	case *parser.BinaryExpression:
		expr.Left = optimizeExpression(expr.Left)
		expr.Right = optimizeExpression(expr.Right)
		if folded, ok := foldBinary(expr); ok {
			return folded
		}
		return expr

	case *parser.AssignExpression:
		expr.Value = optimizeExpression(expr.Value)
		return expr

	case *parser.CallExpression:
		expr.Callee = optimizeExpression(expr.Callee)
		for i := range expr.Arguments {
			expr.Arguments[i] = optimizeExpression(expr.Arguments[i])
		}
		return expr

	case *parser.ArrayExpression:
		for i := range expr.Elements {
			expr.Elements[i] = optimizeExpression(expr.Elements[i])
		}
		return expr

	case *parser.IndexExpression:
		expr.Collection = optimizeExpression(expr.Collection)
		expr.Index = optimizeExpression(expr.Index)
		return expr

	case *parser.IndexAssignExpression:
		expr.Collection = optimizeExpression(expr.Collection)
		expr.Index = optimizeExpression(expr.Index)
		expr.Value = optimizeExpression(expr.Value)
		return expr

	default:
		return expression
	}
}

func foldUnary(expr *parser.UnaryExpression) (parser.Expression, bool) {
	value, ok := literalValue(expr.Right)
	if !ok {
		return nil, false
	}
	switch expr.Operator {
	case "-":
		if number, ok := value.(float64); ok {
			return &parser.NumberExpression{Position: expr.Position, Value: -number}, true
		}
	case "!":
		return &parser.BooleanExpression{Position: expr.Position, Value: !isTruthy(value)}, true
	}
	return nil, false
}

func foldBinary(expr *parser.BinaryExpression) (parser.Expression, bool) {
	left, leftOk := literalValue(expr.Left)
	right, rightOk := literalValue(expr.Right)
	if !leftOk || !rightOk {
		return nil, false
	}

	pos := expr.Position
	op := expr.Operator

	switch op {
	case "==":
		return &parser.BooleanExpression{Position: pos, Value: literalsEqual(left, right)}, true
	case "!=":
		return &parser.BooleanExpression{Position: pos, Value: !literalsEqual(left, right)}, true
	}

	if leftNum, ok := left.(float64); ok {
		if rightNum, ok := right.(float64); ok {
			return foldNumeric(op, leftNum, rightNum, pos)
		}
	}

	if leftStr, ok := left.(string); ok {
		if rightStr, ok := right.(string); ok {
			return foldString(op, leftStr, rightStr, pos)
		}
	}

	if leftBool, ok := left.(bool); ok {
		if rightBool, ok := right.(bool); ok {
			switch op {
			case "&&":
				return &parser.BooleanExpression{Position: pos, Value: leftBool && rightBool}, true
			case "||":
				return &parser.BooleanExpression{Position: pos, Value: leftBool || rightBool}, true
			}
		}
	}

	return nil, false
}

func foldNumeric(op string, a, b float64, pos parser.Position) (parser.Expression, bool) {
	switch op {
	case "+":
		return &parser.NumberExpression{Position: pos, Value: a + b}, true
	case "-":
		return &parser.NumberExpression{Position: pos, Value: a - b}, true
	case "*":
		return &parser.NumberExpression{Position: pos, Value: a * b}, true
	case "/":
		if b == 0 {
			return nil, false
		}
		return &parser.NumberExpression{Position: pos, Value: a / b}, true
	case "%":
		if b == 0 {
			return nil, false
		}
		return &parser.NumberExpression{Position: pos, Value: math.Mod(a, b)}, true
	case "<":
		return &parser.BooleanExpression{Position: pos, Value: a < b}, true
	case "<=":
		return &parser.BooleanExpression{Position: pos, Value: a <= b}, true
	case ">":
		return &parser.BooleanExpression{Position: pos, Value: a > b}, true
	case ">=":
		return &parser.BooleanExpression{Position: pos, Value: a >= b}, true
	}
	return nil, false
}

func foldString(op string, a, b string, pos parser.Position) (parser.Expression, bool) {
	switch op {
	case "+":
		return &parser.StringExpression{Position: pos, Value: a + b}, true
	case "<":
		return &parser.BooleanExpression{Position: pos, Value: a < b}, true
	case "<=":
		return &parser.BooleanExpression{Position: pos, Value: a <= b}, true
	case ">":
		return &parser.BooleanExpression{Position: pos, Value: a > b}, true
	case ">=":
		return &parser.BooleanExpression{Position: pos, Value: a >= b}, true
	}
	return nil, false
}

func literalValue(expr parser.Expression) (any, bool) {
	switch e := expr.(type) {
	case *parser.NumberExpression:
		return e.Value, true
	case *parser.StringExpression:
		return e.Value, true
	case *parser.BooleanExpression:
		return e.Value, true
	default:
		return nil, false
	}
}

func literalsEqual(a, b any) bool {
	switch av := a.(type) {
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	default:
		return false
	}
}

func isTruthy(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	default:
		return true
	}
}

package semantic

import (
	"compilers/parser"
	"fmt"
)

type Analyzer struct {
	environment *Environment
	errors      []error
}

func Analyze(statements []parser.Statement) []error {
	analyzer := &Analyzer{
		environment: NewEnvironment(nil),
	}
	analyzer.Analyze(statements)
	return analyzer.errors
}

func (a *Analyzer) Analyze(statements []parser.Statement) {
	for _, statement := range statements {
		a.visitStatement(statement)
	}
}

func (a *Analyzer) visitStatement(statement parser.Statement) {
	switch stmt := statement.(type) {
	case *parser.VarStatement:
		if stmt.Initializer != nil {
			a.visitExpression(stmt.Initializer)
		}
		if !a.environment.Define(stmt.Name) {
			a.addError(stmt.Position, "variable %q is already defined in this scope", stmt.Name)
		}
	case *parser.PrintStatement:
		a.visitExpression(stmt.Value)
	case *parser.ExpressionStatement:
		a.visitExpression(stmt.Expression)
	case *parser.BlockStatement:
		a.withEnvironment(func() {
			for _, innerStatement := range stmt.Statements {
				a.visitStatement(innerStatement)
			}
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
	default:
		a.errors = append(a.errors, fmt.Errorf("unsupported statement type: %T", statement))
	}
}

func (a *Analyzer) visitExpression(expression parser.Expression) {
	switch expr := expression.(type) {
	case *parser.NumberExpression:
		return
	case *parser.VariableExpression:
		if !a.environment.IsDefined(expr.Name) {
			a.addError(expr.Position, "variable %q is not defined", expr.Name)
		}
	case *parser.AssignExpression:
		a.visitExpression(expr.Value)
		if !a.environment.IsDefined(expr.Name) {
			a.addError(expr.Position, "variable %q is not defined", expr.Name)
		}
	case *parser.BinaryExpression:
		a.visitExpression(expr.Left)
		a.visitExpression(expr.Right)
	case *parser.UnaryExpression:
		a.visitExpression(expr.Right)
	default:
		a.errors = append(a.errors, fmt.Errorf("unsupported expression type: %T", expression))
	}
}

func (a *Analyzer) withEnvironment(fn func()) {
	previous := a.environment
	a.environment = NewEnvironment(previous)
	defer func() {
		a.environment = previous
	}()
	fn()
}

func (a *Analyzer) addError(position parser.Position, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if position.Line > 0 {
		message = fmt.Sprintf("line %d, column %d: %s", position.Line, position.Column, message)
	}
	a.errors = append(a.errors, fmt.Errorf("%s", message))
}

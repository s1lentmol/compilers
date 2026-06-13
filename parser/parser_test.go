package parser

import (
	"compilers/lexer"
	"testing"
)

func parse(t *testing.T, code string) []Statement {
	t.Helper()
	tokens := lexer.NewLexer(code).Tokenize()
	return NewParser(tokens).Parse()
}

func TestPrecedence(t *testing.T) {
	stmts := parse(t, "2 + 3 * 4;")
	expr := stmts[0].(*ExpressionStatement).Expression.(*BinaryExpression)
	if expr.Operator != "+" {
		t.Fatalf("top operator should be +, got %s", expr.Operator)
	}
	right := expr.Right.(*BinaryExpression)
	if right.Operator != "*" {
		t.Fatalf("right operator should be *, got %s", right.Operator)
	}
}

func TestAssociativityOfAssignment(t *testing.T) {
	stmts := parse(t, "a = b = 1;")
	assign := stmts[0].(*ExpressionStatement).Expression.(*AssignExpression)
	if assign.Name != "a" {
		t.Fatalf("outer assign target should be a, got %s", assign.Name)
	}
	if _, ok := assign.Value.(*AssignExpression); !ok {
		t.Fatalf("assignment should be right-associative")
	}
}

func TestFunctionDeclaration(t *testing.T) {
	stmts := parse(t, "func add(a, b) { return a + b; }")
	fn, ok := stmts[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", stmts[0])
	}
	if fn.Name != "add" || len(fn.Parameters) != 2 {
		t.Fatalf("unexpected function: %+v", fn)
	}
}

func TestArrayAndIndex(t *testing.T) {
	stmts := parse(t, "var a = [1, 2]; a[0] = 5;")
	arr := stmts[0].(*VarStatement).Initializer.(*ArrayExpression)
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}
	if _, ok := stmts[1].(*ExpressionStatement).Expression.(*IndexAssignExpression); !ok {
		t.Fatalf("expected IndexAssignExpression")
	}
}

func TestCallChain(t *testing.T) {
	stmts := parse(t, "f(1)(2);")
	outer := stmts[0].(*ExpressionStatement).Expression.(*CallExpression)
	if _, ok := outer.Callee.(*CallExpression); !ok {
		t.Fatalf("expected nested call expression")
	}
}

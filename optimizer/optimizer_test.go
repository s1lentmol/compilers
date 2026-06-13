package optimizer

import (
	"compilers/lexer"
	"compilers/parser"
	"testing"
)

func optimize(t *testing.T, code string) []parser.Statement {
	t.Helper()
	tokens := lexer.NewLexer(code).Tokenize()
	statements := parser.NewParser(tokens).Parse()
	return Optimize(statements)
}

func TestConstantFoldingNumeric(t *testing.T) {
	stmts := optimize(t, "print 2 * 3 + 4;")
	value := stmts[0].(*parser.PrintStatement).Value
	number, ok := value.(*parser.NumberExpression)
	if !ok {
		t.Fatalf("expected folded NumberExpression, got %T", value)
	}
	if number.Value != 10 {
		t.Errorf("got %v, want 10", number.Value)
	}
}

func TestConstantFoldingString(t *testing.T) {
	stmts := optimize(t, `print "a" + "b";`)
	value := stmts[0].(*parser.PrintStatement).Value
	str, ok := value.(*parser.StringExpression)
	if !ok {
		t.Fatalf("expected folded StringExpression, got %T", value)
	}
	if str.Value != "ab" {
		t.Errorf("got %q, want %q", str.Value, "ab")
	}
}

func TestDeadCodeEliminationIf(t *testing.T) {
	stmts := optimize(t, "if (false) print 1; else print 2;")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	print, ok := stmts[0].(*parser.PrintStatement)
	if !ok {
		t.Fatalf("expected PrintStatement, got %T", stmts[0])
	}
	if print.Value.(*parser.NumberExpression).Value != 2 {
		t.Errorf("wrong branch survived")
	}
}

func TestDeadCodeEliminationWhile(t *testing.T) {
	stmts := optimize(t, "while (false) print 1;")
	if len(stmts) != 0 {
		t.Fatalf("expected while(false) to be removed, got %d statements", len(stmts))
	}
}

func TestUnreachableAfterReturn(t *testing.T) {
	stmts := optimize(t, "func f() { return 1; print 2; }")
	body := stmts[0].(*parser.FunctionStatement).Body
	if len(body.Statements) != 1 {
		t.Fatalf("expected unreachable code removed, got %d statements", len(body.Statements))
	}
}

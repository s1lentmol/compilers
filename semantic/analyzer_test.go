package semantic

import (
	"compilers/lexer"
	"compilers/parser"
	"strings"
	"testing"
)

func TestAnalyze(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		code           string
		wantErrors     []string
		wantErrorCount int
	}{
		{
			name: "declared variable in same scope",
			code: "var x = 1;\nprint x;",
		},
		{
			name:           "variable used before declaration",
			code:           "print x;\nvar x = 1;",
			wantErrors:     []string{`line 1, column 7: variable "x" is not defined`},
			wantErrorCount: 1,
		},
		{
			name:           "assignment to undeclared variable",
			code:           "x = 3;",
			wantErrors:     []string{`line 1, column 1: variable "x" is not defined`},
			wantErrorCount: 1,
		},
		{
			name: "outer variable visible in inner block",
			code: "var x = 1;\n{\nprint x;\n}",
		},
		{
			name:           "inner variable not visible outside block",
			code:           "{\nvar x = 1;\n}\nprint x;",
			wantErrors:     []string{`line 4, column 7: variable "x" is not defined`},
			wantErrorCount: 1,
		},
		{
			name:           "duplicate declaration in same block",
			code:           "var x = 1;\nvar x = 2;",
			wantErrors:     []string{`line 2, column 1: variable "x" is already defined in this scope`},
			wantErrorCount: 1,
		},
		{
			name: "shadowing in nested block is allowed",
			code: "var x = 1;\n{\nvar x = 2;\nprint x;\n}\nprint x;",
		},
		{
			name: "if and while conditions report undeclared variables",
			code: "if (flag) print 1;\nwhile (count) print 2;",
			wantErrors: []string{
				`line 1, column 5: variable "flag" is not defined`,
				`line 2, column 8: variable "count" is not defined`,
			},
			wantErrorCount: 2,
		},
		{
			name: "collect multiple errors",
			code: "print x;\ny = z;\nvar x = 1;\nvar x = 2;",
			wantErrors: []string{
				`line 1, column 7: variable "x" is not defined`,
				`line 2, column 5: variable "z" is not defined`,
				`line 2, column 1: variable "y" is not defined`,
				`line 4, column 1: variable "x" is already defined in this scope`,
			},
			wantErrorCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := lexer.NewLexer(tt.code)
			tokens := l.Tokenize()
			p := parser.NewParser(tokens)
			statements := p.Parse()

			errors := Analyze(statements).Errors
			if len(errors) != tt.wantErrorCount {
				t.Fatalf("got %d errors, want %d: %v", len(errors), tt.wantErrorCount, stringifyErrors(errors))
			}

			for _, want := range tt.wantErrors {
				if !containsError(errors, want) {
					t.Fatalf("expected error %q, got %v", want, stringifyErrors(errors))
				}
			}
		})
	}
}

func TestTypeChecking(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      string
		wantError string
	}{
		{
			name:      "string minus number",
			code:      `print "a" - 1;`,
			wantError: "requires numbers",
		},
		{
			name:      "add number and string",
			code:      `print 1 + "a";`,
			wantError: "cannot be applied to types",
		},
		{
			name:      "compare bool",
			code:      "print true < false;",
			wantError: "cannot be applied to type bool",
		},
		{
			name:      "indexing a number",
			code:      "var x = 5;\nprint x[0];",
			wantError: "not indexable",
		},
		{
			name:      "wrong arity",
			code:      "func f(a) { return a; }\nprint f(1, 2);",
			wantError: "expects 1 argument(s), got 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errors := analyze(t, tt.code).Errors
			if len(errors) == 0 {
				t.Fatalf("expected a type error containing %q, got none", tt.wantError)
			}
			if !strings.Contains(stringifyErrors(errors), tt.wantError) {
				t.Fatalf("expected error containing %q, got %v", tt.wantError, stringifyErrors(errors))
			}
		})
	}
}

func TestUnusedWarnings(t *testing.T) {
	t.Parallel()

	result := analyze(t, "var used = 1;\nvar dead = 2;\nprint used;")
	if len(result.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", stringifyErrors(result.Errors))
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(result.Warnings), result.Warnings)
	}
	if !strings.Contains(result.Warnings[0], `"dead"`) {
		t.Fatalf("expected warning about 'dead', got %q", result.Warnings[0])
	}
}

func analyze(t *testing.T, code string) Result {
	t.Helper()
	tokens := lexer.NewLexer(code).Tokenize()
	statements := parser.NewParser(tokens).Parse()
	return Analyze(statements)
}

func containsError(errors []error, want string) bool {
	for _, err := range errors {
		if err.Error() == want {
			return true
		}
	}
	return false
}

func stringifyErrors(errors []error) string {
	values := make([]string, 0, len(errors))
	for _, err := range errors {
		values = append(values, err.Error())
	}
	return strings.Join(values, "; ")
}

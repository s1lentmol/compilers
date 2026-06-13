package interpreter

import (
	"compilers/lexer"
	"compilers/optimizer"
	"compilers/parser"
	"strings"
	"testing"
)

func runCode(t *testing.T, code string) string {
	t.Helper()
	tokens := lexer.NewLexer(code).Tokenize()
	statements := parser.NewParser(tokens).Parse()
	statements = optimizer.Optimize(statements)

	var output strings.Builder
	interp := New()
	interp.SetOutput(&output)
	if err := interp.Run(statements); err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	return output.String()
}

func TestArithmeticAndPrint(t *testing.T) {
	got := runCode(t, "print 10 + 20 * 2;")
	if got != "50\n" {
		t.Errorf("got %q, want %q", got, "50\n")
	}
}

func TestControlFlow(t *testing.T) {
	code := "var i = 0;\nvar sum = 0;\nwhile (i < 5) { sum = sum + i; i = i + 1; }\nprint sum;"
	got := runCode(t, code)
	if got != "10\n" {
		t.Errorf("got %q, want %q", got, "10\n")
	}
}

func TestShortCircuit(t *testing.T) {
	got := runCode(t, "if (false && (1 / 0 == 0)) print 1; else print 2;")
	if got != "2\n" {
		t.Errorf("short-circuit failed: got %q", got)
	}
}

func TestRecursionFactorial(t *testing.T) {
	code := "func fact(n) { if (n <= 1) return 1; return n * fact(n - 1); }\nprint fact(5);"
	got := runCode(t, code)
	if got != "120\n" {
		t.Errorf("got %q, want %q", got, "120\n")
	}
}

func TestRecursionFibonacci(t *testing.T) {
	code := "func fib(n) { if (n < 2) return n; return fib(n-1) + fib(n-2); }\nprint fib(10);"
	got := runCode(t, code)
	if got != "55\n" {
		t.Errorf("got %q, want %q", got, "55\n")
	}
}

func TestClosureIsolation(t *testing.T) {
	code := "func f() { var local = 42; return local; }\nprint f();"
	got := runCode(t, code)
	if got != "42\n" {
		t.Errorf("got %q, want %q", got, "42\n")
	}
}

func TestArraysAndIndexing(t *testing.T) {
	code := "var a = [1, 2, 3];\na[1] = 20;\nprint a[1];\nprint len(a);"
	got := runCode(t, code)
	if got != "20\n3\n" {
		t.Errorf("got %q, want %q", got, "20\n3\n")
	}
}

func TestBubbleSort(t *testing.T) {
	code := `var data = [3, 1, 2];
var n = len(data);
var i = 0;
while (i < n - 1) {
  var j = 0;
  while (j < n - 1 - i) {
    if (data[j] > data[j+1]) {
      var t = data[j];
      data[j] = data[j+1];
      data[j+1] = t;
    }
    j = j + 1;
  }
  i = i + 1;
}
print data;`
	got := runCode(t, code)
	if got != "[1, 2, 3]\n" {
		t.Errorf("got %q, want %q", got, "[1, 2, 3]\n")
	}
}

func TestStringConcatenation(t *testing.T) {
	got := runCode(t, `print "foo" + "bar";`)
	if got != "foobar\n" {
		t.Errorf("got %q, want %q", got, "foobar\n")
	}
}

func TestRuntimeErrorDivisionByZero(t *testing.T) {
	tokens := lexer.NewLexer("var z = 0; print 1 / z;").Tokenize()
	statements := parser.NewParser(tokens).Parse()
	interp := New()
	interp.SetOutput(&strings.Builder{})
	err := interp.Run(statements)
	if err == nil {
		t.Fatal("expected division by zero error")
	}
}

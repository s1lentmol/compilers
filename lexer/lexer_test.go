package lexer

import "testing"

func TestTokenizeBasic(t *testing.T) {
	tokens := NewLexer("var x = 12 + 3;").Tokenize()

	want := []TokenType{VAR, ID, EQ, NUMBER, PLUS, NUMBER, SEMICOLON, EOF}
	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(tokens), len(want), tokens)
	}
	for i, wt := range want {
		if tokens[i].Type != wt {
			t.Errorf("token %d: got %s, want %s", i, tokens[i].Type, wt)
		}
	}
}

func TestTokenizeStringAndEscapes(t *testing.T) {
	tokens := NewLexer(`"hello\nworld"`).Tokenize()
	if tokens[0].Type != STRING {
		t.Fatalf("expected STRING, got %s", tokens[0].Type)
	}
	if tokens[0].Value != "hello\nworld" {
		t.Errorf("got %q, want %q", tokens[0].Value, "hello\nworld")
	}
}

func TestTokenizeFloatAndOperators(t *testing.T) {
	tokens := NewLexer("3.14 >= 2 && !true").Tokenize()
	want := []TokenType{NUMBER, GTEQ, NUMBER, AND, EXCL, TRUE, EOF}
	for i, wt := range want {
		if tokens[i].Type != wt {
			t.Errorf("token %d: got %s, want %s", i, tokens[i].Type, wt)
		}
	}
	if tokens[0].Value != "3.14" {
		t.Errorf("float value: got %q", tokens[0].Value)
	}
}

func TestTokenizeSkipsComments(t *testing.T) {
	tokens := NewLexer("1 // line\n/* block */ 2").Tokenize()
	want := []TokenType{NUMBER, NUMBER, EOF}
	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(tokens), len(want), tokens)
	}
}

func TestMaximalMunch(t *testing.T) {
	tokens := NewLexer("a == b").Tokenize()
	if tokens[1].Type != EQEQ {
		t.Errorf("expected EQEQ, got %s", tokens[1].Type)
	}
}

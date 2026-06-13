package lexer

import (
	"fmt"
	"unicode"
)

type Lexer struct {
	input    []rune
	position int
	line     int
	column   int
}

var Keywords = map[string]TokenType{
	"var":   VAR,
	"print": PRINT,
	"if":    IF,
	"else":  ELSE,
	"while": WHILE,
}

var Operators = map[string]TokenType{
	"==": EQEQ,
	"!=": NEQ,
	"<=": LTEQ,
	">=": GTEQ,
	"&&": AND,
	"||": OR,
	"+":  PLUS,
	"-":  MINUS,
	"*":  STAR,
	"/":  SLASH,
	"=":  EQ,
	"<":  LT,
	">":  GT,
	"!":  EXCL,
	"(":  LPAREN,
	")":  RPAREN,
	"{":  LBRACE,
	"}":  RBRACE,
	";":  SEMICOLON,
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:    []rune(input),
		position: 0,
		line:     1,
		column:   1,
	}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token

	for l.position < len(l.input) {
		current := l.peek()

		if unicode.IsSpace(current) {
			l.next()
			continue
		}

		if unicode.IsDigit(current) {
			tokens = append(tokens, l.readNumber())
			continue
		}

		if unicode.IsLetter(current) {
			tokens = append(tokens, l.readWord())
			continue
		}

		tokens = append(tokens, l.readOperatorOrPunctuation())
	}

	tokens = append(tokens, NewToken(EOF, "\000", l.position, l.line, l.column))
	return tokens
}

func (l *Lexer) readNumber() Token {
	startPos := l.position
	startLine := l.line
	startCol := l.column

	for unicode.IsDigit(l.peek()) {
		l.next()
	}

	text := string(l.input[startPos:l.position])
	return NewToken(NUMBER, text, startPos, startLine, startCol)
}

func (l *Lexer) readWord() Token {
	startPos := l.position
	startLine := l.line
	startCol := l.column

	for unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) {
		l.next()
	}

	text := string(l.input[startPos:l.position])
	tokenType, isKeyword := Keywords[text]
	if !isKeyword {
		tokenType = ID
	}

	return NewToken(tokenType, text, startPos, startLine, startCol)
}

func (l *Lexer) readOperatorOrPunctuation() Token {
	startPos := l.position
	startLine := l.line
	startCol := l.column

	// Пробуем считать двухсимвольные операторы
	if l.position+1 < len(l.input) {
		twoChars := string(l.input[l.position : l.position+2])
		if opType, exists := Operators[twoChars]; exists {
			l.next()
			l.next()
			return NewToken(opType, twoChars, startPos, startLine, startCol)
		}
	}

	// Односимвольные операторы
	oneChar := string(l.input[l.position])
	if opType, exists := Operators[oneChar]; exists {
		l.next()
		return NewToken(opType, oneChar, startPos, startLine, startCol)
	}

	badChar := l.peek()
	panic(fmt.Sprintf("[Lexer Error] Unexpected character '%c' at Line %d, Column %d", badChar, startLine, startCol))
}

func (l *Lexer) peek() rune {
	if l.position >= len(l.input) {
		return 0
	}
	return l.input[l.position]
}

func (l *Lexer) next() rune {
	if l.position >= len(l.input) {
		return 0
	}

	current := l.input[l.position]
	l.position++

	if current == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}

	return current
}

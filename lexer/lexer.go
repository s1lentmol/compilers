package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	input    []rune
	position int
	line     int
	column   int
}

var Keywords = map[string]TokenType{
	"var":    VAR,
	"print":  PRINT,
	"if":     IF,
	"else":   ELSE,
	"while":  WHILE,
	"func":   FUNC,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
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
	"%":  PERCENT,
	"=":  EQ,
	"<":  LT,
	">":  GT,
	"!":  EXCL,
	"(":  LPAREN,
	")":  RPAREN,
	"{":  LBRACE,
	"}":  RBRACE,
	"[":  LBRACKET,
	"]":  RBRACKET,
	",":  COMMA,
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

		if current == '/' && l.peekAt(1) == '/' {
			l.skipLineComment()
			continue
		}

		if current == '/' && l.peekAt(1) == '*' {
			l.skipBlockComment()
			continue
		}

		if unicode.IsDigit(current) {
			tokens = append(tokens, l.readNumber())
			continue
		}

		if unicode.IsLetter(current) || current == '_' {
			tokens = append(tokens, l.readWord())
			continue
		}

		if current == '"' {
			tokens = append(tokens, l.readString())
			continue
		}

		tokens = append(tokens, l.readOperatorOrPunctuation())
	}

	tokens = append(tokens, NewToken(EOF, "", l.position, l.line, l.column))
	return tokens
}

func (l *Lexer) readNumber() Token {
	startPos := l.position
	startLine := l.line
	startCol := l.column

	for unicode.IsDigit(l.peek()) {
		l.next()
	}

	if l.peek() == '.' && unicode.IsDigit(l.peekAt(1)) {
		l.next()
		for unicode.IsDigit(l.peek()) {
			l.next()
		}
	}

	text := string(l.input[startPos:l.position])
	return NewToken(NUMBER, text, startPos, startLine, startCol)
}

func (l *Lexer) readWord() Token {
	startPos := l.position
	startLine := l.line
	startCol := l.column

	for unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_' {
		l.next()
	}

	text := string(l.input[startPos:l.position])
	tokenType, isKeyword := Keywords[text]
	if !isKeyword {
		tokenType = ID
	}

	return NewToken(tokenType, text, startPos, startLine, startCol)
}

func (l *Lexer) readString() Token {
	startPos := l.position
	startLine := l.line
	startCol := l.column

	l.next()

	var builder strings.Builder
	for l.peek() != '"' {
		if l.position >= len(l.input) || l.peek() == '\n' {
			panic(fmt.Sprintf("[Lexer Error] Незакрытая строка на Line %d, Column %d", startLine, startCol))
		}

		ch := l.peek()
		if ch == '\\' {
			l.next()
			esc := l.peek()
			switch esc {
			case 'n':
				builder.WriteRune('\n')
			case 't':
				builder.WriteRune('\t')
			case 'r':
				builder.WriteRune('\r')
			case '\\':
				builder.WriteRune('\\')
			case '"':
				builder.WriteRune('"')
			default:
				panic(fmt.Sprintf("[Lexer Error] Неизвестная escape-последовательность '\\%c' на Line %d, Column %d", esc, l.line, l.column))
			}
			l.next()
			continue
		}

		builder.WriteRune(ch)
		l.next()
	}

	l.next()
	return NewToken(STRING, builder.String(), startPos, startLine, startCol)
}

func (l *Lexer) readOperatorOrPunctuation() Token {
	startPos := l.position
	startLine := l.line
	startCol := l.column

	if l.position+1 < len(l.input) {
		twoChars := string(l.input[l.position : l.position+2])
		if opType, exists := Operators[twoChars]; exists {
			l.next()
			l.next()
			return NewToken(opType, twoChars, startPos, startLine, startCol)
		}
	}

	oneChar := string(l.input[l.position])
	if opType, exists := Operators[oneChar]; exists {
		l.next()
		return NewToken(opType, oneChar, startPos, startLine, startCol)
	}

	badChar := l.peek()
	panic(fmt.Sprintf("[Lexer Error] Неожиданный символ '%c' на Line %d, Column %d", badChar, startLine, startCol))
}

func (l *Lexer) skipLineComment() {
	for l.position < len(l.input) && l.peek() != '\n' {
		l.next()
	}
}

func (l *Lexer) skipBlockComment() {
	l.next()
	l.next()
	for l.position < len(l.input) {
		if l.peek() == '*' && l.peekAt(1) == '/' {
			l.next()
			l.next()
			return
		}
		l.next()
	}
}

func (l *Lexer) peek() rune {
	if l.position >= len(l.input) {
		return 0
	}
	return l.input[l.position]
}

func (l *Lexer) peekAt(offset int) rune {
	idx := l.position + offset
	if idx >= len(l.input) {
		return 0
	}
	return l.input[idx]
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

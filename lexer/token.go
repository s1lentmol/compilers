package lexer

import "fmt"

type Token struct {
	Type     TokenType
	Value    string
	Position int
	Line     int
	Column   int
}

func NewToken(tokenType TokenType, value string, position, line, column int) Token {
	return Token{
		Type:     tokenType,
		Value:    value,
		Position: position,
		Line:     line,
		Column:   column,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("[%d:%d] Token(%s, '%s')", t.Line, t.Column, t.Type, t.Value)
}

package lexer

type TokenType int

const (
	NUMBER TokenType = iota
	ID
	STRING
	TRUE
	FALSE

	VAR
	PRINT
	IF
	ELSE
	WHILE
	FUNC
	RETURN

	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	EQ
	EQEQ
	EXCL
	NEQ
	LT
	GT
	LTEQ
	GTEQ
	AND
	OR

	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACKET
	RBRACKET
	COMMA
	SEMICOLON

	EOF
)

func (t TokenType) String() string {
	return [...]string{
		"NUMBER", "ID", "STRING", "TRUE", "FALSE",
		"VAR", "PRINT", "IF", "ELSE", "WHILE", "FUNC", "RETURN",
		"PLUS", "MINUS", "STAR", "SLASH", "PERCENT",
		"EQ", "EQEQ", "EXCL", "NEQ",
		"LT", "GT", "LTEQ", "GTEQ",
		"AND", "OR",
		"LPAREN", "RPAREN", "LBRACE", "RBRACE", "LBRACKET", "RBRACKET",
		"COMMA", "SEMICOLON", "EOF",
	}[t]
}

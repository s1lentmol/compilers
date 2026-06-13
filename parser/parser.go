package parser

import (
	"compilers/lexer"
	"fmt"
	"strconv"
)

type Parser struct {
	tokens   []lexer.Token
	position int
}

func tokenPosition(token lexer.Token) Position {
	return Position{Line: token.Line, Column: token.Column}
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:   tokens,
		position: 0,
	}
}

func (p *Parser) Parse() []Statement {
	var statements []Statement
	for !p.isAtEnd() {
		statements = append(statements, p.parseDeclaration())
	}
	return statements
}

func (p *Parser) parseDeclaration() Statement {
	if p.match(lexer.VAR) {
		return p.parseVarDeclaration()
	}
	return p.parseStatement()
}

func (p *Parser) parseStatement() Statement {
	if p.match(lexer.IF) {
		return p.parseIfStatement()
	}
	if p.match(lexer.WHILE) {
		return p.parseWhileStatement()
	}
	if p.match(lexer.PRINT) {
		return p.parsePrintStatement()
	}
	if p.match(lexer.LBRACE) {
		return &BlockStatement{
			Position:   tokenPosition(p.previous()),
			Statements: p.parseBlock(),
		}
	}
	return p.parseExpressionStatement()
}

func (p *Parser) parseVarDeclaration() Statement {
	varToken := p.previous()
	name := p.consume(lexer.ID, "Ожидается имя переменной.")
	var initializer Expression

	if p.match(lexer.EQ) {
		initializer = p.parseExpression()
	}

	p.consume(lexer.SEMICOLON, "Ожидается ';' после объявления переменной.")
	return &VarStatement{
		Position:    tokenPosition(varToken),
		Name:        name.Value,
		Initializer: initializer,
	}
}

func (p *Parser) parseIfStatement() Statement {
	ifToken := p.previous()
	p.consume(lexer.LPAREN, "Ожидается '(' после 'if'.")
	condition := p.parseExpression()
	p.consume(lexer.RPAREN, "Ожидается ')' после условия 'if'.")

	thenBranch := p.parseStatement()
	var elseBranch Statement

	if p.match(lexer.ELSE) {
		elseBranch = p.parseStatement()
	}

	return &IfStatement{
		Position:   tokenPosition(ifToken),
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

func (p *Parser) parseWhileStatement() Statement {
	whileToken := p.previous()
	p.consume(lexer.LPAREN, "Ожидается '(' после 'while'.")
	condition := p.parseExpression()
	p.consume(lexer.RPAREN, "Ожидается ')' после условия 'while'.")

	body := p.parseStatement()
	return &WhileStatement{
		Position:  tokenPosition(whileToken),
		Condition: condition,
		Body:      body,
	}
}

func (p *Parser) parsePrintStatement() Statement {
	printToken := p.previous()
	value := p.parseExpression()
	p.consume(lexer.SEMICOLON, "Ожидается ';' после значения.")
	return &PrintStatement{
		Position: tokenPosition(printToken),
		Value:    value,
	}
}

func (p *Parser) parseExpressionStatement() Statement {
	expr := p.parseExpression()
	position := expressionPosition(expr)
	p.consume(lexer.SEMICOLON, "Ожидается ';' после выражения.")
	return &ExpressionStatement{
		Position:   position,
		Expression: expr,
	}
}

func (p *Parser) parseBlock() []Statement {
	var statements []Statement
	for !p.check(lexer.RBRACE) && !p.isAtEnd() {
		statements = append(statements, p.parseDeclaration())
	}
	p.consume(lexer.RBRACE, "Ожидается '}' после блока.")
	return statements
}

func (p *Parser) parseExpression() Expression {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() Expression {
	expr := p.parseLogicalOr()

	if p.match(lexer.EQ) {
		equals := p.previous()
		value := p.parseAssignment()

		if varExpr, ok := expr.(*VariableExpression); ok {
			return &AssignExpression{
				Position: varExpr.Position,
				Name:     varExpr.Name,
				Value:    value,
			}
		}

		panic(fmt.Sprintf("[Parser Error] Line %d: Недопустимая цель для присваивания.", equals.Line))
	}

	return expr
}

func (p *Parser) parseLogicalOr() Expression {
	expr := p.parseLogicalAnd()
	for p.match(lexer.OR) {
		operatorToken := p.previous()
		right := p.parseLogicalAnd()
		expr = &BinaryExpression{
			Position: tokenPosition(operatorToken),
			Left:     expr,
			Operator: operatorToken.Value,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) parseLogicalAnd() Expression {
	expr := p.parseEquality()
	for p.match(lexer.AND) {
		operatorToken := p.previous()
		right := p.parseEquality()
		expr = &BinaryExpression{
			Position: tokenPosition(operatorToken),
			Left:     expr,
			Operator: operatorToken.Value,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) parseEquality() Expression {
	expr := p.parseComparison()
	for p.match(lexer.EQEQ, lexer.NEQ) {
		operatorToken := p.previous()
		right := p.parseComparison()
		expr = &BinaryExpression{
			Position: tokenPosition(operatorToken),
			Left:     expr,
			Operator: operatorToken.Value,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) parseComparison() Expression {
	expr := p.parseTerm()
	for p.match(lexer.LT, lexer.LTEQ, lexer.GT, lexer.GTEQ) {
		operatorToken := p.previous()
		right := p.parseTerm()
		expr = &BinaryExpression{
			Position: tokenPosition(operatorToken),
			Left:     expr,
			Operator: operatorToken.Value,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) parseTerm() Expression {
	expr := p.parseFactor()
	for p.match(lexer.PLUS, lexer.MINUS) {
		operatorToken := p.previous()
		right := p.parseFactor()
		expr = &BinaryExpression{
			Position: tokenPosition(operatorToken),
			Left:     expr,
			Operator: operatorToken.Value,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) parseFactor() Expression {
	expr := p.parseUnary()
	for p.match(lexer.STAR, lexer.SLASH) {
		operatorToken := p.previous()
		right := p.parseUnary()
		expr = &BinaryExpression{
			Position: tokenPosition(operatorToken),
			Left:     expr,
			Operator: operatorToken.Value,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) parseUnary() Expression {
	if p.match(lexer.EXCL, lexer.MINUS) {
		operatorToken := p.previous()
		right := p.parseUnary()
		return &UnaryExpression{
			Position: tokenPosition(operatorToken),
			Operator: operatorToken.Value,
			Right:    right,
		}
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() Expression {
	if p.match(lexer.NUMBER) {
		numberToken := p.previous()
		val, _ := strconv.ParseFloat(numberToken.Value, 64)
		return &NumberExpression{
			Position: tokenPosition(numberToken),
			Value:    val,
		}
	}

	if p.match(lexer.ID) {
		idToken := p.previous()
		return &VariableExpression{
			Position: tokenPosition(idToken),
			Name:     idToken.Value,
		}
	}

	if p.match(lexer.LPAREN) {
		expr := p.parseExpression()
		p.consume(lexer.RPAREN, "Ожидается ')' после выражения.")
		return expr
	}

	panic(fmt.Sprintf("[Parser Error] Line %d, Col %d: Ожидается выражение.", p.peek().Line, p.peek().Column))
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(tokenType lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.position++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == lexer.EOF
}

func (p *Parser) peek() lexer.Token {
	return p.tokens[p.position]
}

func (p *Parser) previous() lexer.Token {
	return p.tokens[p.position-1]
}

func (p *Parser) consume(tokenType lexer.TokenType, message string) lexer.Token {
	if p.check(tokenType) {
		return p.advance()
	}
	panic(fmt.Sprintf("[Parser Error] Line %d, Col %d: %s", p.peek().Line, p.peek().Column, message))
}

func expressionPosition(expr Expression) Position {
	switch value := expr.(type) {
	case *NumberExpression:
		return value.Position
	case *VariableExpression:
		return value.Position
	case *BinaryExpression:
		return value.Position
	case *UnaryExpression:
		return value.Position
	case *AssignExpression:
		return value.Position
	default:
		return Position{}
	}
}

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
	if p.match(lexer.FUNC) {
		return p.parseFunctionDeclaration()
	}
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
	if p.match(lexer.RETURN) {
		return p.parseReturnStatement()
	}
	if p.match(lexer.LBRACE) {
		return p.parseBlockStatement()
	}
	return p.parseExpressionStatement()
}

func (p *Parser) parseFunctionDeclaration() Statement {
	funcToken := p.previous()
	name := p.consume(lexer.ID, "Ожидается имя функции.")

	p.consume(lexer.LPAREN, "Ожидается '(' после имени функции.")
	var parameters []string
	if !p.check(lexer.RPAREN) {
		for {
			param := p.consume(lexer.ID, "Ожидается имя параметра.")
			parameters = append(parameters, param.Value)
			if !p.match(lexer.COMMA) {
				break
			}
		}
	}
	p.consume(lexer.RPAREN, "Ожидается ')' после параметров.")

	p.consume(lexer.LBRACE, "Ожидается '{' перед телом функции.")
	body := p.parseBlockStatement().(*BlockStatement)

	return &FunctionStatement{
		Position:   tokenPosition(funcToken),
		Name:       name.Value,
		Parameters: parameters,
		Body:       body,
	}
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

func (p *Parser) parseReturnStatement() Statement {
	returnToken := p.previous()
	var value Expression
	if !p.check(lexer.SEMICOLON) {
		value = p.parseExpression()
	}
	p.consume(lexer.SEMICOLON, "Ожидается ';' после return.")
	return &ReturnStatement{
		Position: tokenPosition(returnToken),
		Value:    value,
	}
}

func (p *Parser) parseExpressionStatement() Statement {
	start := p.peek()
	expr := p.parseExpression()
	p.consume(lexer.SEMICOLON, "Ожидается ';' после выражения.")
	return &ExpressionStatement{
		Position:   tokenPosition(start),
		Expression: expr,
	}
}

func (p *Parser) parseBlockStatement() Statement {
	braceToken := p.previous()
	var statements []Statement
	for !p.check(lexer.RBRACE) && !p.isAtEnd() {
		statements = append(statements, p.parseDeclaration())
	}
	p.consume(lexer.RBRACE, "Ожидается '}' после блока.")
	return &BlockStatement{
		Position:   tokenPosition(braceToken),
		Statements: statements,
	}
}

func (p *Parser) parseExpression() Expression {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() Expression {
	expr := p.parseLogicalOr()

	if p.match(lexer.EQ) {
		equals := p.previous()
		value := p.parseAssignment()

		switch target := expr.(type) {
		case *VariableExpression:
			return &AssignExpression{
				Position: target.Position,
				Name:     target.Name,
				Value:    value,
			}
		case *IndexExpression:
			return &IndexAssignExpression{
				Position:   target.Position,
				Collection: target.Collection,
				Index:      target.Index,
				Value:      value,
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
	for p.match(lexer.STAR, lexer.SLASH, lexer.PERCENT) {
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
	return p.parseCall()
}

func (p *Parser) parseCall() Expression {
	expr := p.parsePrimary()

	for {
		if p.match(lexer.LPAREN) {
			expr = p.finishCall(expr)
		} else if p.match(lexer.LBRACKET) {
			bracket := p.previous()
			index := p.parseExpression()
			p.consume(lexer.RBRACKET, "Ожидается ']' после индекса.")
			expr = &IndexExpression{
				Position:   tokenPosition(bracket),
				Collection: expr,
				Index:      index,
			}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) finishCall(callee Expression) Expression {
	paren := p.previous()
	var arguments []Expression
	if !p.check(lexer.RPAREN) {
		for {
			arguments = append(arguments, p.parseExpression())
			if !p.match(lexer.COMMA) {
				break
			}
		}
	}
	p.consume(lexer.RPAREN, "Ожидается ')' после аргументов.")
	return &CallExpression{
		Position:  tokenPosition(paren),
		Callee:    callee,
		Arguments: arguments,
	}
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

	if p.match(lexer.STRING) {
		stringToken := p.previous()
		return &StringExpression{
			Position: tokenPosition(stringToken),
			Value:    stringToken.Value,
		}
	}

	if p.match(lexer.TRUE) {
		return &BooleanExpression{Position: tokenPosition(p.previous()), Value: true}
	}

	if p.match(lexer.FALSE) {
		return &BooleanExpression{Position: tokenPosition(p.previous()), Value: false}
	}

	if p.match(lexer.ID) {
		idToken := p.previous()
		return &VariableExpression{
			Position: tokenPosition(idToken),
			Name:     idToken.Value,
		}
	}

	if p.match(lexer.LBRACKET) {
		return p.parseArrayLiteral()
	}

	if p.match(lexer.LPAREN) {
		expr := p.parseExpression()
		p.consume(lexer.RPAREN, "Ожидается ')' после выражения.")
		return expr
	}

	panic(fmt.Sprintf("[Parser Error] Line %d, Col %d: Ожидается выражение.", p.peek().Line, p.peek().Column))
}

func (p *Parser) parseArrayLiteral() Expression {
	bracket := p.previous()
	var elements []Expression
	if !p.check(lexer.RBRACKET) {
		for {
			elements = append(elements, p.parseExpression())
			if !p.match(lexer.COMMA) {
				break
			}
		}
	}
	p.consume(lexer.RBRACKET, "Ожидается ']' после элементов массива.")
	return &ArrayExpression{
		Position: tokenPosition(bracket),
		Elements: elements,
	}
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

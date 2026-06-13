package parser

type Position struct {
	Line   int
	Column int
}

// Node - base interface for all AST nodes
type Node interface{}

// Expression - interface for all expression nodes
type Expression interface {
	Node
}

// Statement - interface for all statement nodes
type Statement interface {
	Node
}

// --- Expressions ---

type NumberExpression struct {
	Position Position
	Value    float64
}

type VariableExpression struct {
	Position Position
	Name     string
}

type BinaryExpression struct {
	Position Position
	Left     Expression
	Operator string
	Right    Expression
}

type UnaryExpression struct {
	Position Position
	Operator string
	Right    Expression
}

type AssignExpression struct {
	Position Position
	Name     string
	Value    Expression
}

// --- Statements ---

type VarStatement struct {
	Position    Position
	Name        string
	Initializer Expression
}

type IfStatement struct {
	Position   Position
	Condition  Expression
	ThenBranch Statement
	ElseBranch Statement
}

type WhileStatement struct {
	Position  Position
	Condition Expression
	Body      Statement
}

type PrintStatement struct {
	Position Position
	Value    Expression
}

type BlockStatement struct {
	Position   Position
	Statements []Statement
}

type ExpressionStatement struct {
	Position   Position
	Expression Expression
}

package parser

type Position struct {
	Line   int
	Column int
}

type Node interface{}

type Expression interface{}

type Statement interface{}

type NumberExpression struct {
	Position Position
	Value    float64
}

type StringExpression struct {
	Position Position
	Value    string
}

type BooleanExpression struct {
	Position Position
	Value    bool
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

type CallExpression struct {
	Position  Position
	Callee    Expression
	Arguments []Expression
}

type ArrayExpression struct {
	Position Position
	Elements []Expression
}

type IndexExpression struct {
	Position   Position
	Collection Expression
	Index      Expression
}

type IndexAssignExpression struct {
	Position   Position
	Collection Expression
	Index      Expression
	Value      Expression
}

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

type FunctionStatement struct {
	Position   Position
	Name       string
	Parameters []string
	Body       *BlockStatement
}

type ReturnStatement struct {
	Position Position
	Value    Expression
}

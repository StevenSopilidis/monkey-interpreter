package ast

import (
	"bytes"

	"github.com/stevensopilidis/monkey/token"
)

// represents node in an AST (can be either Statement or Expression)
type Node interface {
	TokenLiteral() string
	String() string
}

// represents statement node
type Statement interface {
	Node
	statementNode()
}

// represents expression node
type Expression interface {
	Node
	expressionNode()
}

// struct representing a monkey program
// each program is a series of statements and the
// Statements attribute represents all roots of the ASTs created
type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// struct that represenst a Bool (Expression)
type Boolean struct {
	Token token.Token // token.TRUE || token.FALSE
	Value bool
}

func (b Boolean) String() string       { return b.Token.Literal }
func (b Boolean) expressionNode()      {}
func (b Boolean) TokenLiteral() string { return b.Token.Literal }

// struct that represents an identifier (Expression)
type Identifier struct {
	Token token.Token // token.IDENT
	Value string      // identifier name
}

func (i Identifier) String() string { return i.Value }

// satisfy Node interface
func (i Identifier) expressionNode()      {}
func (i Identifier) TokenLiteral() string { return i.Token.Literal }

// struct that represents infix Expression (<expression><operator><expression>)
type InfixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
	Left     Expression
}

func (ie InfixExpression) expressionNode()      {}
func (ie InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// struct that represents prefix Expressions (<prefix_operator><expression>)
type PrefixExpression struct {
	Token    token.Token // ! or -
	Operator string
	Right    Expression
}

func (pe PrefixExpression) expressionNode()      {}
func (pe PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// structst that represent a Float literal (expression)
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl FloatLiteral) expressionNode()      {}
func (fl FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl FloatLiteral) String() string       { return fl.Token.Literal }

// struct that represents a Integer literal (expression)
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il IntegerLiteral) expressionNode()      {}
func (il IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il IntegerLiteral) String() string       { return il.Token.Literal }

// sturct representing a let statement (Statement)
type LetStatement struct {
	Token token.Token // token.Let token
	Name  Identifier  // name of variable
	Value Expression  // expression that produces the value
}

func (ls LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")
	return out.String()
}

// satisfy Node interface
func (ls LetStatement) statementNode()       {}
func (ls LetStatement) TokenLiteral() string { return ls.Token.Literal }

// struct representing a return statement (Statement)
type ReturnStatement struct {
	Token       token.Token // token.RETURN token
	ReturnValue Expression
}

func (rs ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

func (rs ReturnStatement) statementNode()       {}
func (rs ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// struct that represents Expression Statements (so it acts as a wrapper for lines
// that contain only an expression)

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

func (es ExpressionStatement) statementNode()       {}
func (es ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

package parser

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stevensopilidis/monkey/ast"
	"github.com/stevensopilidis/monkey/lexer"
	"github.com/stretchr/testify/require"
)

func TestLetStatements(t *testing.T) {
	input := `
		let x = 5.34;
		let y = 10;
		let foobar = 838383;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	require.NotNil(t, program)
	require.Equal(t, 3, len(program.Statements))

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		testLetStatement(t, stmt, tt.expectedIdentifier)
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) {
	require.Equal(t, s.TokenLiteral(), "let")

	letStmt, ok := s.(ast.LetStatement)
	require.True(t, ok)

	require.Equal(t, name, letStmt.Name.Value)
	require.Equal(t, name, letStmt.Name.TokenLiteral())
}

func TestReturnStatements(t *testing.T) {
	input := `
		return 5;
		return 10.5;
		return 90234123;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	require.NotNil(t, program)
	require.Equal(t, 3, len(program.Statements))

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		require.True(t, ok)
		require.Equal(t, "return", returnStmt.TokenLiteral())
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Equal(t, 1, len(program.Statements))

	stmt, ok := program.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	ident, ok := stmt.Expression.(ast.Identifier)
	require.True(t, ok)

	require.Equal(t, "foobar", ident.Value)
	require.Equal(t, "foobar", ident.TokenLiteral())
}

func TestIntegerExpressions(t *testing.T) {
	input := "5;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Equal(t, 1, len(program.Statements))

	stmt, ok := program.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	literal, ok := stmt.Expression.(ast.IntegerLiteral)
	require.True(t, ok)
	require.Equal(t, "5", literal.TokenLiteral())
}

func TestBooleanExpressions(t *testing.T) {
	tests := []struct {
		input           string
		expectedBoolean bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		require.Equal(t, 1, len(program.Statements))
		exp, ok := program.Statements[0].(ast.ExpressionStatement)
		require.True(t, ok)

		boolean, ok := exp.Expression.(ast.Boolean)
		require.True(t, ok)
		require.Equal(t, tc.expectedBoolean, boolean.Value)
	}
}

func TestFloatExpressions(t *testing.T) {
	input := "5234.23234413;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Equal(t, 1, len(program.Statements))

	stmt, ok := program.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	literal, ok := stmt.Expression.(ast.FloatLiteral)
	require.True(t, ok)
	require.Equal(t, "5234.23234413", literal.TokenLiteral())
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!foobar;", "!", "foobar"},
		{"-foobar;", "-", "foobar"},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}
	for _, tc := range prefixTests {
		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		require.Equal(t, 1, len(program.Statements))
		stmt, ok := program.Statements[0].(ast.ExpressionStatement)
		require.True(t, ok)
		exp, ok := stmt.Expression.(ast.PrefixExpression)
		require.True(t, ok)

		require.Equal(t, tc.operator, exp.Operator)
		testLiteralExpression(t, exp.Right, tc.value)
	}
}

// helper function  function for checking that the expression is the correct identifier
func testIdentifier(t *testing.T, exp ast.Expression, value string) {
	ident, ok := exp.(ast.Identifier)

	require.True(t, ok)
	require.Equal(t, value, ident.Value)
	require.Equal(t, ident.TokenLiteral(), value)
}

// function for testing a literal expression
func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) {
	switch v := expected.(type) {
	case int:
		testIntOrFloatLiteral(t, exp, strconv.Itoa(v))
	case int64:
		testIntOrFloatLiteral(t, exp, strconv.FormatInt(v, 10))
	case float64:
		testIntOrFloatLiteral(t, exp, strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		testBooleanLiterals(t, exp, v)
	case string:
		testIdentifier(t, exp, v)
	}
}

// helper function for testing boolean literals
func testBooleanLiterals(t *testing.T, exp ast.Expression, value bool) {
	bo, ok := exp.(ast.Boolean)
	require.True(t, ok)
	require.Equal(t, value, bo.Value)
	require.Equal(t, bo.TokenLiteral(), fmt.Sprintf("%t", value))
}

// helper function for testing infix expressions
func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) {
	ifExp, ok := exp.(ast.InfixExpression)
	require.True(t, ok)
	testLiteralExpression(t, ifExp.Left, left)
	require.Equal(t, operator, ifExp.Operator)
	testLiteralExpression(t, ifExp.Right, right)
}

// helper function for checking that literals are either IntegerLiterals or FlotLiterals
func testIntOrFloatLiteral(t *testing.T, exp ast.Expression, value string) {
	il, ok := exp.(ast.IntegerLiteral)
	if !ok {
		// FloatLiteral
		fl, ok := exp.(ast.FloatLiteral)
		require.True(t, ok)
		val, err := strconv.ParseFloat(value, 64)
		require.NoError(t, err)
		require.Equal(t, val, fl.Value)
		return
	}
	// IntegerLiteral
	require.True(t, ok)
	val, err := strconv.ParseInt(value, 0, 64)
	require.NoError(t, err)
	require.Equal(t, val, il.Value)
}

// function for testing infix operator parsing
func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"false == false", false, "==", false},
		{"blah == blah", "blah", "==", "blah"},
		{"3432.234234 > 45.234234", 3432.234234, ">", 45.234234},
	}

	for _, tc := range infixTests {
		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		require.Equal(t, 1, len(program.Statements))
		stmt, ok := program.Statements[0].(ast.ExpressionStatement)
		require.True(t, ok)

		exp, ok := stmt.Expression.(ast.InfixExpression)
		require.True(t, ok)

		testInfixExpression(t, exp, tc.leftValue, tc.operator, tc.rightValue)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4.23423423 * 5.234234 == 3 * 1 + 4234.23234 * 5.5324234",
			"((3 + (4.23423423 * 5.234234)) == ((3 * 1) + (4234.23234 * 5.5324234)))",
		},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		require.Equal(t, actual, tc.expected)
	}
}

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
	testCases := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
		{"let fl = 32.24213;", "fl", 32.24213},
	}

	for _, tc := range testCases {
		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		require.Equal(t, 1, len(program.Statements))

		stmt := program.Statements[0]
		testLetStatement(t, stmt, tc.expectedIdentifier)

		val := stmt.(ast.LetStatement).Value
		testLiteralExpression(t, val, tc.expectedValue)
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
	testsCases := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return true;", true},
		{"return foobar;", "foobar"},
	}

	for _, tc := range testsCases {
		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		require.Equal(t, 1, len(program.Statements))

		returnStmt, ok := program.Statements[0].(ast.ReturnStatement)
		require.True(t, ok)
		require.Equal(t, "return", returnStmt.TokenLiteral())
		testLiteralExpression(t, returnStmt.ReturnValue, tc.expectedValue)
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
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
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

// function for testing parsing on if expressions
func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Equal(t, 1, len(program.Statements))
	stmt, ok := program.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	exp, ok := stmt.Expression.(ast.IfExpression)
	require.True(t, ok)

	testInfixExpression(t, exp.Condition, "x", "<", "y")

	// test consequence
	require.Equal(t, 1, len(exp.Consequence.Statements))
	consequence, ok := exp.Consequence.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)
	testIdentifier(t, consequence.Expression, "x")

	// test alternative
	require.Nil(t, exp.Alternative)
}

// function for testing parsing on if-else expressions
func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else {y}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Equal(t, 1, len(program.Statements))
	stmt, ok := program.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	exp, ok := stmt.Expression.(ast.IfExpression)
	require.True(t, ok)

	testInfixExpression(t, exp.Condition, "x", "<", "y")

	// test consequence
	require.Equal(t, 1, len(exp.Consequence.Statements))
	consequence, ok := exp.Consequence.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	testIdentifier(t, consequence.Expression, "x")
	// test alternative
	require.Equal(t, 1, len(exp.Consequence.Statements))
	alternative, ok := exp.Alternative.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	testIdentifier(t, alternative.Expression, "y")
}

// function for testing the parsing of the function's parameters
func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tc := range tests {
		l := lexer.New(tc.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt, ok := program.Statements[0].(ast.ExpressionStatement)
		require.True(t, ok)
		function, ok := stmt.Expression.(ast.FunctionLiteral)
		require.True(t, ok)

		require.Equal(t, len(tc.expectedParams), len(function.Parameters))

		for i, ident := range tc.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}

// function for testing the parsing of functions
func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	require.Equal(t, 1, len(program.Statements))

	stmt, ok := program.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	function, ok := stmt.Expression.(ast.FunctionLiteral)
	require.True(t, ok)

	require.Equal(t, 2, len(function.Parameters))
	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	require.Equal(t, 1, len(function.Body.Statements))

	bodyStmt, ok := function.Body.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

// function for testing call expressions
func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	require.Equal(t, 1, len(program.Statements))

	stmt, ok := program.Statements[0].(ast.ExpressionStatement)
	require.True(t, ok)

	exp, ok := stmt.Expression.(ast.CallExpression)
	require.True(t, ok)

	testIdentifier(t, exp.Function, "add")

	require.Equal(t, 3, len(exp.Arguments))

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

// function for testing the parsing of arguments in a call
func TestCallExpressionParameterParsing(t *testing.T) {
	tests := []struct {
		input         string
		expectedIdent string
		expectedArgs  []string
	}{
		{
			input:         "add();",
			expectedIdent: "add",
			expectedArgs:  []string{},
		},
		{
			input:         "add(1);",
			expectedIdent: "add",
			expectedArgs:  []string{"1"},
		},
		{
			input:         "add(1, 2 * 3, 4 + 5);",
			expectedIdent: "add",
			expectedArgs:  []string{"1", "(2 * 3)", "(4 + 5)"},
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(ast.ExpressionStatement)
		exp, ok := stmt.Expression.(ast.CallExpression)
		require.True(t, ok)

		testIdentifier(t, exp.Function, tt.expectedIdent)

		require.Equal(t, len(tt.expectedArgs), len(exp.Arguments))

		for i, arg := range tt.expectedArgs {
			require.Equal(t, exp.Arguments[i].String(), arg)
		}
	}
}

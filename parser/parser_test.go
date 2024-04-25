package parser

import (
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
		value    string
	}{
		{"!5.324;", "!", "5.324"},
		{"!5;", "!", "5"},
		{"-15;", "-", "15"},
		{"-5.324;", "-", "5.324"},
	}
	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		require.Equal(t, 1, len(program.Statements))
		stmt, ok := program.Statements[0].(ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}
		exp, ok := stmt.Expression.(ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}
		testLiteral(t, exp.Right, tt.value)
	}
}

// for checking that literals are either IntegerLiterals or FlotLiterals
func testLiteral(t *testing.T, exp ast.Expression, value string) {
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
		leftValue  string
		operator   string
		rightValue string
	}{
		{"5 + 5;", "5", "+", "5"},
		{"5 - 5;", "5", "-", "5"},
		{"5 * 5;", "5", "*", "5"},
		{"5 / 5;", "5", "/", "5"},
		{"5 > 5;", "5", ">", "5"},
		{"5 < 5;", "5", "<", "5"},
		{"5 == 5;", "5", "==", "5"},
		{"5 != 5;", "5", "!=", "5"},
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

		testLiteral(t, exp.Left, tc.leftValue)

		require.Equal(t, exp.Operator, tc.operator)

		testLiteral(t, exp.Right, tc.rightValue)
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

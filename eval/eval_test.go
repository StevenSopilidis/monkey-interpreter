package eval

import (
	"fmt"
	"testing"

	"github.com/stevensopilidis/monkey/lexer"
	"github.com/stevensopilidis/monkey/object"
	"github.com/stevensopilidis/monkey/parser"
	"github.com/stretchr/testify/require"
)

func TestLetStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tc := range testCases {
		testIntegerObject(t, testEval(tc.input), tc.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	testCases := []struct {
		input                string
		expectedErrorMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
			if (10 > 1) {
				if (10 > 1) {
					return true + false;
				}
				return 1;
			}
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
	}

	for _, tc := range testCases {
		evaluated := testEval(tc.input)

		errObj, ok := evaluated.(*object.Error)
		require.True(t, ok)

		require.Equal(t, tc.expectedErrorMessage, errObj.Message)
	}
}

func TestReturnStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`
			if (10 > 1) {
				if (10 > 1) {
					return 10;
				}
				return 1;
			}`, 10},
	}

	for _, tc := range testCases {
		evaluated := testEval(tc.input)
		testIntegerObject(t, evaluated, int64(tc.expected))
	}
}

func TestIfElseExpressions(t *testing.T) {
	testCases := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tc := range testCases {
		evaluated := testEval(tc.input)
		expectedValue, ok := tc.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(expectedValue))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) {
	require.Equal(t, obj, NULL)
}

func TestBangOperator(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tc := range testCases {
		evaluated := testEval(tc.input)
		testBooleanObject(t, evaluated, tc.expected)
	}
}

func TestEvalIntegerExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"(10 + 5) * 2", 30},
	}

	for _, tc := range testCases {
		evaluated := testEval(tc.input)
		testIntegerObject(t, evaluated, tc.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1.342 < 2.23423", true},
		{"1.2341 > 2.234", false},
		{"1.21 < 1.21", false},
		{"1.54 > 1.54", false},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"(1 > 2) == (1 > 2)", true},

		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1.325 == 1.325", true},
		{"1.325 != 1.325", false},
		{"1.325 == 2.325", false},
		{"1.325 != 2.325", true},
	}
	for _, tc := range testCases {
		evaluated := testEval(tc.input)
		testBooleanObject(t, evaluated, tc.expected)
	}
}

func TestEvalFloatExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
	}{
		{"23.234234", 23.234234},
		{"5.324", 5.324},
		{"-23.234234", -23.234234},
		{"-5.324", -5.324},
		{"-5.234 + 23.3432", 18.1092},
		{"34.23 + 35.1242 - 324.234132", -254.879932},
		{"34.23 * 6", 205.38},
		{"6 * 34.23", 205.38},
		{"7.21 - 10.42 + (20.28 - 34.28) * 2", -31.21},
		{"7.2 - 0.2 + 1.2 * 2", 9.4},
	}

	for _, tc := range testCases {
		evaluated := testEval(tc.input)
		testFloatObject(t, evaluated, tc.expected)
	}
}

// function for testing Float objects
func testFloatObject(t *testing.T, obj object.Object, expected float64) {
	result, ok := obj.(*object.Float)
	require.True(t, ok)

	require.Equal(t, expected, result.Value)
}

// function for testing boolean objects
func testBooleanObject(t *testing.T, obj object.Object, expected bool) {
	result, ok := obj.(*object.Boolean)
	require.True(t, ok)

	require.Equal(t, expected, result.Value)
}

// helper function for running eval
func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

// function for testing the value of Integer Object
func testIntegerObject(t *testing.T, obj object.Object, expected int64) {
	if obj == nil {
		fmt.Println("NILLL")
	}
	result, ok := obj.(*object.Integer)
	require.True(t, ok)

	require.Equal(t, expected, result.Value)
}

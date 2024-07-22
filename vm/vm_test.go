package vm

import (
	"testing"

	"github.com/stevensopilidis/monkey/ast"
	"github.com/stevensopilidis/monkey/compiler"
	"github.com/stevensopilidis/monkey/lexer"
	"github.com/stevensopilidis/monkey/object"
	"github.com/stevensopilidis/monkey/parser"
	"github.com/stretchr/testify/require"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, testCases []vmTestCase) {
	t.Helper()

	for _, tc := range testCases {
		program := parse(tc.input)

		comp := compiler.New()
		err := comp.Compile(program)
		require.NoError(t, err)

		vm := New(comp.Bytecode())
		err = vm.Run()
		require.NoError(t, err)

		stackElem := vm.LastPoppedStackElement()
		testExpectedObject(t, tc.expected, stackElem)
	}
}

func testExpectedObject(
	t *testing.T,
	expected interface{},
	actual object.Object,
) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		testIntegerObject(t, int64(expected), actual)
	case bool:
		testBooleanObject(t, bool(expected), actual)
	case *object.Null:
		require.Equal(t, actual, Null)
	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Hash)

		require.True(t, ok)
		require.Equal(t, len(expected), len(hash.Pairs))

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			require.True(t, ok)

			testExpectedObject(t, expectedValue, pair.Value)
		}

	case string:
		testStringObject(t, string(expected), actual)
	case []int:
		array, ok := actual.(*object.Array)

		require.True(t, ok)
		require.Equal(t, len(expected), len(array.Elements))
		for i, expectedEl := range expected {
			testIntegerObject(t, int64(expectedEl), array.Elements[i])
		}
	}
}

func testStringObject(t *testing.T, expected string, actual object.Object) {
	result, ok := actual.(*object.String)

	require.True(t, ok)
	require.Equal(t, expected, result.Value)
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testBooleanObject(t *testing.T, expected bool, actual object.Object) {
	result, ok := actual.(*object.Boolean)
	require.True(t, ok)

	require.Equal(t, expected, result.Value)
}

func testIntegerObject(t *testing.T, expected int64, actual object.Object) {
	result, ok := actual.(*object.Integer)

	require.True(t, ok)

	require.Equal(t, expected, result.Value)
}

func TestIntegerArithmetic(t *testing.T) {
	testCases := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(t, testCases)
}

func TestBooleanExpressions(t *testing.T) {
	testCases := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!(if (false) { 5; })", true},
	}

	runVmTests(t, testCases)
}

func TestStringExpressions(t *testing.T) {
	testCases := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}

	runVmTests(t, testCases)
}

func TestArrayLiterals(t *testing.T) {
	testCases := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}

	runVmTests(t, testCases)
}

func TestHashLiterals(t *testing.T) {
	testCases := []vmTestCase{
		{
			"{}", map[object.HashKey]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVmTests(t, testCases)
}

func TestIndexExpressions(t *testing.T) {
	testCases := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", Null},
		{"[1, 2, 3][99]", Null},
		{"[1][-1]", Null},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", Null},
		{"{}[0]", Null},
	}

	runVmTests(t, testCases)
}

func TestConditionals(t *testing.T) {
	testCases := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 } ", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Null},
		{"if (false) {10}", Null},
	}

	runVmTests(t, testCases)
}

func TestGlobalLetStatements(t *testing.T) {
	testCases := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}

	runVmTests(t, testCases)
}

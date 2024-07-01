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

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tc := range tests {
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
		err := testIntegerObject(t, int64(expected), actual)
		require.NoError(t, err)
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testIntegerObject(t *testing.T, expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)

	require.True(t, ok)

	require.Equal(t, expected, result.Value)
	return nil
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
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
	}

	runVmTests(t, tests)
}

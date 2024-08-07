package vm

import (
	"fmt"
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
	case *object.Error:
		errObj, ok := actual.(*object.Error)
		require.True(t, ok)
		require.Equal(t, expected.Message, errObj.Message)
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

	fmt.Println(actual.Inspect())

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

func TestFunctionsWithoutReturnValue(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let noReturn = fn() { };
			noReturn();
			`,
			expected: Null,
		},
		{
			input: `
			let noReturn = fn() { };
			let noReturnTwo = fn() { noReturn(); };
			noReturn();
			noReturnTwo();
			`,
			expected: Null,
		},
	}

	runVmTests(t, testCases)
}

func TestFirstClassFunctions(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
	let returnsOne = fn() { 1; };
	let returnsOneReturner = fn() { returnsOne; };
	returnsOneReturner()();
	`,
			expected: 1,
		},
		{
			input: `
			let returnsOneReturner = fn() {
			let returnsOne = fn() { 1; };
			returnsOne;
			};
			returnsOneReturner()();
			`,
			expected: 1,
		},
	}

	runVmTests(t, testCases)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let one = fn() { let one = 1; one };
			one();
			`,
			expected: 1,
		},
		{
			input: `
		let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
		oneAndTwo();
		`,
			expected: 3,
		},
		{
			input: `
		let vre = 10;
		let oneAndTwo = fn() { let one = 1; let two = 2; one + two + vre; };
		let threeAndFour = fn() { let three = 3; let four = 4; three + four; };
		oneAndTwo() + threeAndFour();
		`,
			expected: 20,
		},
		{
			input: `
		let firstFoobar = fn() { 
			let foobar = 50; 
			foobar; 
		};
		let secondFoobar = fn() { 
			let foobar = 100; 
			foobar; 
		};
		firstFoobar() + secondFoobar();
		`,
			expected: 150,
		},
	}

	runVmTests(t, testCases)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let identity = fn(a) { a; };
			identity(4);
			`,
			expected: 4,
		},
		{
			input: `
			let sum = fn(a, b) { a + b; };
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = fn(a, b) {
			let c = a + b;
			c;
			};
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = fn(a, b) {
			let c = a + b;
			c;
			};
			sum(1, 2) + sum(3, 4);`,
			expected: 10,
		},
		{
			input: `
			let sum = fn(a, b) {
			let c = a + b;
			c;
			};
			let outer = fn() {
			sum(1, 2) + sum(3, 4);
			};
			outer();
			`,
			expected: 10,
		},
		{
			input: `
			let globalNum = 10;
			let sum = fn(a, b) {
			let c = a + b;
			c + globalNum;
			};
			let outer = fn() {
			sum(1, 2) + sum(3, 4) + globalNum;
			};
			outer() + globalNum;
			`,
			expected: 50,
		},
	}

	runVmTests(t, testCases)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	testCases := []vmTestCase{
		{
			input:    `fn() { 1; }(1);`,
			expected: `wrong number of arguments: want=0, got=1`,
		},
		{
			input:    `fn(a) { a; }();`,
			expected: `wrong number of arguments: want=1, got=0`,
		},
		{
			input:    `fn(a, b) { a + b; }(1);`,
			expected: `wrong number of arguments: want=2, got=1`,
		},
	}

	for _, tc := range testCases {
		program := parse(tc.input)
		comp := compiler.New()
		err := comp.Compile(program)

		require.NoError(t, err)
		vm := New(comp.Bytecode())

		err = vm.Run()
		require.NotNil(t, err)

		require.Equal(t, tc.expected, err.Error())
	}
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let fivePlusTen = fn() { 5 + 10; };
			fivePlusTen();
			`,
			expected: 15,
		},
		{
			input: `
			let one = fn() { 1; };
			let two = fn() { 2; };
			one() + two()
			`,
			expected: 3,
		},
		{
			input: `
			let a = fn() { 1 };
			let b = fn() { a() + 1 };
			let c = fn() { b() + 1 };
			c();
			`,
			expected: 3,
		},
		{
			input: `
			let earlyExit = fn() { return 99; 100; };
			earlyExit();
			`,
			expected: 99,
		},
		{
			input: `
			let earlyExit = fn() { return 99; return 100; };
			earlyExit();
			`,
			expected: 99,
		},
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

func TestBuiltinFunctions(t *testing.T) {
	testCases := []vmTestCase{
		// {`len("")`, 0},
		// {`len("four")`, 4},
		// {`len("hello world")`, 11},
		// {
		// 	`len(1)`,
		// 	&object.Error{
		// 		Message: "argument to `len` not supported, got INTEGER",
		// 	},
		// },
		// {`len("one", "two")`,
		// 	&object.Error{
		// 		Message: "wrong number of arguments. got=2, want=1",
		// 	},
		// },
		{`len([1, 2, 3])`, 3},
		// {`len([])`, 0},
		// {`puts("hello", "world!")`, Null},
		// {`first([1, 2, 3])`, 1},
		// {`first([])`, Null},
		// {`first(1)`,
		// 	&object.Error{
		// 		Message: "argument to `first` must be ARRAY, got INTEGER",
		// 	},
		// },
		// {`last([1, 2, 3])`, 3},
		// {`last([])`, Null},
		// {`last(1)`,
		// 	&object.Error{
		// 		Message: "argument to `last` must be ARRAY, got INTEGER",
		// 	},
		// },
		// {`rest([1, 2, 3])`, []int{2, 3}},
		// {`rest([])`, Null},
		// {`push([], 1)`, []int{1}},
		// {`push(1, 1)`,
		// 	&object.Error{
		// 		Message: "argument to `push` must be ARRAY, got INTEGER",
		// 	},
		// },
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

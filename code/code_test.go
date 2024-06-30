package code

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMake(t *testing.T) {
	testCases := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
	}

	for _, tc := range testCases {
		instructions := Make(tc.op, tc.operands...)

		require.Equal(t, len(tc.expected), len(instructions))

		for i, b := range tc.expected {
			require.Equal(t, b, instructions[i])
		}
	}
}

func TestReadOperands(t *testing.T) {
	testCases := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}

	for _, tc := range testCases {
		instructions := Make(tc.op, tc.operands...)

		def, err := Lookup(byte(tc.op))

		require.NoError(t, err)

		operandsRead, n := ReadOperands(def, instructions[1:])
		require.Equal(t, tc.bytesRead, n)

		for i, expectedOperand := range tc.operands {
			require.Equal(t, operandsRead[i], expectedOperand)
		}
	}
}

func TestInstructionString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
`

	concatted := Instructions{}
	for _, instruction := range instructions {
		concatted = append(concatted, instruction...)
	}

	require.Equal(t, concatted.String(), expected)
}

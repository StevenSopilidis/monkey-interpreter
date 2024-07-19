package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Index: 0, Scope: GlobalScope},
		"b": {Name: "b", Index: 1, Scope: GlobalScope},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	require.Equal(t, expected["a"], a)

	b := global.Define("b")
	require.Equal(t, expected["b"], b)
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()

	global.Define("a")
	global.Define("b")

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, exp := range expected {
		result, ok := global.Resolve(exp.Name)
		require.True(t, ok)
		require.Equal(t, exp, result)
	}
}

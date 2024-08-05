package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()

	global.Define("a")
	global.Define("b")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, exp := range expected {
		result, ok := global.Resolve(exp.Name)
		require.True(t, ok)
		require.Equal(t, exp, result)
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()

	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, exp := range expected {
		result, ok := local.Resolve(exp.Name)
		require.True(t, ok)
		require.Equal(t, exp, result)
	}
}

func TestResolveNestedLoacl(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	testCases := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tc := range testCases {
		for _, expected := range tc.expectedSymbols {
			result, ok := tc.table.Resolve(expected.Name)
			require.True(t, ok)
			require.Equal(t, expected, result)
		}
	}
}

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()
	a := global.Define("a")
	require.Equal(t, a, expected["a"])

	b := global.Define("b")
	require.Equal(t, b, expected["b"])

	firstLocal := NewEnclosedSymbolTable(global)
	c := firstLocal.Define("c")
	require.Equal(t, c, expected["c"])

	d := firstLocal.Define("d")
	require.Equal(t, d, expected["d"])

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	e := secondLocal.Define("e")
	require.Equal(t, e, expected["e"])

	f := secondLocal.Define("f")
	require.Equal(t, f, expected["f"])
}

func TestDefineResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "c", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			require.True(t, ok)

			require.Equal(t, sym, result)
		}
	}
}

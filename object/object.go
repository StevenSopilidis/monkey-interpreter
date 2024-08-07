package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/stevensopilidis/monkey/ast"
	"github.com/stevensopilidis/monkey/code"
)

type ObjectType string

const (
	INTEGER_OBJ              = "INTEGER"
	BOOLEAN_OBJ              = "BOOLEAN"
	FLOAT_OBJ                = "FLOAT"
	NULL_OBJ                 = "NULL"
	RETURN_VALUE_OBJ         = "RETURN_VALUE"
	ERROR_OBJ                = "ERROR"
	FUNCTION_OBJ             = "FUNCTION"
	STRING_OBJ               = "STRING"
	Builtin_OBJ              = "Builtin"
	ARRAY_OBJ                = "ARRAY"
	HASH_OBJ                 = "HASH"
	COMPILED_FUNCTION_OBJECT = "COMPILED_FUNCTION"
)

// environment will keep track of the values of the identifiers
type Environment struct {
	store map[string]Object
	// env that current env is enclosed by
	outer *Environment
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		// check if identifier exists in outside environment
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// struct that will be used to index internal hash maps
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// interface that needs to be implemented by structs that are Hashable
type Hashable interface {
	HashKey() HashKey
}

func (b Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (i Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type HashPair struct {
	Key   Object
	Value Object
}

// struct representing hash_map
type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h Hash) Type() ObjectType {
	return HASH_OBJ
}

func (h Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// struct representing array
type Array struct {
	Elements []Object
}

func (arr Array) Type() ObjectType {
	return ARRAY_OBJ
}

func (arr Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}

	for _, e := range arr.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// built in function
type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b Builtin) Type() ObjectType {
	return Builtin_OBJ
}

func (b Builtin) Inspect() string {
	return "Builtin function"
}

// struct representing a string
type String struct {
	Value string
}

func (s String) Type() ObjectType {
	return STRING_OBJ
}

func (s String) Inspect() string {
	return s.Value
}

// struct that represensts an already compiled function
type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int // number of local bindings used by the function
	NumParameters int // nunmber of parameters of function
}

func (cf *CompiledFunction) Type() ObjectType {
	return COMPILED_FUNCTION_OBJECT
}

func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

// struct that represents a function
type Function struct {
	Parameters []ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f Function) Type() ObjectType {
	return FUNCTION_OBJ
}
func (f Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

// struct that defines an error
type Error struct {
	Message string
}

func (e Error) Type() ObjectType {
	return ERROR_OBJ
}
func (e Error) Inspect() string {
	return "ERROR: " + e.Message
}

// struct that wraps a return value
type ReturnValue struct {
	Value Object
}

func (rv ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}
func (rv ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

// struct that will wrap every object (type) in our language
type Object interface {
	Type() ObjectType
	Inspect() string
}

// internal representation of integer
type Integer struct {
	Value int64
}

func (i Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i Integer) Type() ObjectType {
	return INTEGER_OBJ
}

// internal representation of booleans
type Boolean struct {
	Value bool
}

func (i Boolean) Inspect() string {
	return fmt.Sprintf("%t", i.Value)
}

func (i Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}

// internal representation of float
type Float struct {
	Value float64
}

func (i Float) Inspect() string {
	return fmt.Sprintf("%f", i.Value)
}

func (i Float) Type() ObjectType {
	return FLOAT_OBJ
}

// internal represenation of null object
type Null struct{}

func (n Null) Inspect() string {
	return "null"
}

func (i Null) Type() ObjectType {
	return NULL_OBJ
}

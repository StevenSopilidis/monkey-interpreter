package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/stevensopilidis/monkey/ast"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	FLOAT_OBJ        = "FLOAT"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BULTIN_OBJ       = "BULTIN"
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

// built in function
type BultinFunction func(args ...Object) Object

type Bultin struct {
	Fn BultinFunction
}

func (b Bultin) Type() ObjectType {
	return BULTIN_OBJ
}

func (b Bultin) Inspect() string {
	return "bultin function"
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

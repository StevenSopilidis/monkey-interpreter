package object

import "fmt"

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	FLOAT_OBJ        = "FLOAT"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
)

// environment will keep track of the values of the identifiers
type Environment struct {
	store map[string]Object
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
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

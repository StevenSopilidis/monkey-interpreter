package object

import "fmt"

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	FLOAT_OBJ        = "FLOAT"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
)

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
	return INTEGER_OBJ
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

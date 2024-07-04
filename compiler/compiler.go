package compiler

import (
	"fmt"

	"github.com/stevensopilidis/monkey/ast"
	"github.com/stevensopilidis/monkey/code"
	"github.com/stevensopilidis/monkey/object"
)

type Compiler struct {
	instructions code.Instructions // generated bytecode
	constants    []object.Object   // constant pool
}

// Represents the Instructions the compiler generated
// and the constants the compiler evaluated
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		// after each expression statement append opPop opcode to
		// clean the stack
		c.emit(code.OpPop)
	case ast.InfixExpression:
		// less-than operator (<) just reorder left and right branches
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}

			c.emit(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "/":
			c.emit(code.OpDiv)
		case "*":
			c.emit(code.OpMul)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case ast.IntegerLiteral:
		// parse integerliteral and push it to constant pool
		integer := &object.Integer{Value: node.Value}
		/// c.addConstant(integer)) ---> pos of our integerConstant inside the constant pool
		c.emit(code.OpConstant, c.addConstant(integer))

	case ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	}

	return nil
}

// function for generating an instruction
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

// function for pushing instruction into compiler's instruction set
func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

// function for appending an constant int the constant pool
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	// returns identifier of the constant
	return len(c.constants) - 1
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

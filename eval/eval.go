package eval

import (
	"fmt"

	"github.com/stevensopilidis/monkey/ast"
	"github.com/stevensopilidis/monkey/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case ast.Identifier:
		return evalIdentifier(node, env)
	case ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case ast.IfExpression:
		return evalIfExpression(node, env)
	}

	return nil
}

// function for evaluating an identifier
func evalIdentifier(node ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}

	return val
}

// function for checking if object is Error
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}

	return false
}

// function for creating a new Error message
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// function for evaluating if-else expressions
func evalIfExpression(ie ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

// function that evaluates if a condition is truthy (not false or null)
func isTruthy(obj object.Object) bool {
	if obj == NULL || obj == FALSE {
		return false
	}

	return true
}

// function for evaluating a block statement
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
			return result
		}
	}
	return result
}

// function for evaluating an infix expression
func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	_, okBoolLeft := left.(*object.Boolean)
	_, okBoolRight := right.(*object.Boolean)

	if okBoolLeft && okBoolRight && operator == "==" {
		return nativeBoolToBooleanObject(left == right)
	}
	if okBoolLeft && okBoolRight && operator == "!=" {
		return nativeBoolToBooleanObject(left != right)
	}

	if okBoolLeft != okBoolRight {
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	}

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return evalIntegerInfixExpression(operator, left, right)
	}

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ {
		value := left.(*object.Integer).Value
		left = &object.Float{Value: float64(value)}
		return evalFloatInfixExpression(operator, left, right)
	}

	if left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ {
		value := right.(*object.Integer).Value
		right = &object.Float{Value: float64(value)}
		return evalFloatInfixExpression(operator, left, right)
	}

	if left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ {
		return evalFloatInfixExpression(operator, left, right)
	}

	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

// function for evaluating infix expression where at least operands are floats
func evalFloatInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch operator {
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	default:
		return NULL
	}
}

// function for evaluating infix expression where the operands are integers
func evalIntegerInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// function for evaluating a prefix expression
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

// function for evaluating minus operator
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() == object.INTEGER_OBJ {
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	}
	if right.Type() == object.FLOAT_OBJ {
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}
	}

	return newError("unknown operator: -%s", right.Type())
}

// function for evaluating bang operator
func evalBangOperator(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// function that takes ast.Boolean and returns reference to
// on of the two predifined obj.Boolean
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

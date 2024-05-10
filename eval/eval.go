package eval

import (
	"github.com/stevensopilidis/monkey/ast"
	"github.com/stevensopilidis/monkey/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements)
	case ast.ExpressionStatement:
		return Eval(node.Expression)
	case ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node)
	case ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		return object.ReturnValue{Value: val}
	case ast.IfExpression:
		return evalIfExpression(node)
	}

	return nil
}

// function for evaluating if-else expressions
func evalIfExpression(ie ast.IfExpression) object.Object {
	condition := Eval(ie.Condition)

	if isTruthy(condition) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
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
func evalBlockStatement(block *ast.BlockStatement) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement)
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
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

	return NULL
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
		return NULL
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
		return NULL
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

	return NULL
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

func evalProgram(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt)

		if returnValue, ok := result.(object.ReturnValue); ok {
			// if returnValue encountered return it without evaluating
			// left over statements
			return returnValue.Value
		}
	}

	return result
}

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
	case ast.StringLiteral:
		return object.String{Value: node.Value}
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
	case ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return object.Function{Parameters: params, Env: env, Body: body}
	case ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		// evaluate the expressions of the arguments
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	case ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)
	case ast.HashLiteral:
		return evalHashLiteral(node, env)
	}
	return nil
}

// function for evaluating a HashLiteral
func evalHashLiteral(node ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

// function for evaluating IndexExpressions
func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

// function for evaluating indexing at hashes
func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]

	if !ok {
		return NULL
	}

	return pair.Value
}

// function for evaluating IndexExpression for arrays
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObj := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements))

	if idx < 0 || idx >= max {
		return NULL
	}

	return arrayObj.Elements[idx]
}

// function for returning result from function
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case object.Function:
		extendedEnv := extendedFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case object.Bultin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

// function for created extended env for a function
func extendedFunctionEnv(fn object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	// overwrite outer env bindings
	for paramsIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramsIdx])
	}

	return env
}

// function for unwrapping the return value from a function call
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// function for evaluating group of expressions
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, exp := range exps {
		evaluated := Eval(exp, env)
		// if there is an error return an array with a single element
		// that contains the error object
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// function for evaluating an identifier
func evalIdentifier(node ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if bultin, ok := bultins[node.Value]; ok {
		return bultin
	}

	return newError("identifier not found: " + node.Value)
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

	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return evalStringInfixExpression(operator, left, right)
	}

	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

// function for evaluating infix operations applied to strings
func evalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

	leftVal := left.(object.String).Value
	rightVal := right.(object.String).Value
	return object.String{Value: leftVal + rightVal}
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

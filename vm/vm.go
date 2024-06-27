package vm

import (
	"fmt"

	"github.com/stevensopilidis/monkey/code"
	"github.com/stevensopilidis/monkey/compiler"
	"github.com/stevensopilidis/monkey/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	sp           int // stack pointer
}

func New(byteCode *compiler.Bytecode) *VM {
	return &VM{
		constants:    byteCode.Constants,
		instructions: byteCode.Instructions,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++

	return nil
}

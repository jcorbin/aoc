package elvm

import (
	"errors"
	"fmt"
)

// VM is the virtual machine itself.
type VM struct {
	R Registers

	// IP holds the address of the next operation to run.
	IP *int

	// N counts the number of operations perform.
	N int

	// Ops of the currently loaded Program.
	Ops []Instruction
}

// Registers holds register values from a VM.
type Registers [6]int

// Instruction represents a single VM operation. The first element is the
// opcode, the remaining 3 are operands.
type Instruction [4]int

// Name returns the mnemonic name of the instruction.
func (in Instruction) Name() string { return names[in[0]] }

// Program represents a program runnable on a VM.
type Program struct {
	Ops []Instruction

	// IPReg is the register number used for control flow (to store the
	// instruction pointer).
	IPReg int
}

var names = [16]string{
	"addr", "addi",
	"mulr", "muli",
	"banr", "bani",
	"borr", "bori",
	"setr", "seti",
	"gtir", "gtri", "gtrr",
	"eqir", "eqri", "eqrr",
}

var errLimitExceeded = errors.New("operation limit exceeded")

// Load a program into the vm, initializing IP and Ops.
func (vm *VM) Load(prog Program) error {
	if prog.IPReg < 0 {
		pc := 0
		vm.IP = &pc
	} else {
		vm.IP = &vm.R[prog.IPReg]
	}
	vm.Ops = prog.Ops
	// TODO validate / compile ops
	return nil
}

// Step executes a single operation.
func (vm *VM) Step() {
	vm.N++
	in := vm.Ops[*vm.IP]
	switch op, a, b, c := in[0], in[1], in[2], in[3]; op {
	case 0:
		vm.R[c] = vm.R[a] + vm.R[b]
	case 1:
		vm.R[c] = vm.R[a] + b
	case 2:
		vm.R[c] = vm.R[a] * vm.R[b]
	case 3:
		vm.R[c] = vm.R[a] * b
	case 4:
		vm.R[c] = vm.R[a] & vm.R[b]
	case 5:
		vm.R[c] = vm.R[a] & b
	case 6:
		vm.R[c] = vm.R[a] | vm.R[b]
	case 7:
		vm.R[c] = vm.R[a] | b
	case 8:
		vm.R[c] = vm.R[a]
	case 9:
		vm.R[c] = a
	case 10:
		vm.R[c] = btoi(a > vm.R[b])
	case 11:
		vm.R[c] = btoi(vm.R[a] > b)
	case 12:
		vm.R[c] = btoi(vm.R[a] > vm.R[b])
	case 13:
		vm.R[c] = btoi(a == vm.R[b])
	case 14:
		vm.R[c] = btoi(vm.R[a] == b)
	case 15:
		vm.R[c] = btoi(vm.R[a] == vm.R[b])
	}
	*vm.IP++
}

// Execute runs the loaded program, with an optional limit on the number of
// operations to execute, and an optional tracer to observe machine state.
func (vm *VM) Execute(limit int, trc Tracer) error {
	if trc != nil {
		return vm.trace(limit, trc)
	}
	return vm.execute(limit)
}

func (vm *VM) execute(limit int) error {
	for i := 0; *vm.IP < len(vm.Ops); i++ {
		if limit > 0 && i > limit {
			return errLimitExceeded
		}
		vm.Step()
	}
	return nil
}

func (vm *VM) trace(limit int, trc Tracer) error {
	for i := 0; *vm.IP < len(vm.Ops); i++ {
		if limit > 0 && i > limit {
			return errLimitExceeded
		}
		in := vm.Ops[*vm.IP]
		if err := trc.Before(vm, in); err != nil {
			return err
		}
		vm.Step()
		if err := trc.After(vm, in); err != nil {
			return err
		}
	}
	return nil
}

func (r Registers) eq(other Registers) bool {
	for i, rv := range r {
		if rv != other[i] {
			return false
		}
	}
	return true
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (vm *VM) String() string {
	return fmt.Sprintf("N:%v IP:%v R:%v", vm.N, *vm.IP, vm.R)
}

func (in Instruction) String() string {
	op, a, b, c := in[0], in[1], in[2], in[3]
	return fmt.Sprintf("%s %v %v %v", names[op], a, b, c)
}

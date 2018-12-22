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
func (in Instruction) Name() string { return opSpecs[in[0]].name }

// Program represents a program runnable on a VM.
type Program struct {
	Ops []Instruction

	// IPReg is the register number used for control flow (to store the
	// instruction pointer).
	IPReg int
}

type operandMode uint8

const (
	operandIgnore operandMode = iota
	operandValue              // immediate
	operandRead
	operandWrite
)

type opSpec struct {
	name, sym string
	a, b, c   operandMode
}

var opSpecs = [16]opSpec{
	{"addr", "+", operandRead, operandRead, operandWrite},
	{"addi", "+", operandRead, operandValue, operandWrite},
	{"mulr", "*", operandRead, operandRead, operandWrite},
	{"muli", "*", operandRead, operandValue, operandWrite},
	{"banr", "&", operandRead, operandRead, operandWrite},
	{"bani", "&", operandRead, operandValue, operandWrite},
	{"borr", "|", operandRead, operandRead, operandWrite},
	{"bori", "|", operandRead, operandValue, operandWrite},
	{"setr", "=", operandRead, operandIgnore, operandWrite},
	{"seti", "=", operandValue, operandIgnore, operandWrite},
	{"gtir", ">", operandValue, operandRead, operandWrite},
	{"gtri", ">", operandRead, operandValue, operandWrite},
	{"gtrr", ">", operandRead, operandRead, operandWrite},
	{"eqir", "==", operandValue, operandRead, operandWrite},
	{"eqri", "==", operandRead, operandValue, operandWrite},
	{"eqrr", "==", operandRead, operandRead, operandWrite},
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
	case 0: // addr
		vm.R[c] = vm.R[a] + vm.R[b]
	case 1: // addi
		vm.R[c] = vm.R[a] + b
	case 2: // mulr
		vm.R[c] = vm.R[a] * vm.R[b]
	case 3: // muli
		vm.R[c] = vm.R[a] * b
	case 4: // banr
		vm.R[c] = vm.R[a] & vm.R[b]
	case 5: // bani
		vm.R[c] = vm.R[a] & b
	case 6: // borr
		vm.R[c] = vm.R[a] | vm.R[b]
	case 7: // bori
		vm.R[c] = vm.R[a] | b
	case 8: // setr
		vm.R[c] = vm.R[a]
	case 9: // seti
		vm.R[c] = a
	case 10: // gtir
		vm.R[c] = btoi(a > vm.R[b])
	case 11: // gtri
		vm.R[c] = btoi(vm.R[a] > b)
	case 12: // gtrr
		vm.R[c] = btoi(vm.R[a] > vm.R[b])
	case 13: // eqir
		vm.R[c] = btoi(a == vm.R[b])
	case 14: // eqri
		vm.R[c] = btoi(vm.R[a] == b)
	case 15: // eqrr
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
	return fmt.Sprintf("%s %v %v %v", opSpecs[op].name, a, b, c)
}

// Describe returns a high level description of what the instruction
// does (normal infix notation).
func (prog Program) Describe(ip int) string {
	in := prog.Ops[ip]
	op, a, b, c := in[0], in[1], in[2], in[3]
	spec := opSpecs[op]
	if spec.c != operandWrite {
		return fmt.Sprint(in)
	}

	// control flow
	if c == prog.IPReg {
		switch spec.sym {

		case "=":
			switch spec.a {
			case operandValue:
				return fmt.Sprintf("goto %v", a)
			case operandRead:
				return fmt.Sprintf("goto R%v", a)
			}

		case "+":
			mode, by := spec.a, a
			if c == a {
				mode, by = spec.b, b
			}
			switch mode {
			case operandValue:
				return fmt.Sprintf("jump %+d", by)
			case operandRead:
				return fmt.Sprintf("jump +R%v", by)
			}

		}
	}

	// updates
	if c == a || c == b {
		mode, by := spec.a, a
		if c == a {
			mode, by = spec.b, b
		}
		switch mode {
		case operandValue:
			return fmt.Sprintf("R%v %s= %v", c, spec.sym, by)
		case operandRead:
			return fmt.Sprintf("R%v %s= R%v", c, spec.sym, by)
		}
	}

	if spec.b == operandIgnore {
		switch spec.a {
		case operandRead:
			return fmt.Sprintf("%v %s R%v", c, spec.sym, a)
		case operandValue:
			return fmt.Sprintf("%v %s %v", c, spec.sym, a)
		}
	}

	if spec.a == operandRead && spec.b == operandRead {
		return fmt.Sprintf("R%v = %v %s %v", c, a, spec.sym, b)
	} else if spec.a == operandRead && spec.b == operandValue {
		return fmt.Sprintf("R%v = %v %s %v", c, a, spec.sym, b)
	} else if spec.a == operandValue && spec.b == operandRead {
		return fmt.Sprintf("R%v = %v %s %v", c, a, spec.sym, b)
	}

	return in.String()
}

package elvm

import (
	"bytes"
	"io"
	"log"
	"strconv"
)

// Tracer supports debug tracing through VM.Trace.
type Tracer interface {
	Before(vm *VM, in Instruction) error
	After(vm *VM, in Instruction) error
}

type tracers []Tracer

// Tracers creates a compound tracer that calls the given tracers in order,
// stopping on first error.
func Tracers(ts ...Tracer) Tracer {
	if len(ts) == 0 {
		return nil
	}
	a := ts[0]
	for i := 1; i < len(ts); i++ {
		b := ts[i]
		if b == nil || b == Tracer(nil) {
			continue
		}
		if a == nil || a == Tracer(nil) {
			a = b
			continue
		}
		as, haveAs := a.(tracers)
		bs, haveBs := b.(tracers)
		if haveAs && haveBs {
			a = append(as, bs...)
		} else if haveAs {
			a = append(as, b)
		} else if haveBs {
			a = append(tracers{a}, bs...)
		} else {
			a = tracers{a, b}
		}
	}
	return a
}

func (ts tracers) Before(vm *VM, in Instruction) error {
	for _, t := range ts {
		if err := t.Before(vm, in); err != nil {
			return err
		}
	}
	return nil
}

func (ts tracers) After(vm *VM, in Instruction) error {
	for _, t := range ts {
		if err := t.After(vm, in); err != nil {
			return err
		}
	}
	return nil
}

// PrintTracer prints vm state to an io.Writer.
type PrintTracer struct {
	W io.Writer
	bytes.Buffer
	tmp [64]byte
}

// Before prints vm state and the given instruction into the buffer.
func (pt *PrintTracer) Before(vm *VM, in Instruction) error {
	pt.WriteVM(vm)
	pt.WriteByte(' ')
	pt.WriteInstruction(in)
	return nil
}

// After prints vm state into the buffer, then logs and resets it.
func (pt *PrintTracer) After(vm *VM, in Instruction) error {
	pt.WriteVM(vm)
	pt.WriteByte('\n')
	if _, err := pt.W.Write(pt.Bytes()); err != nil {
		return err
	}
	pt.Reset()
	return nil
}

// WriteVM writes VM state to the buffer.
func (pt *PrintTracer) WriteVM(vm *VM) {
	pt.WriteString("N:")
	pt.WriteInt(vm.N)
	pt.WriteString(" IP:")
	pt.WriteInt(*vm.IP)
	pt.WriteString(" R:[")
	pt.WriteInt(vm.R[0])
	pt.WriteInt(vm.R[1])
	pt.WriteInt(vm.R[2])
	pt.WriteInt(vm.R[3])
	pt.WriteInt(vm.R[4])
	pt.WriteInt(vm.R[5])
	pt.WriteByte(']')
}

// WriteInstruction writes an instruction to the buffer.
func (pt *PrintTracer) WriteInstruction(in Instruction) {
	op, a, b, c := in[0], in[1], in[2], in[3]
	pt.WriteString(opSpecs[op].name)
	pt.WriteByte(' ')
	pt.WriteInt(a)
	pt.WriteByte(' ')
	pt.WriteInt(b)
	pt.WriteByte(' ')
	pt.WriteInt(c)
}

// WriteInt writes a base-10 formatted number into the buffer.
func (pt *PrintTracer) WriteInt(n int) {
	pt.Write(strconv.AppendInt(pt.tmp[:0], int64(n), 10))
}

// LogTracer prints vm state to the standard "log" package.
type LogTracer struct {
	PrintTracer
}

// After prints vm state into the buffer, then logs and resets it.
func (lg *LogTracer) After(vm *VM, in Instruction) error {
	lg.WriteVM(vm)
	log.Printf("%s", lg.Bytes())
	lg.Reset()
	return nil
}

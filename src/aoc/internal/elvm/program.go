package elvm

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

// ProgramDecoder supports decoding a Program.
type ProgramDecoder struct {
	*bufio.Scanner

	opmap map[string]int
}

// NewProgramDecoder creates a new program decoder.
func NewProgramDecoder(r io.Reader) *ProgramDecoder {
	dec := &ProgramDecoder{
		Scanner: bufio.NewScanner(r),
	}
	opmap := make(map[string]int, len(opSpecs))
	for code, spec := range opSpecs {
		opmap[spec.name] = code
	}
	dec.opmap = opmap
	return dec
}

var (
	ipPat = regexp.MustCompile(`#ip (\d+)`)
	opPat = regexp.MustCompile(`(\w+) (\d+) (\d+) (\d+)`)
)

// Decode a program, returning it and any error. Stops on the first blank
// line, allowing multiple programs to be decoded from one stream (separated
// by blank lines).
func (dec *ProgramDecoder) Decode() (prog Program, err error) {
	// TODO make #ip lines optional to support control-flow-less applications
	if dec.Scan() {
		line := dec.Text()
		parts := ipPat.FindStringSubmatch(line)
		if len(parts) == 0 {
			return prog, fmt.Errorf("unexpected line %q expected %v", line, ipPat)
		}
		prog.IPReg, _ = strconv.Atoi(parts[1])
	}
	for dec.Scan() {
		line := dec.Text()
		if line == "" {
			break
		}
		parts := opPat.FindStringSubmatch(line)
		if len(parts) == 0 {
			return prog, fmt.Errorf("unexpected line %q expected %v", line, opPat)
		}
		code, def := dec.opmap[parts[1]]
		if !def {
			return prog, fmt.Errorf("invalid op %q", parts[1])
		}

		var in Instruction
		in[0] = code
		in[1], _ = strconv.Atoi(parts[2])
		in[2], _ = strconv.Atoi(parts[3])
		in[3], _ = strconv.Atoi(parts[4])
		prog.Ops = append(prog.Ops, in)
	}
	return prog, dec.Err()
}

// DecodeProgram decodes a program from the given io.Reader.
func DecodeProgram(r io.Reader) (Program, error) {
	return NewProgramDecoder(r).Decode()
}

// Package infernio provides convenience structures for dealing with
// infernal input loading, including:
//
// - a default builtin input to use when none is provided
// - what's the name of that reader?
// - dispatching input between arg/stdin/builtin
package infernio

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"

	"github.com/jcorbin/anansi"
)

const (
	unknownReaderName = "<unknown>"
	builtinReaderName = "<builtin>"
)

// ReaderName returns a name string for the given reader, defaulting to unknownReaderName.
func ReaderName(r io.Reader) string {
	if nom, ok := r.(interface{ Name() string }); ok {
		return nom.Name()
	}
	return unknownReaderName
}

// Builtin creates a NamedReader around a bytes.Reader built from the
// given string, whose name is builtinReaderName.
func Builtin(s string) NamedReader {
	r := bytes.NewReader([]byte(s))
	return NamedReader{r, builtinReaderName}
}

// NamedReader attaches a name to an arbitrary io.Reader.
type NamedReader struct {
	io.Reader
	Nom string
}

// Name returns the nr.Nom.
func (nr NamedReader) Name() string { return nr.Nom }

// SelectInput returns the an io.ReadCloser that will provide the input
// specified by the user at the command line: if the user gave a positional
// argument, read from that file; if stdin is not a terminal, read from it,
// otherwise use the given builtin input, or return an error if that's nil.
func SelectInput(builtin io.Reader) (io.ReadCloser, error) {
	if name := flag.Arg(0); name != "" {
		f, err := os.Open(name)
		return f, err
	}
	if !anansi.IsTerminal(os.Stdin) {
		return ioutil.NopCloser(os.Stdin), nil
	}
	if builtin == nil {
		return nil, errors.New("no input specified")
	}
	return ioutil.NopCloser(builtin), nil
}

// LoadInput calls the given loader function with the io.ReadCloser selected by
// SelectInput, closing it afterwards.
func LoadInput(builtin io.Reader, loader func(r io.Reader) error) error {
	rc, err := SelectInput(builtin)
	if err == nil {
		err = loader(rc)
		if cerr := rc.Close(); err == nil {
			err = cerr
		}
	}
	return err
}

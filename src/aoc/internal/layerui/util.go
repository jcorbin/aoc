package layerui

import (
	"fmt"
	"os"
	"syscall"

	"github.com/jcorbin/anansi"
)

// OpenTermFiles opens /dev/tty if the given files are not terminals,
// returning a usable pair of terminal in/out file handles. It also redirects
// "log" package output to the Logs buffer if it hasn't already been done
// (e.g. by calling OpenLogFile).
func OpenTermFiles(in, out *os.File) (_, _ *os.File, rerr error) {
	if !anansi.IsTerminal(in) {
		f, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
		if err != nil {
			return nil, nil, err
		}
		defer func() {
			if rerr != nil {
				in.Close()
			}
		}()
		in = f
	}
	if !anansi.IsTerminal(out) {
		f, err := os.OpenFile("/dev/tty", syscall.O_WRONLY, 0)
		if err != nil {
			return nil, nil, err
		}
		out = f
	}
	InitLogs()
	return in, out, nil
}

// MustOpenTermFiles is a convenience wrapper for OpenTermFiles that prints to
// stderr and exits on error.
func MustOpenTermFiles(in, out *os.File) (_, _ *os.File) {
	in, out, err := OpenTermFiles(in, out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open terminal files: %v", err)
		os.Exit(1)
	}
	return in, out
}

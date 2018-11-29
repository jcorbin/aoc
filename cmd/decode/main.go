package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	rawMode   = flag.Bool("raw", false, "enable terminal raw mode")
	mouseMode = flag.Bool("mouse", false, "enable terminal mouse reporting")
	altMode   = flag.Bool("alt", false, "enable alternate screen usage")
	stripMode = flag.Bool("strip", false, "strip escape sequences when running non-interactively")
)

func main() {
	flag.Parse()
	switch err := run(os.Stdin, os.Stdout); err {
	case nil:
	case io.EOF:
		fmt.Println(err)
	default:
		log.Fatal(err)
	}
}

func run(in, out *os.File) error {
	if !anansi.IsTerminal(in) || !anansi.IsTerminal(out) {
		return runBatch(in, out)
	}

	term := anansi.NewTerm(out)

	if *mouseMode {
		term.AddMode(
			ansi.ModeMouseSgrExt,
			ansi.ModeMouseBtnEvent,
			ansi.ModeMouseAnyEvent,
		)
	}

	if *altMode {
		term.AddMode(
			ansi.ModeAlternateScreen,
		)
	}

	term.SetEcho(!*rawMode)
	term.SetRaw(*rawMode)

	return term.RunWith(runInteractive)
}

func runBatch(in, out *os.File) (err error) {
	var bufw = bufio.NewWriter(out)
	defer func() {
		if ferr := bufw.Flush(); err == nil {
			err = ferr
		}
	}()

	const readSize = 4096
	var buf bytes.Buffer
	for err == nil {
		buf.Grow(readSize)
		err = readMore(&buf, in)
		if perr := processBatch(&buf, bufw); err == nil {
			err = perr
		}
	}
	if buf.Len() > 0 {
		log.Printf("undecoded trailer content: %q", buf.Bytes())
	}
	return err
}

func processBatch(buf *bytes.Buffer, w io.Writer) (err error) {
	writeRune := func(r rune) (size int, err error) {
		var b [4]byte
		n := utf8.EncodeRune(b[:], r)
		return w.Write(b[:n])
	}
	if rw := w.(interface {
		WriteRune(r rune) (size int, err error)
	}); rw != nil {
		writeRune = rw.WriteRune
	}

	for err == nil && buf.Len() > 0 {
		e, a, n := ansi.DecodeEscape(buf.Bytes())
		if n > 0 {
			buf.Next(n)
		}
		if e == 0 {
			r, n := utf8.DecodeRune(buf.Bytes())
			switch r {
			case 0x90, 0x9B, 0x9D, 0x9E, 0x9F: // DCS, CSI, OSC, PM, APC
				return
			case 0x1B: // ESC
				if p := buf.Bytes(); len(p) == cap(p) {
					return
				}
			}
			buf.Next(n)
			_, err = writeRune(r)
		} else if !*stripMode {
			if e == ansi.SGR {
				attr, _, decErr := ansi.DecodeSGR(a)
				if decErr == nil {
					_, err = fmt.Fprintf(w, "[ansi:SGR %v]", attr)
				} else {
					_, err = fmt.Fprintf(w, "[ansi:SGR ERR:%v %q]", decErr, a)
				}
			} else if len(a) > 0 {
				_, err = fmt.Fprintf(w, "[ansi:%v %q]", e, a)
			} else {
				_, err = fmt.Fprintf(w, "[ansi:%v]", e)
			}
		}
	}
	return err
}

func readMore(buf *bytes.Buffer, r io.Reader) error {
	b := buf.Bytes()
	b = b[len(b):cap(b)]
	n, err := r.Read(b)
	b = b[:n]
	buf.Write(b)
	return err
}

func runInteractive(term *anansi.Term) error {
	const minRead = 128
	var buf bytes.Buffer
	for {
		// read more input…
		buf.Grow(minRead)
		p := buf.Bytes()
		p = p[len(p):cap(p)]
		n, err := os.Stdin.Read(p)
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		_, _ = buf.Write(p[:n])

		// …and process it
		if err := process(term, &buf); err != nil {
			return err
		}
	}
}

func process(term *anansi.Term, buf *bytes.Buffer) error {
	for buf.Len() > 0 {
		// Try to decode an escape sequence…
		e, a, n := ansi.DecodeEscape(buf.Bytes())
		if n > 0 {
			buf.Next(n)
		}

		// …fallback to decoding a rune otherwise…
		if e == 0 {
			r, n := utf8.DecodeRune(buf.Bytes())
			switch r {
			case 0x90, 0x9D, 0x9E, 0x9F: // DCS, OSC, PM, APC
				return nil // …need more bytes to complete a partial string.

			case 0x9B: // CSI
				return nil // …need more bytes to complete a partial control sequence.

			case 0x1B: // ESC
				if p := buf.Bytes(); len(p) == cap(p) {
					return nil // …need more bytes to determine if an escape sequence can be decoded.
				}
				// …pass as literal ESC…
			}

			// …consume and handle the rune.
			buf.Next(n)
			e = ansi.Escape(r)
		}

		handle(term, e, a)
	}
	return nil
}

var prior ansi.Escape

func handle(term *anansi.Term, e ansi.Escape, a []byte) {
	fmt.Printf("%U %v", e, e)

	if len(a) > 0 {
		fmt.Printf(" %q", a)
	}

	// print detail for mouse reporting
	if e == ansi.CSI('M') || e == ansi.CSI('m') {
		btn, pt, err := ansi.DecodeXtermExtendedMouse(e, a)
		if err != nil {
			fmt.Printf(" mouse-err:%v", err)
		} else {
			fmt.Printf(" mouse-%v@%v", btn, pt)
		}
	}

	switch e {

	// ^C to quit
	case 0x03:
		if prior == 0x03 {
			panic("goodbye")
		} else {
			fmt.Printf(" \x1b[91m<press Ctrl-C again to quit>\x1b[0m")
		}

	// ^L to clear
	case 0x0c:
		if prior == 0x0c {
			fmt.Printf("\x1b[2J\x1b[H") // 2 ED CUP
		} else {
			fmt.Printf(" \x1b[93m<press Ctrl-L again to quit>\x1b[0m")
		}

	// ^Z to suspend
	case 0x1a:
		if prior == 0x1a {
			if err := term.RunWithout(suspend); err != nil {
				panic(err)
			}
		} else {
			fmt.Printf(" \x1b[92m<press Ctrl-Z again to suspend>\x1b[0m")
		}

	}

	prior = e

	fmt.Printf("\r\n")
}

func suspend(_ *anansi.Term) error {
	cont := make(chan os.Signal)
	signal.Notify(cont, syscall.SIGCONT)
	log.Printf("suspending")
	if err := syscall.Kill(0, syscall.SIGTSTP); err != nil {
		return err
	}
	<-cont
	log.Printf("resumed")
	return nil
}

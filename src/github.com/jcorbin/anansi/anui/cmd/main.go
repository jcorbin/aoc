package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/jcorbin/anansi/anui"
	anansitest "github.com/jcorbin/anansi/test"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	recs, err := readRecords(os.Stdin)
	if err != nil {
		return err
	}

	for i, rec := range recs {
		if rec.Ok {
			continue
		}

		if rec.Virtual.Rect != rec.Audit.Rect {
			log.Printf("rec[%v] rect mismatch virtual:%v audit:%v)", i, rec.Virtual.Rect, rec.Audit.Rect)
			continue
		}

		auditRunes, _ := anansitest.GridRowData(rec.Audit.Grid)
		auditLines := anansitest.RunesToLines(auditRunes, ' ')
		for j, line := range auditLines {
			log.Printf("rec[%v] line[%v] %q", i, j, line)
		}

		any := false
		for pt := rec.Audit.Rect.Min; pt.Y < rec.Audit.Rect.Max.Y; pt.Y++ {
			for pt.X = rec.Audit.Rect.Min.X; pt.X < rec.Audit.Rect.Max.X; pt.X++ {
				j, _ := rec.Audit.CellOffset(pt)
				ar, vr := rec.Audit.Rune[j], rec.Virtual.Rune[j]
				if ar == 0 {
					ar = ' '
				}
				if vr == 0 {
					vr = ' '
				}
				if ar != vr || rec.Audit.Attr[j] != rec.Virtual.Attr[j] {
					log.Printf(
						"rec[%v] mismatch @%v virtual:%q %v audit:%q %v",
						i, pt,
						string(rec.Virtual.Rune[j]), rec.Virtual.Attr[j],
						string(rec.Audit.Rune[j]), rec.Audit.Attr[j],
					)
					any = true
				}
			}
		}

		if !any {
			log.Printf("rec[%v] ok... %v?", i, rec.Virtual.Eq(rec.Audit.Grid, ' '))
		}
	}

	return nil
}

func readRecords(r io.Reader) (recs []anui.AuditRecord, _ error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var rec anui.AuditRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	return recs, sc.Err()
}

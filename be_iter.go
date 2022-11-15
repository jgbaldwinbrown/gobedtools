package bedtools

import (
	"fmt"
	"bytes"
	"io"
)

type BedIter interface {
	Next() (BedEntry, bool)
}

type BedChan struct {
	BedEChan chan BedEntry
}

func (b BedChan) Next() (BedEntry, bool) {
	be, ok := <-b.BedEChan
	return be, ok
}

func WriteBed(bi BedIter, w io.Writer) error {
	for b, ok := bi.Next(); ok; b, ok = bi.Next() {
		b.Fprint(w)
		fmt.Fprintf(w, "\n")
	}
	return nil
}

type BedReader struct{io.Reader}

func (b BedReader) Bed() (io.Reader, error) {
	return b.Reader, nil
}

func (b BedEntry) Bed() (io.Reader, error) {
	var buf bytes.Buffer
	b.Fprint(&buf)
	fmt.Fprintf(&buf, "\n")
	return &buf, nil
}

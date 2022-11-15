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
	BedEChan <-chan BedEntry
}

func (b BedChan) Next() (BedEntry, bool) {
	be, ok := <-b.BedEChan
	return be, ok
}

func (b BedChan) Bed() (io.Reader, error) {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		err := WriteBed(b, pw)
		if err != nil {
			panic(err)
		}
	}()
	return pr, nil
}

func (b Bed) BedChan() <-chan BedEntry {
	out := make(chan BedEntry, 256)
	go func() {
		for _, e := range b{
			out <- e
		}
		close(out)
	}()
	return out
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

func CollectBed(bi BedIter) []BedEntry {
	var bed []BedEntry
	for val, ok := bi.Next(); ok; val, ok = bi.Next() {
		bed = append(bed, val)
	}
	return bed
}

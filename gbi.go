package bedtools

import (
	"fmt"
	"strconv"
	"bufio"
	"strings"
	"os/exec"
	"os"
	"io"
)

func IntersectCore(abed io.Reader, opts []string, bpath string) (*exec.Cmd, io.Reader, error) {
	cmdstrs := []string{"bedtools", "intersect"}
	cmdstrs = append(cmdstrs, opts...)
	cmdstrs = append(cmdstrs, "-a", "-", "-b", bpath)
	cmd := exec.Command(cmdstrs[0], cmdstrs[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	cmd.Stdin = abed
	cmd.Stderr = os.Stderr
	return cmd, stdout, nil
}

type BedEntry struct {
	Chr string
	Start int64
	End int64
	Fields []string
	Err error
}

type Bedder interface {
	Bed() (io.Reader, error)
}

type Bed []BedEntry

func (b Bed) Bed() (io.Reader, error) {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		for _, entry := range b {
			fmt.Fprintf(pw, "%v\t%v\t%v", entry.Chr, entry.Start, entry.End)
			for _, field := range entry.Fields {
				fmt.Fprintf(pw, "\t%v", field)
			}
			fmt.Fprintf(pw, "\n")
		}
	}()
	return pr, nil
}

func IntersectBedReader(abed io.Reader, opts []string, bpath string) (<-chan BedEntry, error) {
	cmd, r, err := IntersectCore(abed, opts, bpath)
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	entries := make(chan BedEntry, 256)
	go func() {
		defer close(entries)
		s := bufio.NewScanner(r)
		s.Buffer([]byte{}, 1e12)

		for s.Scan() {
			line := strings.Split(s.Text(), "\t")
			if len(line) < 3 {
				entries <- BedEntry{Err: fmt.Errorf("len(line) %v too short", len(line))}
			}
			var entry BedEntry
			entry.Chr = line[0]
			entry.Start, entry.Err = strconv.ParseInt(line[1], 0, 64)
			if entry.Err != nil {
				entries <- entry
			}
			entry.End, entry.Err = strconv.ParseInt(line[2], 0, 64)
			if entry.Err != nil {
				entries <- entry
			}
			entry.Fields = line[3:]
			entries <- entry
		}
	}()

	return entries, nil
}

func IntersectBed(abed Bedder, opts []string, bpath string) (<-chan BedEntry, error) {
	bed, err := abed.Bed()
	if err != nil {
		return nil, err
	}
	return IntersectBedReader(bed, opts, bpath)
}

func IntersectBeds(abed Bedder, opts []string, bbed Bedder) (<-chan BedEntry, error) {
	bfile, err := os.CreateTemp("", "intersect_*.bed")
	if err != nil {
		return nil, err
	}

	bbeddedreader, err := bbed.Bed()
	if err != nil {
		os.Remove(bfile.Name())
		return nil, err
	}
	bbedded := bbeddedreader.(io.ReadCloser)
	defer bbedded.Close()

	_, err = io.Copy(bfile, bbedded)
	bfile.Close()
	if err != nil {
		os.Remove(bfile.Name())
		return nil, err
	}

	out, err := IntersectBed(abed, opts, bfile.Name())
	if err != nil {
		os.Remove(bfile.Name())
		return nil, err
	}

	out2 := make(chan BedEntry, 256)
	go func() {
		defer close(out2)
		defer os.Remove(bfile.Name())
		for entry := range out {
			out2 <- entry
		}
	}()

	return out2, nil
}

package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var problemReport = flag.String("path", "", "Path to problem report")

// reader is an io.Reader that can read base64 compressed data stored as lines
// prefixed with a whitespace
type reader struct {
	r    *bufio.Reader
	data []byte
}

// formatReader is a io.Reader that determines the compression format used
// by the problem report file
type formatReader struct {
	r *bufio.Reader
	z io.ReadCloser
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Extracted core dump is saved as CoreDump.core.\n")
	}
}

func main() {
	flag.Parse()

	if *problemReport == "" {
		flag.Usage()
		os.Exit(1)
	}

	var f *os.File
	var err error
	var r io.ReadCloser
	var data []byte

	if f, err = os.OpenFile(*problemReport, os.O_RDONLY, 0666); err != nil {
		log.Fatalf(`unable to open "%s": %s`, *problemReport, err)
	}

	rdr := bufio.NewReader(f)

	// Skip everything to the line prefixed with `CoreDump:`
	for {
		if data, err = rdr.ReadBytes('\n'); err != nil {
			break
		}
		if bytes.HasPrefix(data, []byte("CoreDump:")) {
			break
		}
	}

	if err != nil {
		log.Fatal("unable to read: %s", err)
	}

	if r, err = newReader(base64.NewDecoder(base64.StdEncoding, &reader{r: rdr})); err != nil {
		log.Fatalf("unable to create reader: %s", err)
	}
	defer r.Close()

	if out, err := os.Create("CoreDump.core"); err != nil {
		log.Fatalf("unable to create output file: %s", err)
	} else {
		if _, err = io.Copy(out, r); err != nil {
			log.Fatalf("unable to save file: %s", err)
		}
	}
}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(r.data) > 0 {
		// Read from the scratch buffer
		n = copy(b, r.data)
		r.data = r.data[n:]
		return
	}
	if len(b) == 0 {
		return 0, nil
	}

	if r.data, err = r.r.ReadBytes('\n'); err != nil {
		return 0, err
	}

	if len(r.data) == 0 || r.data[0] != ' ' {
		return 0, io.EOF
	}

	n = copy(b, r.data[1:])
	r.data = r.data[n+1:]

	return n, nil
}

func newReader(r io.Reader) (io.ReadCloser, error) {
	return &formatReader{r: bufio.NewReader(r)}, nil
}

func (r *formatReader) Read(b []byte) (n int, err error) {
	if r.z == nil {
		var b []byte

		b, err = r.r.Peek(3)

		if bytes.Equal(b, []byte{0x1f, 0x8b, 0x8}) {
			// gzip
			r.z, err = gzip.NewReader(r.r)
		} else {
			// legacy zlib-only format
			r.z, err = zlib.NewReader(r.r)
		}
		if err != nil {
			return 0, err
		}
	}
	return r.z.Read(b)
}

func (r *formatReader) Close() error {
	return r.z.Close()
}

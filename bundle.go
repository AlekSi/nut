package nut

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

type Bundle struct {
	Dependencies *Dependencies
}

const (
	BundleFileName = "nut-bundle.json"
)

// check interface
var (
	_ io.ReaderFrom = &Bundle{}
	_ io.WriterTo   = &Bundle{}
)

// Reads bundle from specified file.
func (bundle *Bundle) ReadFile(fileName string) (err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = bundle.ReadFrom(f)
	return
}

// ReadFrom reads bundle from r until EOF.
// The return value n is the number of bytes read.
// Any error except io.EOF encountered during the read is also returned.
// Implements io.ReaderFrom.
func (bundle *Bundle) ReadFrom(r io.Reader) (n int64, err error) {
	var b []byte
	b, err = ioutil.ReadAll(r)
	n = int64(len(b))
	if err != nil {
		return
	}

	err = json.Unmarshal(b, bundle)
	return
}

// WriteTo writes bundle to w.
// The return value n is the number of bytes written.
// Any error encountered during the write is also returned.
// Implements io.WriterTo.
func (bundle *Bundle) WriteTo(w io.Writer) (n int64, err error) {
	var b []byte
	b, err = json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return
	}

	b = append(b, '\n')
	n1, err := w.Write(b)
	n = int64(n1)
	return
}

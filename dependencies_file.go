package nut

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

type DependenciesFile struct {
	Dependencies Dependencies
}

const (
	DependenciesFileName = "dependencies.json"
)

// check interfaces
var (
	_ io.ReaderFrom = &DependenciesFile{}
	_ io.WriterTo   = &DependenciesFile{}
)

// Reads dependencies from specified file.
func (deps *DependenciesFile) ReadFile(fileName string) (err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = deps.ReadFrom(f)
	return
}

// ReadFrom reads dependencies from r until EOF.
// The return value n is the number of bytes read.
// Any error except io.EOF encountered during the read is also returned.
// Implements io.ReaderFrom.
func (deps *DependenciesFile) ReadFrom(r io.Reader) (n int64, err error) {
	var b []byte
	b, err = ioutil.ReadAll(r)
	n = int64(len(b))
	if err != nil {
		return
	}

	err = json.Unmarshal(b, deps)
	return
}

// Writes dependencies to specified file.
func (deps *DependenciesFile) WriteFile(fileName string) (err error) {
	f, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = deps.WriteTo(f)
	return
}

// WriteTo writes dependencies to w.
// The return value n is the number of bytes written.
// Any error encountered during the write is also returned.
// Implements io.WriterTo.
func (deps *DependenciesFile) WriteTo(w io.Writer) (n int64, err error) {
	var b []byte
	b, err = json.MarshalIndent(deps, "", "  ")
	if err != nil {
		return
	}

	b = append(b, '\n')
	n1, err := w.Write(b)
	n = int64(n1)
	return
}

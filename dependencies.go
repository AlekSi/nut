package nut

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type AddDependencyError struct {
	ex, add *Dependency
}

// check interface
var (
	_ error = &AddDependencyError{}
)

func (e *AddDependencyError) Error() string {
	return fmt.Sprintf("Can't add %s to existing dependecy %s", e.add, e.ex)
}

type Dependencies struct {
	d map[string]Dependency // import path to dependency
}

// check interfaces
var (
	_ fmt.Stringer     = Dependencies{}
	_ json.Marshaler   = &Dependencies{}
	_ json.Unmarshaler = &Dependencies{}
)

func (deps Dependencies) String() string {
	paths := deps.ImportPaths()
	s := make([]string, len(paths))
	for i, path := range paths {
		s[i] = deps.Get(path).String()
	}
	return strings.Join(s, ", ")
}

func (deps *Dependencies) Clear() {
	deps.d = make(map[string]Dependency)
}

func (deps *Dependencies) Add(d *Dependency) (err error) {
	e := deps.Get(d.ImportPath)
	if e == nil {
		deps.d[d.ImportPath] = *d
		return
	}

	e.parse()
	e.parsed.majorMin = maxSection(e.parsed.majorMin, d.MajorMin())
	e.parsed.majorMax = minSection(e.parsed.majorMax, d.MajorMax())
	e.parsed.minorMin = maxSection(e.parsed.minorMin, d.MinorMin())
	e.parsed.minorMax = minSection(e.parsed.minorMax, d.MinorMax())
	e.parsed.patchMin = maxSection(e.parsed.patchMin, d.PatchMin())
	e.parsed.patchMax = minSection(e.parsed.patchMax, d.PatchMax())

	if !e.parsed.valid() {
		err = &AddDependencyError{ex: deps.Get(d.ImportPath), add: d}
		return
	}

	e.Version = e.parsed.String()
	deps.d[d.ImportPath] = *e
	return
}

func (deps *Dependencies) Get(importPath string) (dep *Dependency) {
	if deps.d == nil {
		deps.Clear()
	}
	d, ok := deps.d[importPath]
	if ok {
		dep = &d
	}
	return
}

func (deps *Dependencies) ImportPaths() (paths []string) {
	paths = make([]string, 0, len(deps.d))
	for i := range deps.d {
		paths = append(paths, i)
	}
	sort.Strings(paths)
	return
}

func (deps *Dependencies) MarshalJSON() (b []byte, err error) {
	b = make([]byte, 0, 50*len(deps.d))
	b = append(b, '{')
	paths := deps.ImportPaths()
	for i, p := range paths {
		if i > 0 {
			b = append(b, ',')
		}
		d := deps.Get(p)
		b = append(b, '"')
		b = append(b, d.ImportPath...)
		b = append(b, `":"`...)
		b = append(b, d.Version...)
		b = append(b, '"')
	}
	b = append(b, '}')
	return
}

func (deps *Dependencies) UnmarshalJSON(b []byte) (err error) {
	m := make(map[string]string)
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	deps.Clear() // also make
	var d *Dependency
	for p, v := range m {
		d, err = NewDependency(p, v)
		if err != nil {
			return
		}
		err = deps.Add(d)
		if err != nil {
			return
		}
	}
	return
}

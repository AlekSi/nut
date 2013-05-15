package main

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"

	. "github.com/AlekSi/nut"
)

var (
	cmdLock = &command{
		Run:       runLock,
		UsageLine: "lock",
		Short:     "lock",
	}
)

func init() {
	cmdLock.Long = `
Generates or updates dependencies.json in current directory.

Examples:
    nut lock
`
}

func runLock(cmd *command) {
	dir, err := os.Getwd()
	fatalIfErr(err)

	var importPaths []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") { // skip .git, .hg, etc
			return filepath.SkipDir
		}

		pack, err := build.ImportDir(path, 0)
		if _, ok := err.(*build.NoGoError); ok {
			return nil
		}
		if err != nil {
			return err
		}
		importPaths = append(importPaths, pack.Imports...)
		return nil
	})
	fatalIfErr(err)

	deps := DependenciesFile{}
	imported := make(map[string]bool)
	var path string
	for len(importPaths) > 0 {
		path, importPaths = importPaths[0], importPaths[1:]
		if imported[path] {
			continue
		}
		pack, err := build.Import(path, srcDir, 0)
		fatalIfErr(err)
		imported[path] = true
		if pack.Goroot {
			continue
		}

		vcs, rev, _ := vcsCurrent(filepath.Join(srcDir, pack.ImportPath), true)
		version := vcs + ":" + rev
		if vcs == "" {
			version = "*.*.*"
		}
		dep, err := NewDependency(pack.ImportPath, version)
		fatalIfErr(err)
		err = deps.Dependencies.Add(dep)
		fatalIfErr(err)

		importPaths = append(importPaths, pack.Imports...)
	}

	err = deps.WriteFile(DependenciesFileName)
	fatalIfErr(err)
}

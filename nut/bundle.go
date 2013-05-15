package main

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"

	. "github.com/AlekSi/nut"
)

var (
	cmdBundle = &command{
		Run:       runBundle,
		UsageLine: "bundle",
		Short:     "bundle",
	}
)

func init() {
	cmdBundle.Long = `
Generates or updates bundle nut-bundle.json in current directory.

Examples:
    nut bundle
`
}

func runBundle(cmd *command) {
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

	bundle := Bundle{}
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
		err = bundle.Dependencies.Add(dep)
		fatalIfErr(err)

		importPaths = append(importPaths, pack.Imports...)
	}

	err = bundle.WriteFile(BundleFileName)
	fatalIfErr(err)
}

package main

import (
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"

	. "github.com/AlekSi/nut"
)

var (
	cmdLock = &command{
		Run:       runLock,
		UsageLine: "lock [-v]",
		Short:     "lock",
	}

	lockV bool
)

func init() {
	cmdLock.Long = `
Generates or updates dependencies.json in current directory.

Examples:
    nut lock
`

	cmdLock.Flag.BoolVar(&lockV, "v", false, vHelp)
}

func runLock(cmd *command) {
	if !lockV {
		lockV = Config.V
	}

	// collect import paths
	var importPaths []string
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") { // skip .bzr, .git, .hg
			return filepath.SkipDir
		}
		pack, err := build.ImportDir(path, 0)
		if _, ok := err.(*build.NoGoError); ok {
			return nil
		}
		if err != nil {
			return err
		}
		importPaths = append(importPaths, pack.ImportPath)
		return nil
	})
	fatalIfErr(err)

	deps := Dependencies{}
	imported := make(map[string]bool)
	var path string
	for len(importPaths) > 0 {
		path, importPaths = importPaths[0], importPaths[1:]
		if imported[path] {
			continue
		}
		if strings.HasPrefix(path, ".") || strings.HasPrefix(path, "/") {
			log.Printf("Warning: Skipping import path %q", path)
			imported[path] = true
			continue
		}
		pack, err := build.Import(path, srcDir, 0)
		fatalIfErr(err)
		imported[path] = true
		if pack.Goroot {
			continue
		}

		vcs, root := vcsRoot(filepath.Join(srcDir, pack.ImportPath))
		if root == "" {
			continue
		}
		if imported[root] {
			continue
		}
		imported[root] = true
		rev := vcsCurrent(vcs, root, lockV)
		if rev == "" {
			continue
		}
		root, err = filepath.Rel(srcDir, root)
		fatalIfErr(err)
		dep, err := NewDependency(root, vcs+":"+rev)
		fatalIfErr(err)
		err = deps.Add(dep)
		fatalIfErr(err)
	}

	err = deps.WriteFile(DependenciesFileName)
	fatalIfErr(err)
}

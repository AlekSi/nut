package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

func vcsRoot(dir string) (vcs, root string) {
	if !filepath.IsAbs(dir) {
		panic(fmt.Errorf("%q should be absolute path", dir))
	}

	for dir != srcDir {
		fi, err := ioutil.ReadDir(dir)
		fatalIfErr(err)
		for _, f := range fi {
			switch f.Name() {
			case ".bzr", ".git", ".hg":
				vcs = f.Name()[1:]
				root = dir
				return
			}
		}

		dir = filepath.Join(dir, "..")
	}

	return
}

func vcsCheckout(vcs, rev, dir string, verbose bool) {
	var args string
	switch vcs {
	case "bzr":
		args = "update -r " + rev
	case "git":
		args = "checkout -q " + rev
	case "hg":
		args = "update -c -r " + rev
	default:
		log.Fatalf("VCS %s is not supported.", vcs)
	}

	c := exec.Command(vcs, strings.Split(args, " ")...)
	c.Dir = dir
	if verbose {
		log.Printf("Running %q (in %q)", strings.Join(c.Args, " "), dir)
	}
	out, err := c.CombinedOutput()
	if Config.Debug || err != nil {
		log.Print(string(out))
	}
	fatalIfErr(err)
}

func vcsCurrent(vcs, root string, verbose bool) (rev string) {
	args := map[string]string{
		"bzr": "testament",
		"git": "rev-parse --verify HEAD",
		"hg":  "identify --debug -i",
	}[vcs]

	c := exec.Command(vcs, strings.Split(args, " ")...)
	c.Dir = root
	if verbose {
		log.Printf("Running %q (in %q)", strings.Join(c.Args, " "), c.Dir)
	}
	out, err := c.CombinedOutput()
	if Config.Debug || err != nil {
		log.Print(string(out))
	}
	fatalIfErr(err)

	rev = strings.TrimSpace(string(out))
	switch vcs {
	case "bzr":
		for _, s := range strings.Split(rev, "\n") {
			if strings.HasPrefix(s, "revision-id: ") {
				rev = strings.SplitN(s, " ", 2)[1]
				break
			}
		}
	case "hg":
		if strings.HasSuffix(rev, "+") {
			rev = rev[:len(rev)-1]
		}
	}

	return
}

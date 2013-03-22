package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func vcsCheckout(vcs, rev, dir string, verbose bool) {
	var args string
	switch vcs {
	case "git":
		args = "checkout -q " + rev
	case "hg":
		args = "update -c -r " + rev
	case "bzr":
		args = "update -r " + rev
	default:
		log.Fatalf("VCS %s is not supported.", vcs)
	}

	c := exec.Command(vcs, strings.Split(args, " ")...)
	c.Dir = dir
	if verbose {
		log.Printf("Running %q (in %q)", strings.Join(c.Args, " "), dir)
	}
	out, err := c.CombinedOutput()
	if verbose || err != nil {
		log.Print(string(out))
	}
	fatalIfErr(err)
}

func vcsCurrent(dir string, verbose bool) (vcs, rev string) {
	var args string
	for v, a := range map[string]string{
		"git": "rev-parse --verify HEAD",
		"hg":  "identify --debug -i",
		"bzr": "testament",
	} {
		_, err := os.Stat(filepath.Join(dir, "."+v))
		if err == nil {
			vcs = v
			args = a
			break
		}
	}
	if vcs == "" {
		log.Fatalf("%s: unknown or unsupported VCS.", dir)
	}

	c := exec.Command(vcs, strings.Split(args, " ")...)
	c.Dir = dir
	if verbose {
		log.Printf("Running %q (in %q)", strings.Join(c.Args, " "), dir)
	}
	out, err := c.CombinedOutput()
	if verbose || err != nil {
		log.Print(string(out))
	}
	fatalIfErr(err)

	rev = strings.TrimSpace(string(out))
	switch vcs {
	case "hg":
		if strings.HasSuffix(rev, "+") {
			rev = rev[:len(rev)-1]
		}
	case "bzr":
		for _, s := range strings.Split(rev, "\n") {
			if strings.HasPrefix(s, "revision-id: ") {
				rev = strings.SplitN(s, " ", 2)[1]
				break
			}
		}
	}

	return
}

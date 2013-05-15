package main

import (
	"log"
	"os/exec"
	"strings"
)

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
	if verbose || err != nil {
		log.Print(string(out))
	}
	fatalIfErr(err)
}

func vcsCurrent(dir string, verbose bool) (vcs, rev, root string) {
	// detect vcs and repository root
	for v, a := range map[string]string{
		// FIXME "bzr": "",
		"git": "rev-parse --show-toplevel",
		"hg":  "root",
	} {
		c := exec.Command(v, strings.Split(a, " ")...)
		c.Dir = dir
		if verbose {
			log.Printf("Running %q (in %q)", strings.Join(c.Args, " "), c.Dir)
		}
		out, err := c.CombinedOutput()
		if verbose {
			log.Print(string(out))
		}
		if err == nil {
			vcs = v
			root = strings.TrimSpace(string(out))
			break
		}
		if _, ok := err.(*exec.ExitError); ok {
			err = nil
		}
		fatalIfErr(err)
	}
	if vcs == "" {
		return
	}

	// detect current revision
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
	if verbose || err != nil {
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

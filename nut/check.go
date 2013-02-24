package main

import (
	"go/build"
	"log"
	"os"
	"strings"

	. "github.com/AlekSi/nut"
)

var (
	cmdCheck = &Command{
		Run:       runCheck,
		UsageLine: "check [-v] [filenames]",
		Short:     "check specs and nuts for errors",
	}

	checkV bool
)

func init() {
	cmdCheck.Long = `
Checks given spec (.json) or nut (.nut) files.
If no filenames are given, checks spec nut.json in current directory.

Examples:
    nut check
    nut check test_nut1/nut.json
    nut check test_nut1/test_nut1-0.0.1.nut
`

	cmdCheck.Flag.BoolVar(&checkV, "v", false, vHelp)
}

func runCheck(cmd *Command) {
	if !checkV {
		checkV = Config.V
	}

	args := cmd.Flag.Args()
	if len(args) == 0 {
		args = []string{SpecFileName}
	}

	for _, arg := range args {
		var errors []string

		parts := strings.Split(arg, ".")
		switch parts[len(parts)-1] {
		case "json":
			spec := new(Spec)
			err := spec.ReadFile(arg)
			FatalIfErr(err)
			pack, err := build.ImportDir(".", 0)
			FatalIfErr(err)
			errors = spec.Check()
			errors = append(errors, CheckPackage(pack)...)

		case "nut":
			_, nf := ReadNut(arg)
			errors = nf.Check()

		default:
			log.Fatalf("%q doesn't end with .json or .nut", arg)
		}

		if len(errors) != 0 {
			log.Printf("Found errors in %s:", arg)
			for _, e := range errors {
				log.Printf("    %s", e)
			}
			os.Exit(1)
		}

		if checkV {
			log.Printf("%s looks good.", arg)
		}
	}
}

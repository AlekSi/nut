package main

import (
	"log"
	"os"
)

var (
	cmdUnpack = &Command{
		Run:       runUnpack,
		UsageLine: "unpack [-nc] [-v] [filename]",
		Short:     "unpack nut into current directory",
	}

	unpackNC bool
	unpackV  bool
)

func init() {
	cmdUnpack.Long = `
Unpack nut into current directory.
	`

	cmdUnpack.Flag.BoolVar(&unpackNC, "nc", false, "no check (not recommended)")
	cmdUnpack.Flag.BoolVar(&unpackV, "v", false, vHelp)
}

func runUnpack(cmd *Command) {
	if !unpackV {
		unpackV = config.V
	}

	if len(cmd.Flag.Args()) != 1 {
		log.Fatalf("Expected exactly one filename, got %s", cmd.Flag.Args())
	}
	fileName := cmd.Flag.Args()[0]

	// check nut
	if !unpackNC {
		_, nf := ReadNut(fileName)
		errors := nf.Check()
		if len(errors) != 0 {
			log.Print("Found errors:")
			for _, e := range errors {
				log.Printf("    %s", e)
			}
			log.Fatalf("Please contact nut author.")
		}
	}

	// unpack nut
	dir, err := os.Getwd()
	PanicIfErr(err)
	UnpackNut(fileName, dir, false, unpackV)
	if unpackV {
		log.Printf("%s unpacked.", fileName)
	}
}

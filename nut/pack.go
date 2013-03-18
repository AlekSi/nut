package main

import (
	"go/build"
	"log"

	. "github.com/AlekSi/nut"
)

const (
	NutFilePerm = 0644
)

var (
	cmdPack = &Command{
		Run:       runPack,
		UsageLine: "pack [-nc] [-o filename] [-v]",
		Short:     "pack package in current directory into nut",
	}

	packNC bool
	packO  string
	packV  bool
)

func init() {
	cmdPack.Long = `
Packs package in current directory into nut.

Examples:
    nut pack
`

	cmdPack.Flag.BoolVar(&packNC, "nc", false, "no check (not recommended)")
	cmdPack.Flag.StringVar(&packO, "o", "", "output filename")
	cmdPack.Flag.BoolVar(&packV, "v", false, vHelp)
}

func runPack(cmd *Command) {
	if !packV {
		packV = Config.V
	}

	if len(cmd.Flag.Args()) != 0 {
		log.Fatal("This command does not accept arguments.")
	}

	/*
		packages := build.ImportDir(".", 0)
		if len(packages) > 1 {
			panic(fmt.Errorf("Multi-package nuts are not supported yet\n%#v", packages))
			// implementation will require import overwrites on install to prevent error like
			// "local import "./dir/subdir" in non-local package"
		}
	*/

	ctxt := build.Default
	ctxt.UseAllFiles = true
	pack, err := ctxt.ImportDir(".", 0)
	FatalIfErr(err)

	if pack.Name == "main" {
		log.Fatal(`Binaries (package "main") are not supported yet.`)
	}

	var fileName string
	spec := new(Spec)
	err = spec.ReadFile(SpecFileName)
	FatalIfErr(err)
	nut := Nut{Spec: *spec, Package: *pack}
	if packO == "" {
		fileName = nut.FileName()
	} else {
		fileName = packO
	}

	if !packNC {
		errors := nut.Check()
		if len(errors) != 0 {
			log.Print("Found errors:")
			for _, e := range errors {
				log.Printf("    %s", e)
			}
			log.Fatal("Hint: use 'nut check'.")
		}
	}

	var files []string
	files = append(files, pack.GoFiles...)
	files = append(files, pack.CgoFiles...)
	files = append(files, pack.TestGoFiles...)
	files = append(files, pack.XTestGoFiles...)
	files = append(files, spec.ExtraFiles...)
	files = append(files, SpecFileName)

	PackNut(fileName, files, packV)
	if packV {
		log.Printf("%s created.", fileName)
	}
}

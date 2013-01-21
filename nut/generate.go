package main

import (
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"

	. "github.com/AlekSi/nut"
)

const (
	SpecFilePerm = 0644
)

var (
	cmdGenerate = &Command{
		Run:       runGenerate,
		UsageLine: "generate [-v]",
		Short:     "generate or update spec for package in current directory",
	}

	generateV bool
)

func init() {
	cmdGenerate.Long = `
Generates or updates spec nut.json for package in current directory.
	`

	cmdGenerate.Flag.BoolVar(&generateV, "v", false, vHelp)
}

func runGenerate(cmd *Command) {
	if !generateV {
		generateV = config.V
	}

	action := "updated"
	var err error
	var spec *Spec

	// read spec
	if _, err = os.Stat(SpecFileName); os.IsNotExist(err) {
		action = "generated"
		spec = new(Spec)
	} else {
		spec = ReadSpec(SpecFileName)
	}

	// read package
	pack, err := build.ImportDir(".", 0)
	FatalIfErr(err)

	// add example author
	if len(spec.Authors) == 0 {
		spec.Authors = []Person{{FullName: ExampleFullName, Email: ExampleEmail}}
	}

	// some extra files
	if len(spec.ExtraFiles) == 0 {
		var globs []string
		for _, glob := range []string{"read*", "licen?e*", "copying*", "contrib*", "author*",
			"thank*", "news*", "change*", "install*", "bug*", "todo*"} {
			globs = append(globs, glob, strings.ToUpper(glob), strings.Title(glob))
		}

		for _, glob := range globs {
			files, err := filepath.Glob(glob)
			PanicIfErr(err)
			spec.ExtraFiles = append(spec.ExtraFiles, files...)
		}
	}

	// write spec
	f, err := os.OpenFile(SpecFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, SpecFilePerm)
	PanicIfErr(err)
	defer f.Close()
	_, err = spec.WriteTo(f)
	PanicIfErr(err)

	if generateV {
		log.Printf("%s %s.", SpecFileName, action)
	}

	// check spec and package
	errors := spec.Check()
	errors = append(errors, CheckPackage(pack)...)
	if len(errors) != 0 {
		log.Print("\nNow you should edit nut.json to fix following errors:")
		for _, e := range errors {
			log.Printf("    %s", e)
		}
		log.Print("\nAfter that run 'nut check' to check spec again.")
	}
}

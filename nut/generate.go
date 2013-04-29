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
	specFilePerm = 0644
)

var (
	cmdGenerate = &command{
		Run:       runGenerate,
		UsageLine: "generate [-v]",
		Short:     "generate or update spec for package in current directory",
	}

	generateV bool
)

func init() {
	cmdGenerate.Long = `
Generates or updates spec nut.json for package in current directory.

Examples:
    nut generate
`

	cmdGenerate.Flag.BoolVar(&generateV, "v", false, vHelp)
}

func runGenerate(cmd *command) {
	if !generateV {
		generateV = Config.V
	}

	if len(cmd.Flag.Args()) != 0 {
		log.Fatal("This command does not accept arguments.")
	}

	action := "updated"
	var err error
	var spec *Spec

	// read spec
	spec = new(Spec)
	if _, err = os.Stat(SpecFileName); os.IsNotExist(err) {
		action = "generated"
		spec.Vendor = Config.Vendor
	} else {
		err = spec.ReadFile(SpecFileName)
		fatalIfErr(err)
	}

	// read package
	pack, err := build.ImportDir(".", 0)
	fatalIfErr(err)

	// add example author
	if len(spec.Authors) == 0 {
		spec.Authors = []Person{{FullName: Config.FullName, Email: Config.Email}}
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
			fatalIfErr(err)
			spec.ExtraFiles = append(spec.ExtraFiles, files...)
		}
	}

	// write spec
	f, err := os.OpenFile(SpecFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, specFilePerm)
	fatalIfErr(err)
	defer f.Close()
	_, err = spec.WriteTo(f)
	fatalIfErr(err)

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

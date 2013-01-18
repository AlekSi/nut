package main

import (
	"log"
	"path/filepath"
)

const (
	WorkspaceDirPerm = 0755
)

var (
	cmdInstall = &Command{
		Run:       runInstall,
		UsageLine: "install [-nc] [-p prefix] [-v] [filenames]",
		Short:     "unpack nut and install package",
	}

	installNC bool
	installP  string
	installV  bool
)

func init() {
	cmdInstall.Long = `
Copies nuts into GOPATH/nut/<prefix>, unpacks them into GOPATH/src/<prefix>/<name>/<version>
and installs using 'go install'.
	`

	cmdInstall.Flag.BoolVar(&installNC, "nc", false, "no check (not recommended)")
	cmdInstall.Flag.StringVar(&installP, "p", "localhost", "install prefix in workspace")
	cmdInstall.Flag.BoolVar(&installV, "v", false, vHelp)
}

func runInstall(cmd *Command) {
	if !installV {
		installV = config.V
	}

	for _, arg := range cmd.Flag.Args() {
		_, nf := ReadNut(arg)

		if nf.Name == "main" {
			log.Fatal(`Binaries (package "main") are not supported yet.`)
		}

		// check nut
		if !installNC {
			errors := nf.Check()
			if len(errors) != 0 {
				log.Print("Found errors:")
				for _, e := range errors {
					log.Printf("    %s", e)
				}
				log.Fatalf("Please contact nut author.")
			}
		}

		CopyNut(arg, installP, installV)
		path := filepath.Join(installP, nf.Name, nf.Version.String())
		UnpackNut(arg, filepath.Join(SrcDir, path), true, installV)
		InstallPackage(path, installV)
	}
}

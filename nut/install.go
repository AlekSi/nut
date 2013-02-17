package main

import (
	"io/ioutil"
	"log"
	"os"
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
Copies nuts into GOPATH/nut/<prefix>/<vendor>/<name>-<version>.nut,
unpacks them into GOPATH/src/<prefix>/<vendor>/<name> and
installs using 'go install'.

Examples:
    nut install test_nut1-0.0.1.nut
    nut install -p gonuts.io test_nut1-0.0.1.nut
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
		b, nf := ReadNut(arg)

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
				log.Fatal("Please contact nut author.")
			}
		}

		// copy nut
		dstFile := filepath.Join(NutDir, nf.FilePath(installP))
		if installV {
			log.Printf("Copying %s to %s ...", arg, dstFile)
		}
		PanicIfErr(os.MkdirAll(filepath.Dir(dstFile), WorkspaceDirPerm))
		PanicIfErr(ioutil.WriteFile(dstFile, b, NutFilePerm))

		srcPath := filepath.Join(SrcDir, nf.ImportPath(installP))
		if installV {
			log.Printf("Unpacking into %s ...", srcPath)
		}
		UnpackNut(dstFile, srcPath, true, installV)

		InstallPackage(nf.ImportPath(installP), installV)
	}
}

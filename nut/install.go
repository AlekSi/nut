package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	workspaceDirPerm = 0755
)

var (
	cmdInstall = &command{
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

func runInstall(cmd *command) {
	if !installV {
		installV = Config.V
	}

	for _, arg := range cmd.Flag.Args() {
		b, nf := readNut(arg)

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
		dstFile := filepath.Join(nutDir, nf.FilePath(installP))
		if installV {
			log.Printf("Copying %s to %s ...", arg, dstFile)
		}
		fatalIfErr(os.MkdirAll(filepath.Dir(dstFile), workspaceDirPerm))
		fatalIfErr(ioutil.WriteFile(dstFile, b, nutFilePerm))

		srcPath := filepath.Join(srcDir, nf.ImportPath(installP))
		if installV {
			log.Printf("Unpacking into %s ...", srcPath)
		}
		unpackNut(dstFile, srcPath, true, installV)

		installPackage(nf.ImportPath(installP), installV)
	}
}

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	. "github.com/AlekSi/nut"
)

type Config struct {
	Token string
	V     bool
}

const (
	ConfigFileName = ".nut.json"
	ConfigFilePerm = 0644
	DefaultServer  = "www.gonuts.io"
)

var (
	WorkspaceDir string // current workspace (first path in GOPATH)
	SrcDir       string // src directory in current workspace
	NutDir       string // nut directory in current workspace

	config       Config
	gonutsServer string
	vHelp        string = fmt.Sprintf("be verbose, may be read from ~/%s", ConfigFileName)
)

func init() {
	log.SetFlags(0)

	srcDirs := build.Default.SrcDirs()[1:]
	if len(srcDirs) == 0 {
		env := os.Getenv("GOPATH")
		if env == "" {
			log.Print("GOPATH environment variable is empty.")
		} else {
			log.Printf("Workspaces in GOPATH environment variable (%s), or their src subpaths don't exist.", env)
		}
		log.Fatal("Setup a workspace as described there: http://golang.org/doc/code.html")
	}

	SrcDir = srcDirs[0]
	WorkspaceDir = filepath.Join(SrcDir, "..")
	NutDir = filepath.Join(WorkspaceDir, "nut")

	u, err := user.Current()
	if err != nil {
		_, err = os.Stat(u.HomeDir)
	}
	if err != nil {
		log.Printf("Warning: Can't detect current user home directory: %s", err)
		return
	}

	path := filepath.Join(u.HomeDir, ConfigFileName)
	b, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(b, &config)
	}
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Can't load config from %s: %s\n", path, err)
		config = Config{}
	}

	if !os.IsNotExist(err) {
		b, err = json.MarshalIndent(config, "", "  ")
		if err == nil {
			err = ioutil.WriteFile(path, b, ConfigFilePerm)
		}
		if err != nil {
			log.Printf("Warning: Can't write config to %s: %s\n", path, err)
		}
	}

	gonutsServer = os.Getenv("GONUTS_SERVER")
	if gonutsServer == "" {
		gonutsServer = DefaultServer
	}
}

func PanicIfErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// TODO common functions there are mess for now

// Read spec file.
func ReadSpec(fileName string) (spec *Spec) {
	f, err := os.Open(fileName)
	PanicIfErr(err)
	defer f.Close()
	spec = new(Spec)
	_, err = spec.ReadFrom(f)
	PanicIfErr(err)
	return
}

// Read nut file.
func ReadNut(fileName string) (b []byte, nf *NutFile) {
	var err error
	b, err = ioutil.ReadFile(fileName)
	PanicIfErr(err)
	nf = new(NutFile)
	_, err = nf.ReadFrom(bytes.NewReader(b))
	PanicIfErr(err)
	return
}

// Write nut to GOPATH/nut/<prefix>/<name>-<version>.nut
func WriteNut(b []byte, prefix string, verbose bool) string {
	nf := new(NutFile)
	_, err := nf.ReadFrom(bytes.NewReader(b))
	PanicIfErr(err)

	// create GOPATH/nut/<prefix>
	dir := filepath.Join(NutDir, prefix)
	PanicIfErr(os.MkdirAll(dir, WorkspaceDirPerm))

	// write file
	dstFilepath := filepath.Join(dir, nf.FileName())
	if verbose {
		log.Printf("Writing %s ...", dstFilepath)
	}
	PanicIfErr(ioutil.WriteFile(dstFilepath, b, NutFilePerm))
	return dstFilepath
}

// Copy nut to GOPATH/nut/<prefix>/<name>-<version>.nut
func CopyNut(nutFilepath string, prefix string, verbose bool) {
	b, nf := ReadNut(nutFilepath)

	// create GOPATH/nut/<prefix>
	dir := filepath.Join(NutDir, prefix)
	PanicIfErr(os.MkdirAll(dir, WorkspaceDirPerm))

	// write file
	dstFilepath := filepath.Join(dir, nf.FileName())
	if verbose {
		log.Printf("Copying %s to %s ...", nutFilepath, dstFilepath)
	}
	PanicIfErr(ioutil.WriteFile(dstFilepath, b, NutFilePerm))
}

// Pack files into nut file with given fileName.
func PackNut(fileName string, files []string, verbose bool) {
	// write nut to temporary file first
	nutFile, err := ioutil.TempFile("", "nut-")
	PanicIfErr(err)
	defer func() {
		if nutFile != nil {
			PanicIfErr(os.Remove(nutFile.Name()))
		}
	}()

	nutWriter := zip.NewWriter(nutFile)
	defer func() {
		if nutWriter != nil {
			PanicIfErr(nutWriter.Close())
		}
	}()

	// add files to nut with all meta information
	for _, file := range files {
		if verbose {
			log.Printf("Packing %s ...", file)
		}

		fi, err := os.Stat(file)
		PanicIfErr(err)

		fh, err := zip.FileInfoHeader(fi)
		PanicIfErr(err)
		fh.Name = file

		f, err := nutWriter.CreateHeader(fh)
		PanicIfErr(err)

		b, err := ioutil.ReadFile(file)
		PanicIfErr(err)

		_, err = f.Write(b)
		PanicIfErr(err)
	}

	err = nutWriter.Close()
	nutWriter = nil
	PanicIfErr(err)

	PanicIfErr(nutFile.Close())

	// move file to specified location and fix permissions
	if verbose {
		log.Printf("Creating %s ...", fileName)
	}
	PanicIfErr(os.Rename(nutFile.Name(), fileName))
	nutFile = nil
	PanicIfErr(os.Chmod(fileName, NutFilePerm))
}

// Unpack nut file with given fileName into dir. Creates dir if needed. Removes dir first if asked.
func UnpackNut(fileName string, dir string, removeDir, verbose bool) {
	// check dir
	_, err := os.Stat(dir)
	if err == nil && removeDir {
		if verbose {
			log.Printf("Removing existing directory %s ...", dir)
		}
		os.RemoveAll(dir)
	}
	PanicIfErr(os.MkdirAll(dir, WorkspaceDirPerm))

	_, nf := ReadNut(fileName)

	for _, file := range nf.Reader.File {
		if verbose {
			log.Printf("Unpacking %s ...", file.Name)
		}

		rc, err := file.Open()
		PanicIfErr(err)
		defer rc.Close()

		b, err := ioutil.ReadAll(rc)
		PanicIfErr(err)

		PanicIfErr(ioutil.WriteFile(filepath.Join(dir, file.Name), b, file.Mode()))
	}
}

// Call 'go install <path>'.
func InstallPackage(path string, verbose bool) {
	args := []string{"install"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, path)
	c := exec.Command("go", args...)
	if verbose {
		log.Printf("Running %q", strings.Join(c.Args, " "))
	}
	out, err := c.CombinedOutput()
	if verbose || err != nil {
		log.Print(string(out))
	}
	if err != nil {
		log.Fatal(err)
	}
}

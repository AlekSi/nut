package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	. "github.com/AlekSi/nut"
)

type ConfigFile struct {
	Token string
	V     bool
	Debug bool
}

const (
	ConfigFileName = ".nut.json"
	ConfigFilePerm = 0644
)

var (
	WorkspaceDir string // current workspace (first path in GOPATH)
	SrcDir       string // src directory in current workspace
	NutDir       string // nut directory in current workspace

	// Maps import prefixes to hosts serving nuts.
	// Three reasons for it:
	//   - third-party nut servers (TODO to be implemented);
	//   - testing with dev_appserver;
	//   - no GAE for second-level domains.
	NutImportPrefixes = map[string]string{"gonuts.io": "www.gonuts.io"}

	Config ConfigFile
	vHelp  string = fmt.Sprintf("be verbose (may be read from ~/%s)", ConfigFileName)
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
		log.Fatal("Setup a workspace (GOPATH) as described there: http://golang.org/doc/code.html")
	}

	SrcDir = srcDirs[0]
	WorkspaceDir = filepath.Join(SrcDir, "..")
	NutDir = filepath.Join(WorkspaceDir, "nut")

	// detect home dir
	u, err := user.Current()
	if err == nil {
		_, err = os.Stat(u.HomeDir)
	}
	if err != nil {
		log.Printf("Warning: Can't detect current user home directory: %s", err)
		return
	}

	// load config if file exists
	path := filepath.Join(u.HomeDir, ConfigFileName)
	b, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(b, &Config)
	}
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Can't load config from %s: %s\n", path, err)
		Config = ConfigFile{}
	}

	// write config if file exists
	if !os.IsNotExist(err) {
		b, err = json.MarshalIndent(Config, "", "  ")
		if err == nil {
			err = ioutil.WriteFile(path, b, ConfigFilePerm)
		}
		if err != nil {
			log.Printf("Warning: Can't write config to %s: %s\n", path, err)
		}
	}

	// set logger flags
	if Config.Debug {
		log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	}

	// for development
	env := os.Getenv("GONUTS_IO_SERVER")
	if env != "" {
		u, err := url.Parse(env)
		FatalIfErr(err)
		NutImportPrefixes["gonuts.io"] = u.Host
	}
}

func FatalIfErr(err error) {
	if err != nil {
		if Config.Debug {
			// show full backtraces
			log.Panic(err)
		} else {
			log.Fatal(err)
		}
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
	FatalIfErr(err)
}

// TODO common functions below are mess for now

// Read nut file.
func ReadNut(fileName string) (b []byte, nf *NutFile) {
	var err error
	b, err = ioutil.ReadFile(fileName)
	FatalIfErr(err)
	nf = new(NutFile)
	_, err = nf.ReadFrom(bytes.NewReader(b))
	FatalIfErr(err)
	return
}

// Write nut to GOPATH/nut/<prefix>/<name>-<version>.nut
func WriteNut(b []byte, prefix string, verbose bool) string {
	nf := new(NutFile)
	_, err := nf.ReadFrom(bytes.NewReader(b))
	FatalIfErr(err)

	// create GOPATH/nut/<prefix>
	dir := filepath.Join(NutDir, prefix)
	FatalIfErr(os.MkdirAll(dir, WorkspaceDirPerm))

	// write file
	dstFilepath := filepath.Join(dir, nf.FileName())
	if verbose {
		log.Printf("Writing %s ...", dstFilepath)
	}
	FatalIfErr(ioutil.WriteFile(dstFilepath, b, NutFilePerm))
	return dstFilepath
}

// Pack files into nut file with given fileName.
func PackNut(fileName string, files []string, verbose bool) {
	// write nut to temporary file first
	nutFile, err := ioutil.TempFile("", "nut-")
	FatalIfErr(err)
	defer func() {
		if nutFile != nil {
			FatalIfErr(os.Remove(nutFile.Name()))
		}
	}()

	nutWriter := zip.NewWriter(nutFile)
	defer func() {
		if nutWriter != nil {
			FatalIfErr(nutWriter.Close())
		}
	}()

	// add files to nut with all meta information
	for _, file := range files {
		if verbose {
			log.Printf("Packing %s ...", file)
		}

		fi, err := os.Stat(file)
		FatalIfErr(err)

		fh, err := zip.FileInfoHeader(fi)
		FatalIfErr(err)
		fh.Name = file

		f, err := nutWriter.CreateHeader(fh)
		FatalIfErr(err)

		b, err := ioutil.ReadFile(file)
		FatalIfErr(err)

		_, err = f.Write(b)
		FatalIfErr(err)
	}

	err = nutWriter.Close()
	nutWriter = nil
	FatalIfErr(err)

	FatalIfErr(nutFile.Close())

	// move file to specified location and fix permissions
	if verbose {
		log.Printf("Creating %s ...", fileName)
	}
	_, err = os.Stat(fileName)
	if err == nil {
		// required on Windows
		FatalIfErr(os.Remove(fileName))
	}
	FatalIfErr(os.Rename(nutFile.Name(), fileName))
	nutFile = nil
	FatalIfErr(os.Chmod(fileName, NutFilePerm))
}

// Unpack nut file with given fileName into dir, overwriting files.
// Creates dir if needed. Removes dir first if asked.
func UnpackNut(fileName string, dir string, removeDir, verbose bool) {
	// check dir
	_, err := os.Stat(dir)
	if err == nil && removeDir {
		if verbose {
			log.Printf("Removing existing directory %s ...", dir)
		}
		os.RemoveAll(dir)
	}
	FatalIfErr(os.MkdirAll(dir, WorkspaceDirPerm))

	nf := new(NutFile)
	err = nf.ReadFile(fileName)
	FatalIfErr(err)

	for _, file := range nf.Reader.File {
		if verbose {
			log.Printf("Unpacking %s ...", file.Name)
		}

		src, err := file.Open()
		FatalIfErr(err)

		dst, err := os.OpenFile(filepath.Join(dir, file.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		FatalIfErr(err)

		_, err = io.Copy(dst, src)
		FatalIfErr(err)

		src.Close()
		dst.Close()
	}
}

// Return imports present in NutImportPrefixes without altering them.
func NutImports(imports []string) (nuts []string) {
	for _, imp := range imports {
		p := strings.Split(imp, "/")
		if _, ok := NutImportPrefixes[p[0]]; ok {
			nuts = append(nuts, imp)
		}
	}
	return
}

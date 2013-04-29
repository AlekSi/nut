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

type configFile struct {
	Token    string
	Vendor   string
	FullName string
	Email    string
	V        bool
	Debug    bool
}

const (
	configFileName = ".nut.json"
	configFilePerm = 0644
	nutFilePerm    = 0644
)

var (
	// Maps import prefixes to hosts serving nuts.
	// Three reasons for it:
	//   - third-party nut servers (TODO to be implemented);
	//   - testing with dev_appserver;
	//   - no GAE for second-level domains.
	NutImportPrefixes = map[string]string{"gonuts.io": "www.gonuts.io"}

	Config configFile

	workspaceDir string // current workspace (first path in GOPATH)
	srcDir       string // src directory in current workspace
	nutDir       string // nut directory in current workspace

	vHelp string = fmt.Sprintf("be verbose (may be read from ~/%s)", configFileName)
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

	srcDir = srcDirs[0]
	workspaceDir = filepath.Join(srcDir, "..")
	nutDir = filepath.Join(workspaceDir, "nut")

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
	path := filepath.Join(u.HomeDir, configFileName)
	b, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(b, &Config)
	}
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Can't load config from %s: %s\n", path, err)
		Config = configFile{}
	}

	if Config.FullName == "" {
		Config.FullName = ExampleFullName
	}
	if Config.Email == "" {
		Config.Email = ExampleEmail
	}

	// write config if file exists
	if !os.IsNotExist(err) {
		b, err = json.MarshalIndent(Config, "", "  ")
		if err == nil {
			err = ioutil.WriteFile(path, b, configFilePerm)
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
		fatalIfErr(err)
		NutImportPrefixes["gonuts.io"] = u.Host
	}
}

func fatalIfErr(err error) {
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
func installPackage(path string, verbose bool) {
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
	fatalIfErr(err)
}

// TODO common functions below are mess for now

// Read nut file.
func readNut(fileName string) (b []byte, nf *NutFile) {
	var err error
	b, err = ioutil.ReadFile(fileName)
	fatalIfErr(err)
	nf = new(NutFile)
	_, err = nf.ReadFrom(bytes.NewReader(b))
	fatalIfErr(err)
	return
}

// Write nut to GOPATH/nut/<prefix>/<name>-<version>.nut
func writeNut(b []byte, prefix string, verbose bool) string {
	nf := new(NutFile)
	_, err := nf.ReadFrom(bytes.NewReader(b))
	fatalIfErr(err)

	// create GOPATH/nut/<prefix>
	dir := filepath.Join(nutDir, prefix)
	fatalIfErr(os.MkdirAll(dir, workspaceDirPerm))

	// write file
	dstFilepath := filepath.Join(dir, nf.FileName())
	if verbose {
		log.Printf("Writing %s ...", dstFilepath)
	}
	fatalIfErr(ioutil.WriteFile(dstFilepath, b, nutFilePerm))
	return dstFilepath
}

// Pack files into nut file with given fileName.
func packNut(fileName string, files []string, verbose bool) {
	// write nut to temporary file first
	nutFile, err := ioutil.TempFile("", "nut-")
	fatalIfErr(err)
	defer func() {
		if nutFile != nil {
			fatalIfErr(os.Remove(nutFile.Name()))
		}
	}()

	nutWriter := zip.NewWriter(nutFile)
	defer func() {
		if nutWriter != nil {
			fatalIfErr(nutWriter.Close())
		}
	}()

	// add files to nut with all meta information
	for _, file := range files {
		if verbose {
			log.Printf("Packing %s ...", file)
		}

		fi, err := os.Stat(file)
		fatalIfErr(err)

		fh, err := zip.FileInfoHeader(fi)
		fatalIfErr(err)
		fh.Name = file

		f, err := nutWriter.CreateHeader(fh)
		fatalIfErr(err)

		b, err := ioutil.ReadFile(file)
		fatalIfErr(err)

		_, err = f.Write(b)
		fatalIfErr(err)
	}

	err = nutWriter.Close()
	nutWriter = nil
	fatalIfErr(err)

	fatalIfErr(nutFile.Close())

	// move file to specified location and fix permissions
	if verbose {
		log.Printf("Creating %s ...", fileName)
	}
	_, err = os.Stat(fileName)
	if err == nil {
		// required on Windows
		fatalIfErr(os.Remove(fileName))
	}
	fatalIfErr(os.Rename(nutFile.Name(), fileName))
	nutFile = nil
	fatalIfErr(os.Chmod(fileName, nutFilePerm))
}

// Unpack nut file with given fileName into dir, overwriting files.
// Creates dir if needed. Removes dir first if asked.
func unpackNut(fileName string, dir string, removeDir, verbose bool) {
	// check dir
	_, err := os.Stat(dir)
	if err == nil && removeDir {
		if verbose {
			log.Printf("Removing existing directory %s ...", dir)
		}
		fatalIfErr(os.RemoveAll(dir))
	}
	fatalIfErr(os.MkdirAll(dir, workspaceDirPerm))

	nf := new(NutFile)
	fatalIfErr(nf.ReadFile(fileName))

	for _, file := range nf.Reader.File {
		if verbose {
			log.Printf("Unpacking %s ...", file.Name)
		}

		src, err := file.Open()
		fatalIfErr(err)

		dst, err := os.OpenFile(filepath.Join(dir, file.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		fatalIfErr(err)

		_, err = io.Copy(dst, src)
		fatalIfErr(err)

		fatalIfErr(src.Close())
		fatalIfErr(dst.Close())
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

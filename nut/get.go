package main

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	. "github.com/AlekSi/nut"
)

var (
	cmdGet = &Command{
		Run:       runGet,
		UsageLine: "get [-p prefix] [-v] [name or URL]",
		Short:     "download and install nut",
	}

	getP string
	getV bool
)

func init() {
	cmdGet.Long = `
Download and install nut from http://gonuts.io/ or specified URL.
	`

	cmdGet.Flag.StringVar(&getP, "p", "", "install prefix in workspace, uses hostname if omitted")
	cmdGet.Flag.BoolVar(&getV, "v", false, vHelp)
}

func ArgToURL(s string) (url *url.URL) {
	var err error

	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		url, err = url.Parse(s)
		PanicIfErr(err)
		return
	}

	p := strings.Split(s, "/")
	if len(p) > 1 && (p[0] == DefaultServer) {
		s = strings.Join(p[1:], "/")
	}
	url, err = url.Parse(fmt.Sprintf("http://%s/%s", GonutsServer, s))
	PanicIfErr(err)
	return
}

func get(url *url.URL) (b []byte, err error) {
	if getV {
		log.Printf("Getting %s ...", url)
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/zip")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode/100 != 2 {
		err = fmt.Errorf("Status code %d", res.StatusCode)
		return
	}

	if getV {
		log.Printf("Status code %d", res.StatusCode)
	}

	return
}

func nutImports(imports []string) (nuts []string) {
	for _, imp := range imports {
		if strings.HasPrefix(imp, DefaultServer+"/") {
			nuts = append(nuts, imp[len(DefaultServer)+1:])
		}
	}
	return
}

func runGet(cmd *Command) {
	if !getV {
		getV = config.V
	}

	args := cmd.Flag.Args()

	// zero arguments is a special case – install dependencies for package in current directory
	if len(args) == 0 {
		pack, err := build.ImportDir(".", 0)
		PanicIfErr(err)
		args = nutImports(pack.Imports)
		if getV && len(args) != 0 {
			log.Printf("%s depends on nuts: %s", pack.Name, strings.Join(args, ","))
		}
	}

	installPaths := make([]string, 0, len(args))
	for len(args) != 0 {
		arg := args[0]
		args = args[1:]

		url := ArgToURL(arg)
		b, err := get(url)
		PanicIfErr(err)

		nf := new(NutFile)
		nf.ReadFrom(bytes.NewReader(b))
		deps := nutImports(nf.Imports)
		if getV && len(deps) != 0 {
			log.Printf("%s depends on nuts: %s", nf.Name, strings.Join(deps, ","))
		}
		args = append(args, deps...)

		p := getP
		if p == "" {
			if strings.Contains(url.Host, ":") {
				p, _, err = net.SplitHostPort(url.Host)
				PanicIfErr(err)
			} else {
				p = url.Host
			}
			if strings.HasPrefix(p, "www.") {
				p = p[4:]
			}
		}
		fileName := WriteNut(b, p, getV)
		path := filepath.Join(p, nf.Name, nf.Version.String())

		UnpackNut(fileName, filepath.Join(SrcDir, path), true, getV)
		installPaths = append(installPaths, path)
	}

	for _, path := range installPaths {
		InstallPackage(path, getV)
	}
}

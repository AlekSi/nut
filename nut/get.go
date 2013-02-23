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
		UsageLine: "get [-p prefix] [-v] [name, import path or URL]",
		Short:     "download and install nut and dependencies",
	}

	getP string
	getV bool
)

func init() {
	cmdGet.Long = `
Downloads and installs nut and dependencies from http://gonuts.io/ or specified URL.

Examples:
    nut install AlekSi/nut
    nut install AlekSi/nut/0.2.0
    nut install gonuts.io/AlekSi/nut
    nut install gonuts.io/AlekSi/nut/0.2.0
    nut install http://gonuts.io/AlekSi/nut
    nut install http://gonuts.io/AlekSi/nut/0.2.0
`

	cmdGet.Flag.StringVar(&getP, "p", "", "install prefix in workspace, uses hostname from URL if omitted")
	cmdGet.Flag.BoolVar(&getV, "v", false, vHelp)
}

func ArgToURL(s string) *url.URL {
	var p []string
	var host string
	var ok bool

	// full URL - as is
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		goto parse
	}

	p = strings.Split(s, "/")
	if len(p) > 0 {
		host, ok = NutImportPrefixes[p[0]]
	}
	if ok {
		// import path style
		p[0] = "http://" + host
		s = strings.Join(p, "/")
	} else {
		// short style
		s = fmt.Sprintf("http://%s/%s", NutImportPrefixes["gonuts.io"], s)
	}

parse:
	u, err := url.Parse(s)
	PanicIfErr(err)
	return u
}

func get(url *url.URL) (b []byte, err error) {
	if getV {
		log.Printf("Getting %s ...", url)
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "nut getter")
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

func runGet(cmd *Command) {
	if !getV {
		getV = Config.V
	}

	args := cmd.Flag.Args()

	// zero arguments is a special case – install dependencies for package in current directory
	if len(args) == 0 {
		pack, err := build.ImportDir(".", 0)
		PanicIfErr(err)
		args = NutImports(pack.Imports)
		if getV && len(args) != 0 {
			log.Printf("%s depends on nuts: %s", pack.Name, strings.Join(args, ","))
		}
	}

	installPaths := make(map[string]bool, len(args))
	for len(args) != 0 {
		arg := args[0]
		args = args[1:]

		url := ArgToURL(arg)
		b, err := get(url)
		PanicIfErr(err)

		nf := new(NutFile)
		nf.ReadFrom(bytes.NewReader(b))
		deps := NutImports(nf.Imports)
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
		installPaths[path] = true
	}

	for path := range installPaths {
		InstallPackage(path, getV)
	}
}

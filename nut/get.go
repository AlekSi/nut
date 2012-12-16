package main

import (
	"bytes"
	"fmt"
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
	if len(p) > 1 && (p[0] == DefaultServer[4:]) {
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

func runGet(cmd *Command) {
	if !getV {
		getV = config.V
	}

	args := cmd.Flag.Args()
	paths := make([]string, 0, len(args))
	for {
		if len(args) == 0 {
			break
		}
		arg := args[0]
		args = args[1:]

		url := ArgToURL(arg)
		b, err := get(url)
		PanicIfErr(err)

		nf := new(NutFile)
		nf.ReadFrom(bytes.NewReader(b))

		for _, imp := range nf.Imports {
			if strings.HasPrefix(imp, DefaultServer[4:]+"/") {
				d := imp[len(DefaultServer[4:])+1:]
				if getV {
					log.Printf("%s %s (%s) depends on %s.", nf.Name, nf.Version, arg, d)
				}
				args = append(args, d)
			}
		}

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
		paths = append(paths, path)
	}

	for _, path := range paths {
		InstallPackage(path, getV)
	}
}

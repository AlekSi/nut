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

const (
	maxDownloads = 4
)

var (
	cmdGet = &Command{
		Run:       runGet,
		UsageLine: "get [-p prefix] [-server server] [-v] [name or URL]",
		Short:     "download and install nut",
	}

	getP string
	getS string
	getV bool
)

func init() {
	cmdGet.Long = `
Download and install nut from http://gonuts.io/ or specified URL.
	`

	cmdGet.Flag.StringVar(&getP, "p", "", "install prefix in workspace, uses hostname if omitted")
	cmdGet.Flag.StringVar(&getS, "server", "www.gonuts.io", "server to use")
	cmdGet.Flag.BoolVar(&getV, "v", false, "be verbose")
}

func ArgToURL(s string) (url *url.URL) {
	var err error

	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		url, err = url.Parse(s)
		PanicIfErr(err)
		return
	}

	switch strings.Count(s, "/") {
	case 0, 1:
		url, err = url.Parse(fmt.Sprintf("http://%s/%s", getS, s))
	case 2:
		p := strings.Split(s, "/")
		url, err = url.Parse(fmt.Sprintf("http://%s/%s-%s.nut", p[0], p[1], p[2]))
	default:
		log.Panicf("Failed to parse argument %q", s)
	}

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
	defer res.Body.Close()
	if err != nil {
		return
	}

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode/100 != 2 {
		err = fmt.Errorf("Status code %d", res.StatusCode)
		return
	}

	return
}

func runGet(cmd *Command) {
	for _, arg := range cmd.Flag.Args() {
		url := ArgToURL(arg)
		b, err := get(url)
		PanicIfErr(err)

		nf := new(NutFile)
		nf.ReadFrom(bytes.NewReader(b))

		p := getP
		if p == "" {
			p, _, err = net.SplitHostPort(url.Host)
			PanicIfErr(err)
		}
		fileName := WriteNut(b, p, getV)
		path := filepath.Join(p, nf.Name, nf.Version.String())

		UnpackNut(fileName, filepath.Join(SrcDir, path), true, getV)
		InstallPackage(path, getV)
	}
}

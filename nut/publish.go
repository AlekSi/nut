package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var (
	cmdPublish = &Command{
		Run:       runPublish,
		UsageLine: "publish [-server server] [-token token] [-v] [filename]",
		Short:     "publish nut on gonuts.io",
	}

	publishServer string
	publishToken  string
	publishV      bool
)

func init() {
	cmdPublish.Long = `
Publish nut on http://gonuts.io/ (requires registration with Google account).
	`

	tokenHelp := fmt.Sprintf("access token (see http://gonuts.io/-/me), may be read from ~/%s", ConfigFileName)
	cmdPublish.Flag.StringVar(&publishServer, "server", "www.gonuts.io", "server to use")
	cmdPublish.Flag.StringVar(&publishToken, "token", "", tokenHelp)
	cmdPublish.Flag.BoolVar(&publishV, "v", false, "be verbose")
}

func runPublish(cmd *Command) {
	if publishToken == "" {
		publishToken = config.Token
	}

	// otherwise localhost:8080 will be parsed as url.URL{Scheme:"localhost", Opaque:"8080"}
	if !(strings.HasPrefix(publishServer, "http://") || strings.HasPrefix(publishServer, "https://")) {
		publishServer = "http://" + publishServer
	}
	url, err := url.Parse(publishServer)
	PanicIfErr(err)

	url.RawQuery = "token=" + publishToken

	for _, arg := range cmd.Flag.Args() {
		b, nf := ReadNut(arg)
		url.Path = fmt.Sprintf("/%s/%s", nf.Name, nf.Version)

		if publishV {
			log.Printf("Putting %s to %s ...", arg, url)
		}
		req, err := http.NewRequest("PUT", url.String(), bytes.NewReader(b))
		PanicIfErr(err)
		req.Header.Set("Content-Type", "application/zip")
		req.ContentLength = int64(len(b)) // set Content-Length explicitly: dev_appserver.py doesn't support chunked encoding

		res, err := http.DefaultClient.Do(req)
		PanicIfErr(err)

		b, err = ioutil.ReadAll(res.Body)
		PanicIfErr(err)
		res.Body.Close()

		var body map[string]interface{}
		err = json.Unmarshal(b, &body)
		if err != nil {
			log.Print(err)
		}

		m, ok := body["Message"]
		if ok {
			ok = res.StatusCode/100 == 2
		} else {
			m = string(b)
		}
		if !ok {
			log.Fatal(m)
		}
		if publishV {
			log.Print(m)
		}
	}
}

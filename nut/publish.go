package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var (
	cmdPublish = &Command{
		Run:       runPublish,
		UsageLine: "publish [-token token] [-v] [filename]",
		Short:     "publish nut on gonuts.io",
	}

	publishToken string
	publishV     bool
)

func init() {
	cmdPublish.Long = `
Publishes nut on http://gonuts.io/ (requires registration with Google account).

Examples:
    nut publish test_nut1-0.0.1.nut
`

	tokenHelp := fmt.Sprintf("access token from http://gonuts.io/-/me (may be read from ~/%s)", ConfigFileName)
	cmdPublish.Flag.StringVar(&publishToken, "token", "", tokenHelp)
	cmdPublish.Flag.BoolVar(&publishV, "v", false, vHelp)
}

func runPublish(cmd *Command) {
	if publishToken == "" {
		publishToken = Config.Token
	}
	if !publishV {
		publishV = Config.V
	}

	url, err := url.Parse("http://" + NutImportPrefixes["gonuts.io"])
	FatalIfErr(err)

	url.RawQuery = "token=" + publishToken

	for _, arg := range cmd.Flag.Args() {
		b, nf := ReadNut(arg)
		url.Path = fmt.Sprintf("/%s/%s/%s", nf.Vendor, nf.Name, nf.Version)

		if publishV {
			log.Printf("Putting %s to %s ...", arg, url)
		}
		req, err := http.NewRequest("PUT", url.String(), bytes.NewReader(b))
		FatalIfErr(err)
		req.Header.Set("User-Agent", "nut publisher")
		req.Header.Set("Content-Type", "application/zip")
		req.ContentLength = int64(len(b))

		res, err := http.DefaultClient.Do(req)
		FatalIfErr(err)
		b, err = ioutil.ReadAll(res.Body)
		FatalIfErr(err)
		err = res.Body.Close()
		FatalIfErr(err)

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

package main_test

import (
	. "."
	. "launchpad.net/gocheck"
)

type G struct {
	old string
}

var _ = Suite(&G{})

func (g *G) SetUpSuite(*C) {
	g.old = NutImportPrefixes["gonuts.io"]
	NutImportPrefixes["gonuts.io"] = "server"
	NutImportPrefixes["express42.com"] = "express42.com"
}

func (g *G) TearDownSuite(*C) {
	NutImportPrefixes["gonuts.io"] = g.old
	delete(NutImportPrefixes, "express42.com")
}

func (*G) TestArgToURL(c *C) {
	// short style
	c.Check(ArgToURL("AlekSi/test_nut1").String(), Equals, "http://server/AlekSi/test_nut1")
	c.Check(ArgToURL("AlekSi/test_nut1/0.0.1").String(), Equals, "http://server/AlekSi/test_nut1/0.0.1")

	// import path style
	c.Check(ArgToURL("gonuts.io/AlekSi/test_nut1").String(), Equals, "http://server/AlekSi/test_nut1")
	c.Check(ArgToURL("gonuts.io/AlekSi/test_nut1/0.0.1").String(), Equals, "http://server/AlekSi/test_nut1/0.0.1")
	c.Check(ArgToURL("express42.com/nuts/AlekSi/test_nut1").String(), Equals, "http://express42.com/nuts/AlekSi/test_nut1")
	c.Check(ArgToURL("express42.com/nuts/AlekSi/test_nut1/0.0.1").String(), Equals, "http://express42.com/nuts/AlekSi/test_nut1/0.0.1")

	// full URL - as is
	c.Check(ArgToURL("http://www.gonuts.io/AlekSi/test_nut1").String(), Equals, "http://www.gonuts.io/AlekSi/test_nut1")
	c.Check(ArgToURL("http://www.gonuts.io/AlekSi/test_nut1/0.0.1").String(), Equals, "http://www.gonuts.io/AlekSi/test_nut1/0.0.1")
	c.Check(ArgToURL("http://localhost:8080/AlekSi/test_nut1-0.0.1.nut").String(), Equals, "http://localhost:8080/AlekSi/test_nut1-0.0.1.nut")
	c.Check(ArgToURL("http://example.com/nuts/test_nut1-0.0.1.nut").String(), Equals, "http://example.com/nuts/test_nut1-0.0.1.nut")
	c.Check(ArgToURL("https://example.com/nuts/test_nut1-0.0.1.nut").String(), Equals, "https://example.com/nuts/test_nut1-0.0.1.nut")
}

func (*G) TestNutImports(c *C) {
	actual := NutImports([]string{"fmt", "log/syslog", "github.com/AlekSi/nut", "gonuts.io/AlekSi/test_nut1"})
	c.Check(actual, DeepEquals, []string{"gonuts.io/AlekSi/test_nut1"})
}

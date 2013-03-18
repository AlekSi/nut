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

func (*G) TestParseArg(c *C) {
	data := [][3]string{
		// short style
		{"aleksi/test_nut1", "http://server/aleksi/test_nut1", "gonuts.io"},
		{"aleksi/test_nut1/0.0.1", "http://server/aleksi/test_nut1/0.0.1", "gonuts.io"},

		// import path style
		{"gonuts.io/aleksi/test_nut1", "http://server/aleksi/test_nut1", "gonuts.io"},
		{"gonuts.io/aleksi/test_nut1/0.0.1", "http://server/aleksi/test_nut1/0.0.1", "gonuts.io"},
		{"express42.com/nuts/aleksi/test_nut1", "http://express42.com/nuts/aleksi/test_nut1", "express42.com"},
		{"express42.com/nuts/aleksi/test_nut1/0.0.1", "http://express42.com/nuts/aleksi/test_nut1/0.0.1", "express42.com"},

		// full URL - as is
		{"http://www.gonuts.io/aleksi/test_nut1", "http://www.gonuts.io/aleksi/test_nut1", "gonuts.io"},
		{"http://www.gonuts.io/aleksi/test_nut1/0.0.1", "http://www.gonuts.io/aleksi/test_nut1/0.0.1", "gonuts.io"},
		{"http://localhost:8080/aleksi/test_nut1-0.0.1.nut", "http://localhost:8080/aleksi/test_nut1-0.0.1.nut", "localhost"},
		{"http://example.com/nuts/test_nut1-0.0.1.nut", "http://example.com/nuts/test_nut1-0.0.1.nut", "example.com"},
		{"https://example.com/nuts/test_nut1-0.0.1.nut", "https://example.com/nuts/test_nut1-0.0.1.nut", "example.com"},
	}

	for _, d := range data {
		u, prefix := ParseArg(d[0])
		c.Check(u.String(), Equals, d[1])
		c.Check(prefix, Equals, d[2])
	}
}

func (*G) TestNutImports(c *C) {
	actual := NutImports([]string{"fmt", "log/syslog", "github.com/aleksi/nut", "gonuts.io/aleksi/test_nut1"})
	c.Check(actual, DeepEquals, []string{"gonuts.io/aleksi/test_nut1"})
}

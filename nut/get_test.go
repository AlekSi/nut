package main_test

import (
	. "."
	. "launchpad.net/gocheck"
)

type G struct{}

var _ = Suite(&G{})

func (*G) TestArgToURL(c *C) {
	var orig string
	orig, GonutsServer = GonutsServer, "server"
	defer func() {
		GonutsServer = orig
	}()

	// short style
	c.Check(ArgToURL("test_nut1").String(), Equals, "http://server/test_nut1")
	c.Check(ArgToURL("test_nut1/0.0.1").String(), Equals, "http://server/test_nut1/0.0.1")

	// import path style
	c.Check(ArgToURL("gonuts.io/test_nut1").String(), Equals, "http://server/test_nut1")
	c.Check(ArgToURL("gonuts.io/test_nut1/0.0.1").String(), Equals, "http://server/test_nut1/0.0.1")

	// full URL - as is
	c.Check(ArgToURL("http://www.gonuts.io/test_nut1").String(), Equals, "http://www.gonuts.io/test_nut1")
	c.Check(ArgToURL("http://www.gonuts.io/test_nut1/0.0.1").String(), Equals, "http://www.gonuts.io/test_nut1/0.0.1")
	c.Check(ArgToURL("http://localhost:8080/test_nut1-0.0.1.nut").String(), Equals, "http://localhost:8080/test_nut1-0.0.1.nut")
	c.Check(ArgToURL("http://example.com/test_nut1-0.0.1.nut").String(), Equals, "http://example.com/test_nut1-0.0.1.nut")
}

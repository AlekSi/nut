package main_test

import (
	. "."
	. "launchpad.net/gocheck"
)

type G struct{}

var _ = Suite(&G{})

func (*G) TestArgToURL(c *C) {
	c.Check(ArgToURL("test_nut1").String(), Equals, "http://www.gonuts.io/test_nut1")
	c.Check(ArgToURL("test_nut1/0.0.1").String(), Equals, "http://www.gonuts.io/test_nut1/0.0.1")
	c.Check(ArgToURL("http://gonuts.io/test_nut1").String(), Equals, "http://gonuts.io/test_nut1")
	c.Check(ArgToURL("http://gonuts.io/test_nut1/0.0.1").String(), Equals, "http://gonuts.io/test_nut1/0.0.1")

	c.Check(ArgToURL("localhost/test_nut1/0.0.1").String(), Equals, "http://localhost/test_nut1-0.0.1.nut")

	c.Check(ArgToURL("http://localhost/test_nut1-0.0.1.nut").String(), Equals, "http://localhost/test_nut1-0.0.1.nut")
}

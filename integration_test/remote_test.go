package integration_test

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	. "launchpad.net/gocheck"
)

type R struct{}

var _ = Suite(&R{})

func (*R) SetUpTest(c *C) {
	setupTest(c)

	server := os.Getenv("GONUTS_IO_SERVER")
	u, err := url.Parse("http://" + server + "/debug/prepare_test")
	c.Assert(err, IsNil)
	res, err := http.Get(u.String())
	c.Assert(err, IsNil)
	res.Body.Close()
	c.Assert(res.StatusCode, Equals, 200)
}

func (r *R) TearDownTest(c *C) {
	r.SetUpTest(c)
}

func (*R) TestPublishGet(c *C) {
	if testing.Short() {
		c.Skip("-short passed")
		return
	}

	_, stderr := runNut(c, TestNut1, "pack -v")
	c.Check(strings.HasSuffix(stderr, `test_nut1-0.0.1.nut created.`), Equals, true)
	gitNoDiff(c, TestNut1)

	_, stderr = runNut(c, TestNut2, "pack -v")
	c.Check(strings.HasSuffix(stderr, `test_nut2-0.0.2.nut created.`), Equals, true)
	gitNoDiff(c, TestNut2)

	_, stderr = runNut(c, TestNut1, "publish -v test_nut1-0.0.1.nut")
	c.Check(strings.HasSuffix(stderr, `Nut debug/test_nut1 version 0.0.1 published.`), Equals, true)
	gitNoDiff(c, TestNut1)

	_, stderr = runNut(c, TestNut1, "publish -v test_nut1-0.0.1.nut", 1)
	c.Check(strings.HasSuffix(stderr, `Nut debug/test_nut1 version 0.0.1 already exists.`), Equals, true)
	gitNoDiff(c, TestNut1)

	_, stderr = runNut(c, TestNut2, "publish -v test_nut2-0.0.2.nut")
	c.Check(strings.HasSuffix(stderr, `Nut debug/test_nut2 version 0.0.2 published.`), Equals, true)
	gitNoDiff(c, TestNut2)

	// _, stderr = runNut(c, "", "get -v test_nut2/0.0.2")
	// c.Check(strings.HasSuffix(stderr, `Nut "test_nut2" version "0.0.2" published.`), Equals, true)
}

package integration_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/AlekSi/nut/nut"
	. "launchpad.net/gocheck"
)

type R struct{}

var _ = Suite(&R{})

func (*R) SetUpTest(c *C) {
	if testing.Short() {
		c.Skip("-short passed")
		return
	}

	setupTest(c)

	server := os.Getenv("GONUTS_IO_SERVER")
	u, err := url.Parse(server + "/debug/prepare_test")
	c.Assert(err, IsNil)
	u.RawQuery = "token=" + main.Config.Token // importing package main... what a hack
	res, err := http.Get(u.String())
	c.Assert(err, IsNil)
	body, err := ioutil.ReadAll(res.Body)
	c.Assert(err, IsNil)
	res.Body.Close()
	c.Assert(res.StatusCode, Equals, 200, Commentf("%s", body))
}

func (r *R) TearDownTest(c *C) {
	r.SetUpTest(c)
}

func (*R) TestPublishGet(c *C) {
	_, stderr := runNut(c, TestNut1, "pack -v")
	c.Check(strings.HasSuffix(stderr, `test_nut1-0.0.1.nut created.`), Equals, true)
	gitNoDiff(c, TestNut1)

	_, stderr = runNut(c, TestNut2, "pack -v")
	c.Check(strings.HasSuffix(stderr, `test_nut2-0.2.0.nut created.`), Equals, true)
	gitNoDiff(c, TestNut2)

	_, stderr = runNut(c, TestNut3, "pack -v")
	c.Check(strings.HasSuffix(stderr, `test_nut3-0.3.0.nut created.`), Equals, true)
	gitNoDiff(c, TestNut3)

	_, stderr = runNut(c, TestNut1, "publish -v test_nut1-0.0.1.nut")
	c.Check(strings.HasSuffix(stderr, `Nut debug/test_nut1 version 0.0.1 published.`), Equals, true)
	gitNoDiff(c, TestNut1)

	_, stderr = runNut(c, TestNut1, "publish -v test_nut1-0.0.1.nut", 1)
	c.Check(strings.HasSuffix(stderr, `Nut debug/test_nut1 version 0.0.1 already exists.`), Equals, true)
	gitNoDiff(c, TestNut1)

	_, stderr = runNut(c, TestNut2, "publish -v test_nut2-0.2.0.nut")
	c.Check(strings.HasSuffix(stderr, `Nut debug/test_nut2 version 0.2.0 published.`), Equals, true)
	gitNoDiff(c, TestNut2)

	_, stderr = runNut(c, TestNut3, "publish -v test_nut3-0.3.0.nut")
	c.Check(strings.HasSuffix(stderr, `Nut debug/test_nut3 version 0.3.0 published.`), Equals, true)
	gitNoDiff(c, TestNut3)

	_, stderr = runNut(c, "", "get -v debug/test_nut3/0.3.0")
	c.Check(strings.Count(stderr, "gonuts.io/test_nut1-0.0.1.nut"), Equals, 1)
	c.Check(strings.Count(stderr, "gonuts.io/test_nut2-0.2.0.nut"), Equals, 1)
	c.Check(strings.Count(stderr, "gonuts.io/test_nut3-0.3.0.nut"), Equals, 1)
	c.Check(strings.HasSuffix(stderr, `gonuts.io/debug/test_nut3`), Equals, true)
}

func (*R) TestGet404(c *C) {
	_, stderr := runNut(c, "", "get -v debug/test_nut3/7.8.999", 1)
	c.Check(strings.Contains(stderr, "Status code 404"), Equals, true)
	_, stderr = runNut(c, "", "get -v debug/no-such-nut", 1)
	c.Check(strings.Contains(stderr, "Status code 404"), Equals, true)
	_, stderr = runNut(c, "", "get -v debug", 1)
	c.Check(stderr, Equals, "invalid argument")
}

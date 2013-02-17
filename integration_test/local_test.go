package integration_test

import (
	"os"
	"runtime"
	"strings"

	. "launchpad.net/gocheck"
)

type L struct{}

var _ = Suite(&L{})

func (*L) SetUpTest(c *C) {
	setupTest(c)
}

func (l *L) TearDownTest(c *C) {
	l.SetUpTest(c)
}

func (*L) TestGenerateCheck(c *C) {
	_, stderr := runNut(c, TestNut1, "generate -v")
	c.Check(stderr, Equals, "nut.json updated.")
	gitNoDiff(c, TestNut1)

	_, stderr = runNut(c, TestNut1, "check -v")
	c.Check(stderr, Equals, "nut.json looks good.")
	gitNoDiff(c, TestNut1)

	c.Check(os.Remove(TestNut2+"/nut.json"), Equals, nil)
	_, stderr = runNut(c, TestNut2, "generate -v")
	expected := `
nut.json generated.

Now you should edit nut.json to fix following errors:
    Version "0.0.0" is invalid.
    "Crazy Nutter" is not a real person.

After that run 'nut check' to check spec again.`[1:]
	c.Check(stderr, Equals, expected)
	_, err := os.Stat(TestNut2 + "/nut.json")
	c.Check(err, Equals, nil)

	_, stderr = runNut(c, TestNut2, "check -v", 1)
	expected = `
Found errors in nut.json:
    Version "0.0.0" is invalid.
    "Crazy Nutter" is not a real person.`[1:]
	c.Check(stderr, Equals, expected)

	c.Check(os.Remove(TestNut3+"/test_nut3.go"), Equals, nil)
	_, stderr = runNut(c, TestNut3, "generate -v", 1)
	c.Check(stderr, Equals, "no Go source files in .")

	_, stderr = runNut(c, TestNut3, "check -v", 1)
	c.Check(strings.HasPrefix(stderr, "no Go source files in ."), Equals, true)
}

func (*L) TestPackCheckUnpack(c *C) {
	_, stderr := runNut(c, TestNut1, "pack -v")
	c.Check(strings.HasSuffix(stderr, "test_nut1-0.0.1.nut created."), Equals, true)
	gitNoDiff(c, TestNut1)

	_, stderr = runNut(c, TestNut1, "check -v test_nut1-0.0.1.nut")
	c.Check(strings.HasSuffix(stderr, "test_nut1-0.0.1.nut looks good."), Equals, true)
	gitNoDiff(c, TestNut1)

	c.Check(os.Remove(TestNut1+"/test_nut1.go"), Equals, nil)
	_, stderr = runNut(c, TestNut1, "unpack -v test_nut1-0.0.1.nut")
	c.Check(strings.HasSuffix(stderr, "test_nut1-0.0.1.nut unpacked."), Equals, true)
	gitNoDiff(c, TestNut1)

	c.Check(os.Remove(TestNut2+"/nut.json"), Equals, nil)
	runNut(c, TestNut2, "generate -v")
	_, stderr = runNut(c, TestNut2, "pack -v", 1)
	c.Check(strings.HasPrefix(stderr, "Found errors:"), Equals, true)
	_, stderr = runNut(c, TestNut2, "pack -nc -v")
	c.Check(strings.HasSuffix(stderr, "test_nut2-0.0.0.nut created."), Equals, true)

	_, stderr = runNut(c, TestNut2, "check -v test_nut2-0.0.0.nut", 1)
	c.Check(strings.HasPrefix(stderr, "Found errors in test_nut2-0.0.0.nut:"), Equals, true)

	c.Check(os.Remove(TestNut3+"/README"), Equals, nil)
	_, stderr = runNut(c, TestNut3, "pack -nc -v", 1)

	if runtime.GOOS == "windows" {
		c.Check(strings.HasSuffix(stderr, "README: The system cannot find the file specified."), Equals, true)
	} else {
		c.Check(strings.HasSuffix(stderr, "README: no such file or directory"), Equals, true)
	}
}

package integration_test

import (
	"os"
	"strings"

	. "launchpad.net/gocheck"
)

type L struct{}

var _ = Suite(&L{})

func (*L) SetUpTest(c *C) {
	for _, dir := range []string{TestNut1, TestNut2, TestNut3} {
		runCommand(c, dir, "git", "reset --hard origin/master")
		runCommand(c, dir, "git", "clean -xdf")
	}
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

You should fix following issues:
    Version "0.0.0" is invalid.
    "Crazy Nutter" is not a real person.

After that run 'nut check' to check spec again.`[1:]
	c.Check(stderr, Equals, expected)
	_, err := os.Stat(TestNut2 + "/nut.json")
	c.Check(err, Equals, nil)

	_, stderr = runNut(c, TestNut2, "check -v", 1)
	expected = `
Found issues in nut.json:
    Version "0.0.0" is invalid.
    "Crazy Nutter" is not a real person.`[1:]
	c.Check(stderr, Equals, expected)

	c.Check(os.Remove(TestNut3+"/test_nut3.go"), Equals, nil)
	_, stderr = runNut(c, TestNut3, "generate -v", 1)
	c.Check(stderr, Equals, "no Go source files in .")

	_, stderr = runNut(c, TestNut3, "check -v", 1)
	c.Check(strings.HasPrefix(stderr, "no Go source files in ."), Equals, true)
}

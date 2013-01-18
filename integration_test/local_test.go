package integration_test

import (
	"os"

	. "launchpad.net/gocheck"
)

type L struct{}

var _ = Suite(&L{})

func (*L) SetUpTest(c *C) {
	for _, nut := range []string{"test_nut1", "test_nut2", "test_nut3"} {
		dir := "../../" + nut
		runCommand(c, dir, "git", "reset --hard origin/master")
		runCommand(c, dir, "git", "clean -xdf")
	}
}

func (*L) TestGenerate(c *C) {
	_, stderr := runNut(c, "../../test_nut1", "generate -v")
	c.Check(stderr, Equals, "nut.json updated.")

	c.Check(os.Remove("../../test_nut2/nut.json"), Equals, nil)
	_, stderr = runNut(c, "../../test_nut2", "generate -v")
	expected := `
nut.json generated.

You should fix following issues:
    Version "0.0.0" is invalid.
    "Crazy Nutter" is not a real person.

After that run 'nut check' to check spec again.`[1:]
	c.Check(stderr, Equals, expected)
	_, err := os.Stat("../../test_nut2/nut.json")
	c.Check(err, Equals, nil)

	_, stderr = runNut(c, "../../", "generate -v", 1)
	c.Check(stderr, Equals, "no Go source files in .")
	_, err = os.Stat("../../nut.json")
	c.Check(os.IsNotExist(err), Equals, true)
}

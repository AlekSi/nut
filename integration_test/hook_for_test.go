package integration_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	. "launchpad.net/gocheck"
)

// Global gocheck hook.
func TestIntegration(t *testing.T) { TestingT(t) }

var (
	TestNut1 = "../../test_nut1"
	TestNut2 = "../../test_nut2"
	TestNut3 = "../../test_nut3"
	nutBin   string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	nutBin = filepath.Join(wd, "..", "gonut.exe")
}

func setupTest(c *C) {
	for _, dir := range []string{TestNut1, TestNut2, TestNut3} {
		runCommand(c, dir, "git", "reset --hard origin/master")
		runCommand(c, dir, "git", "clean -xdf")
	}

	oa := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	for _, dir := range []string{
		"../../../../gonuts.io/",
		"../../../../localhost/",
		"../../../../../pkg/" + oa + "/gonuts.io/",
		"../../../../../pkg/" + oa + "/localhost/",
		"../../../../../bin/",
		"../../../../../nut/",
	} {
		c.Logf("%s", dir)
		c.Assert(os.RemoveAll(dir), IsNil)
	}
}

func runCommand(c *C, dir, command string, args string, exitCode ...int) (stdout, stderr string) {
	var expectedCode int
	switch len(exitCode) {
	case 0:
	case 1:
		expectedCode = exitCode[0]
	default:
		c.Fatal("Invalid invacation of runCommand")
	}

	o, e := bytes.Buffer{}, bytes.Buffer{}
	cmd := exec.Command(command, strings.Split(args, " ")...)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = &o, &e
	err := cmd.Run()
	stdout, stderr = strings.TrimSpace(o.String()), strings.TrimSpace(e.String())
	c.Logf("%s: %s %s", dir, command, args)
	c.Logf("stdout: %s", stdout)
	c.Logf("stderr: %s", stderr)

	if err == nil {
		if expectedCode == 0 {
			return
		} else {
			c.Fatalf("Expected exit code %d, got 0.", expectedCode)
		}
	}

	ee, ok := err.(*exec.ExitError)
	if !ok {
		c.Fatal(err)
	}
	actualCode := ee.Sys().(syscall.WaitStatus).ExitStatus() // why it's so hard?..
	if expectedCode != actualCode {
		c.Fatalf("Expected exit code %d, got %d.", expectedCode, actualCode)
	}

	return
}

func runNut(c *C, dir string, args string, exitCode ...int) (stdout, stderr string) {
	return runCommand(c, dir, nutBin, args, exitCode...)
}

func runGo(c *C, dir string, args string, exitCode ...int) (stdout, stderr string) {
	return runCommand(c, dir, "go", args, exitCode...)
}

func gitNoDiff(c *C, dir string) {
	runCommand(c, dir, "git", "diff --exit-code")
}

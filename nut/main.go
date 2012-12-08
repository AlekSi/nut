// Package main implements 'nut' command. Code there is not considered to be a public API, and may change without notice.
package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"text/template"
)

// A Command is an implementation of a nut command like nut get or nut install.
type Command struct {
	// Run runs the command.
	// To access args use cmd.Flag.Args().
	Run func(cmd *Command)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'nut help' output.
	Short string

	// Long is the long message shown in the 'nut help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	log.Printf("usage: %s", c.UsageLine)
	log.Printf("\n%s\n\n", strings.TrimSpace(c.Long))
	log.Print("Flags:")
	c.Flag.PrintDefaults()
}

// Commands lists the available commands.
// The order here is the order in which they are printed by 'nut help'.
var Commands = []*Command{cmdCheck, cmdGenerate, cmdGet, cmdInstall, cmdPack, cmdPublish, cmdUnpack}

var usageTemplate = template.Must(template.New("top").Parse(`Nut is a tool for managing versioned Go source code packages.

Usage:

    nut command [arguments]

The commands are:
{{range .}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use "nut help [command]" for more information about a command.

`))

// help implements the 'help' command.
func help(args ...string) {
	if len(args) == 0 {
		flag.Usage()
		os.Exit(0)
	}
	if len(args) != 1 {
		log.Print("usage: nut help [command]\n\nToo many arguments given.")
		os.Exit(2)
	}

	arg := args[0]
	for _, cmd := range Commands {
		if cmd.Name() == arg {
			cmd.Usage()
			os.Exit(0)
		}
	}

	log.Printf("Unknown help topic %#q.  Run 'nut help'.", arg)
	os.Exit(2)
}

func main() {
	flag.Usage = func() {
		usageTemplate.Execute(os.Stderr, Commands)
		flag.PrintDefaults()
	}

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		help()
		panic("not reached")
	}
	if args[0] == "help" {
		help(args[1:]...)
		panic("not reached")
	}

	for _, cmd := range Commands {
		if cmd.Name() == args[0] {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			cmd.Run(cmd)
			os.Exit(0)
		}
	}

	log.Printf("nut: unknown subcommand %q\nRun 'nut help' for usage.", args[0])
	os.Exit(2)
}

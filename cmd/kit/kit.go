package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dvyukov/kit/pkg/kitp"
)

type Command struct {
	run   func() error
	args  *[]string
	flags *flag.FlagSet
	desc  string
}

var commands = make(map[string]*Command)

func register(name string, run func() error, args *[]string, flags *flag.FlagSet, desc string) {
	commands[name] = &Command{
		run:   run,
		args:  args,
		flags: flags,
		desc:  desc,
	}
}

const workdir = ".kit"

var (
	network    kitp.UserID
	configFile = filepath.Join(workdir, "config")
	keyFile    = filepath.Join(workdir, "key")
)

func main() {
	failf := func(msg string, args ...interface{}) {
		fmt.Fprintf(os.Stderr, msg+"\n", args...)
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		failf("usage: ...")
	}
	cmd := commands[os.Args[1]]
	if cmd == nil {
		failf("unknown command: %v", os.Args[1])
	}
	if cmd.flags != nil {
		if err := cmd.flags.Parse(os.Args[2:]); err != nil {
			failf("%v", err)
		}
		if cmd.args == nil && len(cmd.flags.Args()) != 0 {
			failf("arguments are not expected here")
		}
		if cmd.args != nil {
			*cmd.args = cmd.flags.Args()
		}
	}
	if err := cmd.run(); err != nil {
		failf("%v", err)
	}
}

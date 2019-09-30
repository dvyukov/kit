package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/dvyukov/kit/pkg/journal"
	"github.com/dvyukov/kit/pkg/kitp"
)

func init() {
	var flags flag.FlagSet
	flags.StringVar(&mailFlags.foo, "foo", "default", "foo description")
	register("mail", mailRun, &mailFlags.args, &flags, "mail changes")
}

var mailFlags = struct {
	args []string
	foo  string
}{}

func mailRun() error {
	journal, err := journal.Open(workdir)
	if err != nil {
		return fmt.Errorf("failed to open journal: %v", err)
	}
	if len(journal.Messages()) == 0 {
		return fmt.Errorf("not inited")
	}
	user, msgID, seq, err := journal.Me()
	if err != nil {
		return err
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return err
	}
	msg := &kitp.Raw{
		Hdr: kitp.Hdr{
			Seq:  seq,
			User: user,
			Prev: msgID,
		},
	}
	data := &kitp.MsgChange{
		Diff: "+some code",
	}
	if err := msg.AddData(data); err != nil {
		return err
	}
	if err := msg.Seal(key); err != nil {
		return err
	}
	if err := journal.Append(msg); err != nil {
		return fmt.Errorf("failed to write journal: %v", err)
	}
	return nil
}

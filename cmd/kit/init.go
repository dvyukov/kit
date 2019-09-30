package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/dvyukov/kit/pkg/journal"
	"github.com/dvyukov/kit/pkg/kitp"
)

func init() {
	var flags flag.FlagSet
	flags.StringVar(&initFlags.foo, "foo", "default", "foo description")
	register("init", initRun, &initFlags.args, &flags, "initialize workspace")
}

var initFlags = struct {
	args []string
	foo  string
}{}

func initRun() error {
	journal, err := journal.Open(workdir)
	if err != nil {
		return fmt.Errorf("failed to open journal: %v", err)
	}
	if len(journal.Messages()) != 0 {
		return fmt.Errorf("already inited")
	}
	cfg := &Config{
		Hubs: map[string]string{
			"public key of a hub": "address of the hub",
		},
	}
	if err := writeConfig(configFile, cfg); err != nil {
		return err
	}
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	msg := &kitp.Raw{
		Hdr: kitp.Hdr{
			Seq: 0,
		},
	}
	copy(msg.User[:], pubKey)
	reg := &kitp.MsgRegister{
		Email: "my@email.com",
	}
	if err := msg.AddData(reg); err != nil {
		return err
	}
	seed := privKey.Seed()
	if err := msg.Seal(seed); err != nil {
		return err
	}
	if err := journal.Append(msg); err != nil {
		return fmt.Errorf("failed to write journal: %v", err)
	}
	if err := ioutil.WriteFile(keyFile, seed, 0400); err != nil {
		return fmt.Errorf("failed to write key: %v", err)
	}
	return nil
}

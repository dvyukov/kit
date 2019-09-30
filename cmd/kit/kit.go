package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/dvyukov/kit/pkg/journal"
	"github.com/dvyukov/kit/pkg/kitp"
)

type Config struct {
	HubAddr string
}

func main() {
	if len(os.Args) < 2 {
		failf("usage: ...")
	}
	journal, err := journal.New(".kit")
	if err != nil {
		failf("failed to open journal: %v", err)
	}
	msgs, err := journal.Read()
	if err != nil {
		failf("failed to read journal: %v", err)
	}
	cfg := new(Config)
	cfg.HubAddr = ":34567"
	cfgFile := filepath.Join(".kit", "config")
	cmd := os.Args[1]
	switch cmd {
	case "init":
		cfgData, err := json.Marshal(cfg)
		if err != nil {
			failf("failed to marshal config: %v", err)
		}
		if err := os.MkdirAll(".kit", 0755); err != nil {
			failf("failed to create .kit dir: %v", err)
		}
		if err := ioutil.WriteFile(cfgFile, cfgData, 0644); err != nil {
			failf("failed to write config file: %v", err)
		}
		if len(msgs) == 0 {
			msg := &kitp.Msg{
				Hdr: kitp.Hdr{
					Seq: 42,
				},
				Data: &kitp.MsgRegister{
					Name: "foo",
				},
			}
			if err := journal.Write(msg); err != nil {
				failf("failed to write journal: %v", err)
			}
		}
		return
	}
	cfgData, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		failf("failed to read config file: %v", err)
	}
	if err := json.Unmarshal(cfgData, cfg); err != nil {
		failf("failed to parse config: %v", err)
	}
	switch cmd {
	case "mail":
	case "sync":
		conn, err := net.Dial("tcp", cfg.HubAddr)
		if err != nil {
			failf("failed to connect to hub at %v: %v", cfg.HubAddr, err)
		}
		defer conn.Close()
		w := bufio.NewWriter(conn)
		for _, msg := range msgs {
			if err := kitp.Serialize(w, msg); err != nil {
				failf("failed to send: %v", err)
			}
		}
		log.Printf("send %v msgs", len(msgs))
		if err := w.Flush(); err != nil {
			failf("failed to send: %v", err)
		}
		if err := conn.(*net.TCPConn).CloseWrite(); err != nil {
			failf("failed to close connection: %v", err)
		}
		r := bufio.NewReader(conn)
		for i := 0; ; i++ {
			msg, err := kitp.Deserialize(r)
			if err != nil {
				failf("failed to read msg: %v", err)
			}
			if msg == nil {
				log.Printf("recv %v msgs", i)
				break
			}
			journal.Append(msg)
		}
	default:
		failf("unknown command %v", cmd)
	}
}
func fail(err error) {
	failf("%v", err)
}

func failf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

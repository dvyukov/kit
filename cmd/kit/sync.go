package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/dvyukov/kit/pkg/journal"
	"github.com/dvyukov/kit/pkg/kitp"
)

func init() {
	var flags flag.FlagSet
	flags.StringVar(&syncFlags.foo, "foo", "default", "foo description")
	register("sync", syncRun, nil, &flags, "sync with hub")
}

var syncFlags = struct {
	foo string
}{}

func syncRun() error {
	journal, err := journal.Open(workdir)
	if err != nil {
		return fmt.Errorf("failed to open journal: %v", err)
	}
	if len(journal.Messages()) == 0 {
		return fmt.Errorf("not inited")
	}
	me := journal.Messages()[0].User

	cfg, err := readConfig(configFile)
	if err != nil {
		return err
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return err
	}

	for hubKey, hubAddr := range cfg.Hubs {
		c, err := net.Dial("tcp", hubAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to hub at %v: %v", hubAddr, err)
			continue
		}
		defer c.Close()
		//log.Printf("connected to %x at %v", userID[:], addr)

		var userID kitp.UserID
		if _, err := hex.Decode(userID[:], []byte(hubKey)); err != nil {
			log.Fatalf("failed to decode hub %q: %v", hubKey, err)
		}

		cursor, err := journal.PeerCursor(userID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read peer cursor: %v\n", err)
		}

		conn, err := kitp.NewClientConn(c, network, me, userID, key, cursor, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to handshake with %v: %v", hubAddr, err)
			continue
		}
		msgs := journal.SelectCursor(conn.RemoteUser, conn.RemoteSync)
		var senderr chan error
		if len(msgs) != 0 {
			fmt.Fprintf(os.Stderr, "send:\n")
			for _, msg := range msgs {
				fmt.Fprintf(os.Stderr, "%v\n", msg)
			}
			senderr = make(chan error, 1)
			go func() {
				senderr <- conn.SendMany(msgs)
			}()
		}
		var lastRecv *kitp.Raw
		for i := 0; ; i++ {
			msg, err := conn.Recv()
			if err != nil {
				return fmt.Errorf("failed to read msg: %v", err)
			}
			if msg == nil {
				//fmt.Printf("recv %v msgs\n", i)
				break
			}
			if i == 0 {
				fmt.Printf("recv:\n")
			}
			fmt.Fprintf(os.Stderr, "%v\n", msg)
			if ok, err := journal.AppendDup(msg, &userID); err != nil {
				fmt.Printf("failed to append: %v\n", err)
			} else if !ok {
				fmt.Printf("\tdup\n")
			} else {
				lastRecv = msg
			}
		}
		if senderr != nil {
			if err := <-senderr; err != nil {
				return fmt.Errorf("failed to send: %v", err)
			}
		}
		if lastRecv != nil {
			// Ack the last message, so that we don't send
			// everything back on the next sync.
			fmt.Fprintf(os.Stderr, "send:\n")
			fmt.Fprintf(os.Stderr, "%v\n", lastRecv)
			if err := conn.Send(lastRecv); err != nil {
				return fmt.Errorf("failed to send: %v", err)
			}
		}

		/*
			if err := c.(*net.TCPConn).CloseWrite(); err != nil {
				return fmt.Errorf("failed to close connection: %v", err)
			}
		*/

		return nil
	}
	return fmt.Errorf("can't connect to any hub")

	/*
		defer conn.Close()
		w := bufio.NewWriter(conn)
		for _, msg := range journal.Messages() {
			if err := kitp.Serialize(w, msg); err != nil {
				return fmt.Errorf("failed to send: %v", err)
			}
		}
		log.Printf("send %v msgs", len(journal.Messages()))
		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to send: %v", err)
		}
		if err := conn.(*net.TCPConn).CloseWrite(); err != nil {
			return fmt.Errorf("failed to close connection: %v", err)
		}
		r := bufio.NewReader(conn)
		for i := 0; ; i++ {
			msg, err := kitp.Deserialize(r)
			if err != nil {
				return fmt.Errorf("failed to read msg: %v", err)
			}
			if msg == nil {
				log.Printf("recv %v msgs", i)
				break
			}
			journal.Append(msg)
		}

		return nil
	*/
}

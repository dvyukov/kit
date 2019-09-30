package main

import (
	_ "encoding/json"
	"fmt"
	"log"
	//"io"
	"bufio"
	"net"

	"github.com/dvyukov/kit/pkg/journal"
	"github.com/dvyukov/kit/pkg/kitp"
)

const addr = ":34567"
const workdir = "/tmp/kithub"

func main() {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %v: %v", addr, err)
	}
	log.Printf("listening on %v", addr)

	journal, err := journal.New(workdir)
	if err != nil {
		log.Fatalf("failed to open journal: %v", err)
	}
	for i := 0; ; i++ {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}
		conn := &Conn{
			index:   i,
			conn:    c,
			r:       bufio.NewReader(c),
			w:       bufio.NewWriter(c),
			journal: journal,
		}
		go conn.serve()
	}
}

type Conn struct {
	index   int
	conn    net.Conn
	r       *bufio.Reader
	w       *bufio.Writer
	journal *journal.Journal
}

func (conn *Conn) serve() {
	defer func() {
		conn.w.Flush()
		conn.conn.Close()
	}()
	if err := conn.serveImpl(); err != nil {
		log.Printf("conn #%v: %v", conn.index, err)
	}
}

func (conn *Conn) serveImpl() error {
	if err := conn.handshake(); err != nil {
		return fmt.Errorf("failed to handshake: %v", err)
	}
	for i := 0; ; i++ {
		msg, err := kitp.Deserialize(conn.r)
		if err != nil {
			return fmt.Errorf("failed to read msg: %v", err)
		}
		if msg == nil {
			log.Printf("conn #%v: read %v msgs", conn.index, i)
			break
		}
		conn.journal.Append(msg)
	}
	msgs, err := conn.journal.Read()
	if err != nil {
		return fmt.Errorf("failed to read journal: %v", err)
	}
	log.Printf("conn #%v: sending %v msgs", conn.index, len(msgs))
	for _, msg := range msgs {
		if err := kitp.Serialize(conn.w, msg); err != nil {
			return fmt.Errorf("failed to write message: %v", err)
		}
	}
	return nil
}

func (conn *Conn) handshake() error {
	return nil
}

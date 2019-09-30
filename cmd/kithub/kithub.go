package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/dvyukov/kit/pkg/journal"
	"github.com/dvyukov/kit/pkg/kitp"
)

const (
	configFile = "config"
	keyFile    = "key"
)

var network kitp.UserID

type Config struct {
	Addrs []string
	Hubs  map[string]string `json:"hubs"`
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "init" {
		if err := initialize(); err != nil {
			log.Fatal(err)
		}
		return
	}
	hub, err := NewHub(".")
	if err != nil {
		log.Fatal(err)
	}
	go hub.loop()

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("failde to read key: %v", err)
	}
	cfg, err := readConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	for _, addr := range cfg.Addrs {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to listen on %v: %v", addr, err)
		}
		log.Printf("listening on %v", addr)
		go func() {
			for {
				c, err := ln.Accept()
				if err != nil {
					log.Printf("failed to accept: %v", err)
					continue
				}
				go func() {
					conn, err := kitp.NewServerConn(c, network[:], hub.Me[:], key, hub.GetCursor)
					if err != nil {
						log.Printf("handshake with %v failed: %v", c.RemoteAddr(), err)
						return
					}
					log.Printf("connection from %x at %v", conn.RemoteUser[:4], c.RemoteAddr())
					serveConn(conn, hub)
					log.Printf("disconnected from %x at %v", conn.RemoteUser[:4], c.RemoteAddr())
				}()
			}
		}()
	}
	for user, addr := range cfg.Hubs {
		addr := addr
		var userID kitp.UserID
		if _, err := hex.Decode(userID[:], []byte(user)); err != nil {
			log.Fatalf("failed to decode hub %q: %v", user, err)
		}
		go func() {
			// TODO: reconnect.
			c, err := net.Dial("tcp", addr)
			if err != nil {
				log.Printf("failed to connect to %v: %v", addr, err)
				return
			}
			defer c.Close()
			log.Printf("connected to %x at %v", userID[:4], addr)
			// TODO: ok to access concurrnetly?
			conn, err := kitp.NewClientConn(c, network, hub.Me, userID, key, hub.GetCursor(&userID), true)
			if err != nil {
				log.Printf("failed to handshake with %v: %v", addr, err)
				return
			}
			serveConn(conn, hub)
			log.Printf("disconnected from %x at %v", userID[:4], addr)
		}()
	}
	select {}
}

func serveConn(conn *kitp.Conn, hub *Hub) {
	errc := make(chan error, 2)
	ctx := hub.Register(conn.RemoteUser, conn.RemoteSync)

	go func() {
		for msgs := range ctx.send {
			buf := bytes.NewBuffer(nil)
			for _, msg := range msgs {
				fmt.Fprintf(buf, "%v\n", msg)
			}
			log.Printf("user %x send:\n%s", conn.RemoteUser[:4], buf.Bytes())
			if err := conn.SendMany(msgs); err != nil {
				errc <- err
				return
			}
		}
		if err := conn.CloseWrite(); err != nil {
			errc <- err
		}
		errc <- nil
	}()

	go func() {
		defer hub.Unregister(ctx)
		for {
			msg, err := conn.Recv()
			if err != nil {
				errc <- err
				return
			}
			if msg == nil {
				errc <- nil
				return
			}
			log.Printf("user %x recv:\n%v", conn.RemoteUser[:4], msg)
			hub.Recv <- clientMsg{ctx, msg}
		}
	}()

	for i := 0; i < 2; i++ {
		err := <-errc
		if err != nil {
			log.Printf("conn error: %v", err)
			conn.Close()
		}
	}
}

func initialize() error {
	journal, err := journal.Open(".")
	if err != nil {
		return fmt.Errorf("failed to open journal: %v", err)
	}
	if len(journal.Messages()) != 0 {
		log.Printf("already inited as %x", journal.Messages()[0].User)
		return nil
	}
	cfg := &Config{
		Addrs: []string{"127.0.0.1:34567"},
		Hubs: map[string]string{
			"public key of another hub": "address of the hub",
		},
	}
	if err := writeConfig(configFile, cfg); err != nil {
		return err
	}
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	msg := &kitp.Raw{}
	copy(msg.User[:], pubKey)
	reg := &kitp.MsgRegister{}
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
	log.Printf("inited as %x", pubKey)
	return nil
}

func readConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	cfg := new(Config)
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}
	// TODO: verify data.
	return cfg, nil
}

func writeConfig(file string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	if err := ioutil.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}
	return nil
}

/*
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
*/

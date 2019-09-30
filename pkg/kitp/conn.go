package kitp

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"unsafe"
)

type Conn struct {
	c          net.Conn
	r          *bufio.Reader
	w          *bufio.Writer
	RemoteUser UserID
	RemoteSync Sync
}

func NewClientConn(c net.Conn, network, user, peer UserID, key []byte, cursor *MsgID, follow bool) (*Conn, error) {
	hello := &Hello{
		Ver:     Ver0,
		Network: network,
		User:    user,
	}
	if follow {
		hello.Flags |= SyncFollow
	}
	if cursor != nil {
		hello.Flags |= SyncCursor
		hello.Cursor = *cursor
	}
	if _, err := io.ReadFull(rand.Reader, hello.Nonce[:]); err != nil {
		return nil, err
	}
	// TODO: the handshake procedure is most likely insecure and needs to be reimplemented.
	sign := ed25519.Sign(ed25519.NewKeyFromSeed(key), hello.Nonce[:])
	if len(hello.Sign) != len(sign) {
		return nil, fmt.Errorf("signature size mismatch: %v/%v", len(hello.Sign), len(sign))
	}
	copy(hello.Sign[:], sign)

	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, hello)
	if _, err := c.Write(buf.Bytes()); err != nil {
		return nil, err
	}
	// Receive ack.
	ackData := make([]byte, unsafe.Sizeof(HelloAck{}))
	if _, err := io.ReadFull(c, ackData); err != nil {
		return nil, err
	}
	ack := &HelloAck{}
	if err := binary.Read(bytes.NewReader(ackData), binary.LittleEndian, ack); err != nil {
		return nil, err
	}
	if ack.Ver != Ver0 {
		return nil, fmt.Errorf("bad version in hello ack %x", ack.Ver)
	}
	if !ed25519.Verify(ed25519.PublicKey(peer[:]), hello.Nonce[:], ack.Sign[:]) {
		return nil, fmt.Errorf("bad hello ack signature")
	}
	return newConn(c, peer, ack.Sync)
}

func NewServerConn(c net.Conn, network, user, key []byte, getCursor func(*UserID) *MsgID) (*Conn, error) {
	helloData := make([]byte, unsafe.Sizeof(Hello{}))
	if _, err := io.ReadFull(c, helloData); err != nil {
		return nil, err
	}
	hello := &Hello{}
	if err := binary.Read(bytes.NewReader(helloData), binary.LittleEndian, hello); err != nil {
		return nil, err
	}
	if hello.Ver != Ver0 {
		return nil, fmt.Errorf("bad version in hello %x", hello.Ver)
	}
	if !bytes.Equal(hello.Network[:], network) {
		return nil, fmt.Errorf("hello for a different network")
	}
	if !ed25519.Verify(ed25519.PublicKey(hello.User[:]), hello.Nonce[:], hello.Sign[:]) {
		return nil, fmt.Errorf("bad hello signature")
	}
	// Send ack.
	ack := &HelloAck{
		Ver: Ver0,
		Sync: Sync{
			Flags: SyncFollow,
		},
	}
	if getCursor != nil {
		if cursor := getCursor(&hello.User); cursor != nil {
			ack.Flags |= SyncCursor
			ack.Cursor = *cursor
		}
	}
	// TODO: invoke a callback to get Sync for the user.
	sign := ed25519.Sign(ed25519.NewKeyFromSeed(key), hello.Nonce[:])
	if len(ack.Sign) != len(sign) {
		return nil, fmt.Errorf("signature size mismatch: %v/%v", len(ack.Sign), len(sign))
	}
	copy(ack.Sign[:], sign)
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, ack)
	if _, err := c.Write(buf.Bytes()); err != nil {
		return nil, err
	}
	return newConn(c, hello.User, hello.Sync)
}

func newConn(c net.Conn, user UserID, sync Sync) (*Conn, error) {
	conn := &Conn{
		c:          c,
		r:          bufio.NewReader(c),
		w:          bufio.NewWriter(c),
		RemoteUser: user,
		RemoteSync: sync,
	}
	return conn, nil
}

func (conn *Conn) Recv() (*Raw, error) {
	return Deserialize(conn.r)
}

func (conn *Conn) Send(msg *Raw) error {
	if err := Serialize(conn.w, msg); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	return conn.w.Flush()
}

func (conn *Conn) SendMany(msgs []*Raw) error {
	for _, msg := range msgs {
		if err := Serialize(conn.w, msg); err != nil {
			return fmt.Errorf("failed to write message: %v", err)
		}
	}
	return conn.w.Flush()
}

func (conn *Conn) Close() error {
	return conn.c.Close()
}

func (conn *Conn) CloseWrite() error {
	return conn.c.(*net.TCPConn).CloseWrite()
}

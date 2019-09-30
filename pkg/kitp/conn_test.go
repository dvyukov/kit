package kitp

import (
	"crypto/ed25519"
	"io"
	"math/rand"
	"testing"
)

func TestMsg(t *testing.T) {
	rnd := rand.New(rand.NewSource(0))
	pubKey, privKey, err := ed25519.GenerateKey(rnd)
	if err != nil {
		t.Fatal(err)
	}
	seed := privKey.Seed()
	msg := &Raw{
		Hdr: Hdr{
			Seq:  42,
			Time: 12345,
			Type: TypeRegister,
		},
		Data: []byte("this is data"),
	}
	copy(msg.User[:], pubKey)
	io.ReadFull(rnd, msg.Prev[:])
	io.ReadFull(rnd, msg.Link[:])
	if err := msg.Seal(seed); err != nil {
		t.Fatalf("failed to seal: %v", err)
	}
}

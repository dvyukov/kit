package journal

import (
	"crypto/ed25519"
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/dvyukov/kit/pkg/kitp"
)

func TestJournal(t *testing.T) {
	os.Remove("/tmp/kit/data")
	store, err := Open("/tmp/kit")
	if err != nil {
		t.Fatal(err)
	}
	rnd := rand.New(rand.NewSource(0))
	pubKey, privKey, err := ed25519.GenerateKey(rnd)
	if err != nil {
		t.Fatal(err)
	}
	msgs := []*kitp.Raw{
		{
			Hdr: kitp.Hdr{
				Seq:  0,
				Type: kitp.TypeRegister,
			},
			/*
				Data: &kitp.MsgRegister{
					Name: "foo",
				},
			*/
		},
		{
			Hdr: kitp.Hdr{
				Seq:  1,
				Prev: kitp.MsgID{1},
				Type: kitp.TypeRegister,
			},
			/*
				Data: &kitp.MsgChange{
					Name: "change",
				},
			*/
		},
	}

	for i, msg := range msgs {
		copy(msg.User[:], pubKey)
		if i != 0 {
			copy(msg.Prev[:], msgs[i-1].ID[:])
		}
		if err := msg.Seal(privKey.Seed()); err != nil {
			t.Fatalf("failed to seal: %v", err)
		}
		if err := store.Append(msg); err != nil {
			t.Fatal(err)
		}
	}
	msgs1, err := store.Read()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(msgs1, msgs) {
		t.Fatalf("got:  %#v\nwant: %#v", msgs1, msgs)
	}
	//t.Logf("%#v", *msgs[0])
}

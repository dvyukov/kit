package journal

import (
	"os"
	"reflect"
	"testing"

	"github.com/dvyukov/kit/pkg/kitp"
)

func TestJournal(t *testing.T) {
	os.Remove("/tmp/kit/data")
	store, err := New("/tmp/kit")
	if err != nil {
		t.Fatal(err)
	}
	msgs := []*kitp.Msg{
		{
			Hdr: kitp.Hdr{
				Seq: 42,
			},
			Data: &kitp.MsgRegister{
				Name: "foo",
			},
		},
		{
			Hdr: kitp.Hdr{
				Seq: 43,
			},
			Data: &kitp.MsgChange{
				Name: "change",
			},
		},
	}
	for _, msg := range msgs {
		if err := store.Write(msg); err != nil {
			t.Fatal(err)
		}
	}
	msgs1, err := store.ReadUnraw()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(msgs1, msgs) {
		t.Fatalf("got:  %#v\nwant: %#v", msgs1, msgs)
	}
	//t.Logf("%#v", *msgs[0])
}

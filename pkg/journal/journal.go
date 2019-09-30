package journal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/dvyukov/kit/pkg/kitp"
)

type Journal struct {
	dir   string
	users string
	data  string
}

func New(dir string) (*Journal, error) {
	store := &Journal{
		dir:   dir,
		users: filepath.Join(dir, "users"),
		data:  filepath.Join(dir, "data"),
	}
	if err := os.MkdirAll(store.users, 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(store.data, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	f.Close()
	users, err := ioutil.ReadDir(store.users)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		fmt.Printf("user: %q\n", user.Name())
	}
	return store, nil
}

func (store *Journal) Read() ([]*kitp.Raw, error) {
	f, err := os.Open(store.data)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bf := bufio.NewReader(f)
	var msgs []*kitp.Raw
	for {
		msg := new(kitp.Raw)
		if err := binary.Read(bf, binary.LittleEndian, &msg.Hdr); err != nil {
			if err != io.EOF {
				return nil, err
			}
			return msgs, nil
		}
		msg.Data = make([]byte, msg.Size)
		if _, err := io.ReadFull(bf, msg.Data); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
}

func (journal *Journal) ReadUnraw() ([]*kitp.Msg, error) {
	raws, err := journal.Read()
	if err != nil {
		return nil, err
	}
	msgs := make([]*kitp.Msg, len(raws))
	for i, raw := range raws {
		msg := new(kitp.Msg)
		msg.Hdr = raw.Hdr
		switch raw.Type {
		case kitp.TypeRegister:
			msg.Data = new(kitp.MsgRegister)
		case kitp.TypeChange:
			msg.Data = new(kitp.MsgChange)
		}
		if err := json.Unmarshal(raw.Data, msg.Data); err != nil {
			return nil, err
		}
		msgs[i] = msg
	}
	return msgs, nil
}

func (journal *Journal) Append(msg *kitp.Raw) error {
	f, err := os.OpenFile(journal.data, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.LittleEndian, &msg.Hdr); err != nil {
		return err
	}
	buf.Write(msg.Data)
	if _, err := f.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (journal *Journal) Write(msg *kitp.Msg) error {
	msg.Ver = kitp.Ver0
	msg.Time = uint64(time.Now().Unix())
	switch msg.Data.(type) {
	case *kitp.MsgRegister:
		msg.Type = kitp.TypeRegister
	case *kitp.MsgChange:
		msg.Type = kitp.TypeChange
	}
	raw, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}
	msg.Size = uint32(len(raw))
	rawMsg := &kitp.Raw{
		Hdr:  msg.Hdr,
		Data: raw,
	}
	return journal.Append(rawMsg)
}

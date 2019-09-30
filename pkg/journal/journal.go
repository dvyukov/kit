package journal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dvyukov/kit/pkg/kitp"
)

type Journal struct {
	dir     string
	userDir string
	data    string

	msgs  []*kitp.Raw
	users map[kitp.UserID]userInfo
}

type userInfo struct {
	nextSeq uint64
	lastMsg kitp.MsgID
}

func Open(dir string) (*Journal, error) {
	journal := &Journal{
		dir:     dir,
		userDir: filepath.Join(dir, "users"),
		data:    filepath.Join(dir, "data"),
		users:   make(map[kitp.UserID]userInfo),
	}
	if err := os.MkdirAll(journal.userDir, 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(journal.data, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	f.Close()
	/*
		users, err := ioutil.ReadDir(journal.userDir)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			fmt.Printf("user: %q\n", user.Name())
		}
	*/
	msgs, err := journal.read()
	if err != nil {
		return nil, err
	}
	for _, msg := range msgs {
		dup, err := journal.checkAppend(msg)
		if err != nil {
			return nil, err
		}
		if dup {
			return nil, fmt.Errorf("duplicate message in journal")
		}
		journal.append(msg)
	}
	return journal, nil
}

// TODO: review passing by value/pointer/slice?
func (journal *Journal) PeerCursor(user kitp.UserID) (*kitp.MsgID, error) {
	file := filepath.Join(journal.userDir, fmt.Sprintf("%x", user[:]))
	data, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	msgID := new(kitp.MsgID)
	copy((*msgID)[:], data)
	return msgID, nil
}

func (journal *Journal) SelectCursor(peer kitp.UserID, sync kitp.Sync) []*kitp.Raw {
	msgs := journal.msgs
	if sync.Flags&kitp.SyncCursor != 0 {
		for i := len(msgs) - 1; i >= 0; i-- {
			if bytes.Equal(msgs[i].ID[:], sync.Cursor[:]) {
				//.fmt.Printf("cursor %x: selected %v/%v\n", sync.Cursor[:4], len(msgs)-i-1, len(msgs))
				msgs = msgs[i+1:]
				break
			}
		}
		// TODO: handle the case if the message is not found.
		// What do we want to do? Sending everything may be unwise.
		// Would need to notify remote side to send us vector clock.
	}
	return msgs
}

func (journal *Journal) updateCursor(user *kitp.UserID, msg *kitp.MsgID) error {
	file := filepath.Join(journal.userDir, fmt.Sprintf("%x", (*user)[:]))
	return ioutil.WriteFile(file, (*msg)[:], 0644)
}

func (journal *Journal) Me() (user kitp.UserID, msgID kitp.MsgID, seq uint64, err error) {
	if len(journal.msgs) == 0 {
		err = fmt.Errorf("journal is empty")
	}
	user = journal.msgs[0].User
	ui := journal.users[user]
	msgID = ui.lastMsg
	seq = ui.nextSeq
	return
}

func (journal *Journal) checkAppend(msg *kitp.Raw) (bool, error) {
	ui := journal.users[msg.User]
	// TODO: fix the check.
	if msg.Seq < ui.nextSeq {
		return true, nil
	}
	if msg.Seq != ui.nextSeq {
		return false, fmt.Errorf("user %x: msg has bad seq %v, expect %v", msg.User, msg.Seq, ui.nextSeq)
	}
	if !bytes.Equal(msg.Prev[:], ui.lastMsg[:]) {
		return false, fmt.Errorf("user %x: msg has bad prev %x, expect %x", msg.User, msg.Prev, ui.lastMsg)
	}
	return false, nil
}

func (journal *Journal) append(msg *kitp.Raw) {
	ui := journal.users[msg.User]
	ui.nextSeq = msg.Seq + 1
	ui.lastMsg = msg.ID
	journal.users[msg.User] = ui
	journal.msgs = append(journal.msgs, msg)
}

func (journal *Journal) read() ([]*kitp.Raw, error) {
	f, err := os.Open(journal.data)
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

func (journal *Journal) Read() ([]*kitp.Raw, error) {
	return journal.msgs, nil
}

func (journal *Journal) Messages() []*kitp.Raw {
	return journal.msgs
}

func (journal *Journal) Append(msg *kitp.Raw) error {
	ok, err := journal.AppendDup(msg, nil)
	if err == nil && !ok {
		err = fmt.Errorf("duplicate message")
	}
	return err
}

func (journal *Journal) AppendDup(msg *kitp.Raw, from *kitp.UserID) (bool, error) {
	if err := msg.Verify(); err != nil {
		return false, err
	}
	// TODO: verify Seq/Prev/Link
	dup, err := journal.checkAppend(msg)
	if err != nil {
		return false, err
	}
	if !dup {
		f, err := os.OpenFile(journal.data, os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			return false, err
		}
		defer f.Close()
		if err := kitp.Serialize(f, msg); err != nil {
			return false, err
		}
		// TODO: do we need to do a copy?
		journal.append(msg)
	}
	if from != nil {
		err := journal.updateCursor(from, &msg.ID)
		if err != nil {
			// TODO: the cursor is only an optimization
			// and the message is already in the journal.
			fmt.Fprintf(os.Stderr, "journal: failed to update cursor for user %x: %v\n", (*from)[:4], err)
		}
	}
	return !dup, nil
}

func isZero(x []byte) bool {
	for _, v := range x {
		if v != 0 {
			return false
		}
	}
	return true
}

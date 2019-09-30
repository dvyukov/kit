package kitp

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func (msg *Raw) Verify() error {
	if msg.Ver != Ver0 {
		return fmt.Errorf("bad version 0x%x", msg.Ver)
	}
	if msg.Size > MaxDataSize {
		return fmt.Errorf("too large data %v", msg.Size)
	}
	if msg.Size != uint32(len(msg.Data)) {
		return fmt.Errorf("data size mismatch: %v/%v", msg.Size, len(msg.Data))
	}
	if msg.Type == 0 || msg.Type >= typeCount {
		return fmt.Errorf("bad type: %v", msg.Type)
	}
	if msg.Time == 0 {
		return fmt.Errorf("no time")
	}
	if isZero(msg.User[:]) {
		return fmt.Errorf("no user")
	}
	if isZero(msg.ID[:]) {
		return fmt.Errorf("no id")
	}
	if msg.Seq == 0 != isZero(msg.Prev[:]) {
		return fmt.Errorf("seq/prev mismatch: %v/%v", msg.Seq, msg.Prev)
	}
	return msg.checkSign()
}

func (msg *Raw) checkSign() error {
	msg1 := *msg
	for i := range msg1.ID {
		msg1.ID[i] = 0
	}
	for i := range msg1.Sign {
		msg1.Sign[i] = 0
	}
	data, err := msg1.serialize()
	if err != nil {
		return err
	}
	if !ed25519.Verify(ed25519.PublicKey(msg.User[:]), data, msg.Sign[:]) {
		return fmt.Errorf("bad message signature")
	}
	msg1.Sign = msg.Sign
	data, err = msg1.serialize()
	if err != nil {
		return err
	}
	id := sha256.Sum256(data)
	if !bytes.Equal(msg.ID[:], id[:]) {
		return fmt.Errorf("bad message id")
	}
	return nil
}

func (msg *Raw) Seal(key []byte) error {
	msg.Ver = Ver0
	if msg.Size == 0 {
		msg.Size = uint32(len(msg.Data))
	}
	if msg.Time == 0 {
		msg.Time = uint64(time.Now().Unix())
	}
	if !isZero(msg.ID[:]) {
		return fmt.Errorf("non zero id when signing")
	}
	if !isZero(msg.Sign[:]) {
		return fmt.Errorf("non zero sign when signing")
	}
	data, err := msg.serialize()
	if err != nil {
		return err
	}
	sign := ed25519.Sign(ed25519.NewKeyFromSeed(key), data)
	if len(msg.Sign) != len(sign) {
		return fmt.Errorf("signature size mismatch: %v/%v", len(msg.Sign), len(sign))
	}
	copy(msg.Sign[:], sign)
	data, err = msg.serialize()
	if err != nil {
		return err
	}
	id := sha256.Sum256(data)
	if len(msg.ID) != len(id) {
		return fmt.Errorf("id size mismatch: %v/%v", len(msg.ID), len(id))
	}
	copy(msg.ID[:], id[:])
	return msg.Verify()
}

func Deserialize(r io.Reader) (*Raw, error) {
	msg := new(Raw)
	if err := binary.Read(r, binary.LittleEndian, &msg.Hdr); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	if msg.Size > MaxDataSize {
		return nil, fmt.Errorf("too large data %v", msg.Size)
	}
	msg.Data = make([]byte, msg.Size)
	if _, err := io.ReadFull(r, msg.Data); err != nil {
		return nil, err
	}
	if err := msg.Verify(); err != nil {
		return nil, err
	}
	return msg, nil
}

func Serialize(w io.Writer, msg *Raw) error {
	if err := msg.Verify(); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &msg.Hdr); err != nil {
		return err
	}
	if _, err := w.Write(msg.Data); err != nil {
		return err
	}
	return nil
}

func (msg *Raw) serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.LittleEndian, &msg.Hdr); err != nil {
		return nil, err
	}
	if _, err := buf.Write(msg.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (msg *Raw) AddData(data interface{}) error {
	switch data.(type) {
	case *MsgRegister:
		msg.Type = TypeRegister
	case *MsgChange:
		msg.Type = TypeChange
	default:
		return fmt.Errorf("unknown data type %#v", data)
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg.Data = raw
	return nil
}

func (msg *Raw) GetData() (interface{}, error) {
	var data interface{}
	switch msg.Type {
	case TypeRegister:
		data = new(MsgRegister)
	case TypeChange:
		data = new(MsgChange)
	default:
		return nil, fmt.Errorf("bad msg type %v", msg.Type)
	}
	if err := json.Unmarshal(msg.Data, data); err != nil {
		return nil, err
	}
	return data, nil
}

func isZero(x []byte) bool {
	for _, v := range x {
		if v != 0 {
			return false
		}
	}
	return true
}

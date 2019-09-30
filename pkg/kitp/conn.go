package kitp

import (
	"encoding/binary"
	"fmt"
	"io"
)

func Deserialize(r io.Reader) (*Raw, error) {
	msg := new(Raw)
	if err := binary.Read(r, binary.LittleEndian, &msg.Hdr); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	if msg.Ver != Ver0 {
		return nil, fmt.Errorf("bad version 0x%x", msg.Ver)
	}
	if msg.Size > MaxDataSize {
		return nil, fmt.Errorf("too large data %v", msg.Size)
	}
	msg.Data = make([]byte, msg.Size)
	if _, err := io.ReadFull(r, msg.Data); err != nil {
		return nil, err
	}
	return msg, nil
}

func Serialize(w io.Writer, msg *Raw) error {
	msg.Ver = Ver0
	msg.Size = uint32(len(msg.Data))
	if msg.Size > MaxDataSize {
		return fmt.Errorf("too large data %v", msg.Size)
	}
	if err := binary.Write(w, binary.LittleEndian, &msg.Hdr); err != nil {
		return err
	}
	if _, err := w.Write(msg.Data); err != nil {
		return err
	}
	return nil
}

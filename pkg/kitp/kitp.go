package kitp

import (
	"encoding/hex"
	"fmt"
)

type (
	MsgID     [32]byte // SHA-256
	UserID    [32]byte // Ed25519 public key
	Signature [64]byte // Ed25519 signature
	Version   uint64
	SyncFlags uint64
	MsgType   uint32

	Hello struct {
		Ver     Version
		Network UserID
		User    UserID
		Nonce   [64]byte
		Sign    Signature
		Sync
	}

	HelloAck struct {
		Ver Version
		//User UserID
		Sign Signature
		Sync
	}

	Sync struct {
		Flags  SyncFlags
		Cursor MsgID
		Size   uint64
		// Followed by Size ClockElem's.
	}

	ClockElem struct {
		User UserID
		Seq  uint64
	}

	Hdr struct {
		Ver  Version
		Sign Signature
		User UserID
		ID   MsgID
		Prev MsgID
		Link MsgID // TODO: do we need this here?
		Seq  uint64
		Time uint64
		Type MsgType
		Size uint32
	}

	// TODO: rename to Msg.
	Raw struct {
		Hdr
		Data []byte
	}

	MsgRegister struct {
		Email string
	}

	MsgChange struct {
		Diff string
	}
)

const (
	Ver0        Version = 0x000000000074696b
	MaxDataSize         = 64 << 10 // let's start small and safe
)

const (
	typeNone MsgType = iota
	TypeRegister
	TypeChange
	typeCount
)

const (
	SyncCursor SyncFlags = 1 << iota
	SyncAll
	SyncFollow
)

func (msg *MsgID) String() string {
	return hex.EncodeToString((*msg)[:])
}

func (user *UserID) String() string {
	return hex.EncodeToString((*user)[:])
}

func (sign *Signature) String() string {
	return hex.EncodeToString((*sign)[:])
}

func (t MsgType) String() string {
	switch t {
	case TypeRegister:
		return "register"
	case TypeChange:
		return "change"
	default:
		return fmt.Sprintf("bad(%v)", uint32(t))
	}
}

func (msg *Raw) String() string {
	text := ""
	if len(msg.Data) <= 30 {
		text = string(msg.Data)
	} else {
		text = string(msg.Data[:27]) + "..."
	}
	const w = 4
	return fmt.Sprintf("u:%x i:%x s:%v t:%-8v sz:%v %s",
		msg.User[:w], msg.ID[:w], msg.Seq, msg.Type, msg.Size, text)
}

package kitp

import (
	"encoding/hex"
)

type (
	MsgID     [32]byte // SHA-256
	UserID    [32]byte // Ed25519 public key
	Signature [64]byte // Ed25519 signature

	/*
		Hello struct {
			Ver  uint64
			User UserID
		}

		HelloAck struct {
			Ver  uint64
			User UserID
		}

		Auth struct {
			Ver uint64
			Seq uint64
		}

		AuthAck struct {
			Ver uint64
			Seq uint64
		}
	*/

	Hdr struct {
		Ver  uint64
		Sign Signature
		User UserID
		ID   MsgID
		Prev MsgID
		Link MsgID
		Seq  uint64
		Time uint64
		Type uint32
		Size uint32
	}

	Raw struct {
		Hdr
		Data []byte
	}

	Msg struct {
		Hdr
		Data interface{}
	}

	MsgRegister struct {
		Name string
		// ...
	}

	MsgChange struct {
		Name string
		// ...
	}
)

const (
	Ver0        = 0x000000000074696b
	MaxDataSize = 64 << 10 // let's start small and safe
)

const (
	TypeRegister = iota
	TypeChange
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

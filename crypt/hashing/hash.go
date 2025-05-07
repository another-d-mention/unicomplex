package hashing

import (
	"encoding/hex"
	"io"
)

type Sum []byte

func (s Sum) HexEncoded() string {
	return hex.EncodeToString(s)
}

func (s Sum) String() string {
	return hex.EncodeToString(s)
}

type HashWriter interface {
	io.Writer
	Sum() Sum
}

type Hasher interface {
	Reset()
	Bytes(data []byte) Sum
	Reader(r io.Reader) Sum
	GetWriter(dest io.Writer) HashWriter
}

package hashing

import (
	"hash"
	"hash/crc32"
	"hash/crc64"
	"io"
)

type CRCHasher struct {
	hasher hash.Hash
}

func NewCRC32Hasher() CRCHasher {
	return CRCHasher{
		hasher: crc32.NewIEEE(),
	}
}

func NewCRC64HAsher() CRCHasher {
	return CRCHasher{
		hasher: crc64.New(crc64.MakeTable(crc64.ISO)),
	}
}

func (h CRCHasher) Reset() {
	h.hasher.Reset()
}

func (h CRCHasher) Bytes(data []byte) Sum {
	h.hasher.Write(data)
	return h.hasher.Sum(nil)
}

func (h CRCHasher) Reader(r io.Reader) Sum {
	_, _ = io.Copy(h.hasher, r)
	return h.hasher.Sum(nil)
}

func (h CRCHasher) GetWriter(w io.Writer) HashWriter {
	return crcHashWriter{
		writer: w,
		hasher: h.hasher,
	}
}

type crcHashWriter struct {
	writer io.Writer
	hasher hash.Hash
}

func (h crcHashWriter) Write(p []byte) (n int, err error) {
	h.hasher.Write(p)
	return h.writer.Write(p)
}

func (h crcHashWriter) Sum() Sum {
	return h.hasher.Sum(nil)
}

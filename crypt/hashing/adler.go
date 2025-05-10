package hashing

import (
	"hash"
	"hash/adler32"
	"io"
)

type AdlerHasher struct {
	hasher hash.Hash
}

func NewAdler32Hasher() AdlerHasher {
	return AdlerHasher{
		hasher: adler32.New(),
	}
}

func (h AdlerHasher) Reset() {
	h.hasher.Reset()
}

func (h AdlerHasher) Bytes(data []byte) Sum {
	h.hasher.Write(data)
	return h.hasher.Sum(nil)
}

func (h AdlerHasher) Reader(r io.Reader) Sum {
	_, _ = io.Copy(h.hasher, r)
	return h.hasher.Sum(nil)
}

func (h AdlerHasher) GetWriter(w io.Writer) HashWriter {
	return adlerHashWriter{
		writer: w,
		hasher: h.hasher,
	}
}

type adlerHashWriter struct {
	writer io.Writer
	hasher hash.Hash
}

func (h adlerHashWriter) Write(p []byte) (n int, err error) {
	h.hasher.Write(p)
	return h.writer.Write(p)
}

func (h adlerHashWriter) Sum() Sum {
	return h.hasher.Sum(nil)
}

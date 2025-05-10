package hashing

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"io"
)

type ShaHasher struct {
	hasher hash.Hash
}

func NewSHA1Hasher() ShaHasher {
	return ShaHasher{
		hasher: sha1.New(),
	}
}

func NewSHA256Hasher() ShaHasher {
	return ShaHasher{
		hasher: sha256.New(),
	}
}

func NewSHA384Hasher() ShaHasher {
	return ShaHasher{
		hasher: sha512.New384(),
	}
}

func NewSHA512Hasher() ShaHasher {
	return ShaHasher{
		hasher: sha512.New(),
	}
}

func (h ShaHasher) Reset() {
	h.hasher.Reset()
}

func (h ShaHasher) Bytes(data []byte) Sum {
	h.hasher.Write(data)
	return h.hasher.Sum(nil)
}

func (h ShaHasher) Reader(r io.Reader) Sum {
	_, _ = io.Copy(h.hasher, r)
	return h.hasher.Sum(nil)
}

func (h ShaHasher) GetWriter(w io.Writer) HashWriter {
	return shaHashWriter{writer: w, hasher: h.hasher}
}

type shaHashWriter struct {
	writer io.Writer
	hasher hash.Hash
}

func (h shaHashWriter) Write(p []byte) (n int, err error) {
	h.hasher.Write(p)
	return h.writer.Write(p)
}

func (h shaHashWriter) Sum() Sum {
	return h.hasher.Sum(nil)
}

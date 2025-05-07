package hashing

import (
	"crypto/md5"
	"hash"
	"io"
)

type MD5Hasher struct {
	hasher hash.Hash
}

func NewMD5Hasher() MD5Hasher {
	return MD5Hasher{
		hasher: md5.New(),
	}
}

func (h MD5Hasher) Reset() {
	h.hasher.Reset()
}

func (h MD5Hasher) Bytes(data []byte) Sum {
	h.hasher.Write(data)
	return h.hasher.Sum(nil)
}

func (h MD5Hasher) Reader(r io.Reader) Sum {
	_, _ = io.Copy(h.hasher, r)
	return h.hasher.Sum(nil)
}

func (h MD5Hasher) GetWriter(w io.Writer) HashWriter {
	return md5HashWriter{
		writer: w,
		hasher: h.hasher,
	}
}

type md5HashWriter struct {
	writer io.Writer
	hasher hash.Hash
}

func (h md5HashWriter) Write(p []byte) (n int, err error) {
	h.hasher.Write(p)
	return h.writer.Write(p)
}

func (h md5HashWriter) Sum() Sum {
	return h.hasher.Sum(nil)
}

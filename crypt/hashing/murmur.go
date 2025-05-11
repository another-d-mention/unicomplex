package hashing

import (
	"encoding/binary"
	"io"
)

type murmurHasher struct {
	data []byte
	d    digest128
	full bool
}

func NewMurmur128Hasher() Hasher {
	return &murmurHasher{}
}

func NewMurmur256Hasher() Hasher {
	return &murmurHasher{
		full: true,
	}
}

func (h *murmurHasher) Reset() {
	h.data = nil
}

func (h *murmurHasher) Bytes(data []byte) Sum {
	h.data = data
	return h.Sum()
}

func (h *murmurHasher) Reader(r io.Reader) Sum {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil
	}
	h.data = data
	return h.Sum() // Assuming Sum() returns the same type as Bytes()
}

func (h *murmurHasher) GetWriter(w io.Writer) HashWriter {
	return &murmurHashWriter{writer: w, h: h}
}

func (h *murmurHasher) Sum() Sum {
	v1, v2, v3, v4 := h.d.sum256(h.data)
	if h.full {
		var b = make([]byte, 32)
		binary.BigEndian.AppendUint64(b[:0], v1)
		binary.BigEndian.AppendUint64(b[:8], v2)
		binary.BigEndian.AppendUint64(b[:16], v3)
		binary.BigEndian.AppendUint64(b[:24], v4)
		return Sum(b)
	} else {
		var b = make([]byte, 16)
		binary.BigEndian.AppendUint64(b[:0], v1)
		binary.BigEndian.AppendUint64(b[:8], v2)
		return Sum(b)
	}
}

type murmurHashWriter struct {
	writer io.Writer
	h      *murmurHasher
}

func (w *murmurHashWriter) Write(p []byte) (n int, err error) {
	w.h.data = append(w.h.data, p...)
	return len(p), nil
}

func (w *murmurHashWriter) Sum() Sum {
	return w.h.Sum()
}

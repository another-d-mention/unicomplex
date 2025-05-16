package gcm

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

const (
	DefaultChunkSize = 64 * 1024
)

type GCM struct {
	rw   io.ReadWriter
	aead cipher.AEAD
}

type Password = cipher.AEAD

func NewPassword(key []byte) (Password, error) {
	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aead, nil
}

func New(password Password) (*GCM, error) {
	return &GCM{aead: password}, nil
}

func (g *GCM) Encrypt(p []byte) ([]byte, error) {
	nonce := make([]byte, g.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := g.aead.Seal(nil, nonce, p, nil)
	out := make([]byte, len(ciphertext)+g.aead.NonceSize())
	copy(out, nonce)
	copy(out[g.aead.NonceSize():], ciphertext)
	return out, nil
}

func (g *GCM) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < g.aead.NonceSize() {
		return nil, io.ErrUnexpectedEOF
	}
	nonce := ciphertext[:g.aead.NonceSize()]
	ciphertext = ciphertext[g.aead.NonceSize():]
	return g.aead.Open(nil, nonce, ciphertext, nil)
}

type StreamWriter struct {
	*GCM
	writer    io.Writer
	writeBuf  *bytes.Buffer
	chunkSize int
}

func NewStreamWriter(password Password, w io.Writer, chunkSize int) (*StreamWriter, error) {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	chunkSize = chunkSize * 1024

	g, err := New(password)
	if err != nil {
		return nil, err
	}
	return &StreamWriter{GCM: g, writer: w, chunkSize: chunkSize, writeBuf: &bytes.Buffer{}}, nil
}

func (e *StreamWriter) Write(p []byte) (int, error) {
	e.writeBuf.Write(p)

	nonceSize := e.aead.NonceSize()
	tagSize := 16 // AES-GCM tag size
	overhead := nonceSize + 4 + tagSize
	maxPlaintext := e.chunkSize - overhead
	if maxPlaintext <= 0 {
		return 0, errors.New("chunk size too small for encryption framing")
	}

	written := len(p)

	for e.writeBuf.Len() >= maxPlaintext {
		chunk := make([]byte, maxPlaintext)
		if _, err := e.writeBuf.Read(chunk); err != nil {
			return written, err
		}
		if _, err := e.writeChunk(chunk); err != nil {
			return written, err
		}
	}
	return written, nil
}

func (e *StreamWriter) Flush() error {
	if e.writeBuf.Len() == 0 {
		return nil
	}
	chunk := e.writeBuf.Bytes()
	e.writeBuf.Reset()
	_, err := e.writeChunk(chunk)
	return err
}

func (e *StreamWriter) writeChunk(chunk []byte) (int, error) {
	nonceSize := e.aead.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return 0, err
	}

	ciphertext := e.aead.Seal(nil, nonce, chunk, nil)

	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(ciphertext)))

	if _, err := e.writer.Write(nonce); err != nil {
		return 0, err
	}
	if _, err := e.writer.Write(lenBuf[:]); err != nil {
		return 0, err
	}
	if _, err := e.writer.Write(ciphertext); err != nil {
		return 0, err
	}
	return len(chunk), nil
}

type StreamReader struct {
	*GCM
	reader  io.Reader
	readBuf *bytes.Buffer
}

func NewStreamReader(password Password, r io.Reader) (*StreamReader, error) {
	g, err := New(password)
	if err != nil {
		return nil, err
	}
	return &StreamReader{GCM: g, reader: r, readBuf: &bytes.Buffer{}}, nil
}

func (e *StreamReader) Read(p []byte) (int, error) {
	if e.readBuf.Len() == 0 {
		nonceSize := e.aead.NonceSize()
		nonce := make([]byte, nonceSize)
		if _, err := io.ReadFull(e.reader, nonce); err != nil {
			return 0, err
		}

		var lenBuf [4]byte
		if _, err := io.ReadFull(e.reader, lenBuf[:]); err != nil {
			return 0, err
		}
		ciphertextLen := binary.BigEndian.Uint32(lenBuf[:])
		ciphertext := make([]byte, ciphertextLen)
		if _, err := io.ReadFull(e.reader, ciphertext); err != nil {
			return 0, err
		}

		plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return 0, err
		}
		e.readBuf.Write(plaintext)
	}

	return e.readBuf.Read(p)
}

type Stream struct {
	*GCM
	*StreamReader
	*StreamWriter
}

func NewStream(password Password, rw io.ReadWriter, chunkSize int) (*Stream, error) {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	chunkSize = chunkSize * 1024

	g, err := New(password)
	if err != nil {
		return nil, err
	}
	return &Stream{
		GCM:          g,
		StreamReader: &StreamReader{GCM: g, reader: rw},
		StreamWriter: &StreamWriter{GCM: g, writer: rw, chunkSize: chunkSize},
	}, nil
}

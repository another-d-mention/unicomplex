package gcm

import (
	"bytes"
	"io"
	"testing"
)

func TestGCM(t *testing.T) {
	password, err := NewPassword([]byte("test-password"))
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := New(password)
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("Hello, GCM!")
	ciphertext, err := gcm.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	dectext, err := gcm.Decrypt(ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, dectext) {
		t.Error("Decrypted text does not match original plaintext")
	}

	plaintext = []byte("Hello, GCM!, Hello, GCM!, Hello, GCM!, Hello, GCM!, Hello, GCM!")
	buf := bytes.NewBuffer(nil)
	w, err := NewStreamWriter(password, buf, 32)
	if err != nil {
		t.Fatal(err)
	}
	n, err := w.Write(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(plaintext) {
		t.Error("Expected to write all plaintext, but only wrote", n)
	}
	if e := w.Flush(); e != nil {
		t.Fatal(e)
	}

	r, err := NewStreamReader(password, bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted text does not match original plaintext")
	}
}

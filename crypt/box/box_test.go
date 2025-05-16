package box

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestBox(t *testing.T) {
	uid := uuid.New()
	k, ok := KeyFromUUID(uid)
	if !ok {
		t.Error("failed to generate key")
	}
	msg := []byte("Hello, Unicomplex!")
	enc, ok := k.Encrypt(msg)
	if !ok {
		t.Error("failed to encrypt message")
	}
	dec, ok := k.Decrypt(enc)
	if !ok {
		t.Error("failed to decrypt message")
	}
	if !bytes.Equal(msg, dec) {
		t.Error("encrypted and decrypted messages do not match")
	}
}

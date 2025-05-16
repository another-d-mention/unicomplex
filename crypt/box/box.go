package box

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
	"github.com/google/uuid"
)

const cryptKeySize = 32
const tagKeySize = 32

// KeySize is the number of bytes a valid key should be.
const KeySize = cryptKeySize + tagKeySize

// Overhead is the number of bytes of overhead when boxing a message.
const Overhead = aes.BlockSize + sha512.Size

var (
	errInvalidCipherText = fmt.Errorf("invalid ciphertext")
)

// The default source for random data is the crypto/rand package's Reader.
var pRNG = rand.Reader

type nonce []byte

// Key is the encryption key used by the box to encrypt/decrypt data
type Key []byte

// String returns a hex encoded string of the Key
func (k Key) String() string {
	return fmt.Sprintf("%x", []byte(k))
}

// Encrypt uses the key to encrypt the given message and returns the encrypted data and a boolean indicating if the
// operation succeeded
func (k Key) Encrypt(msg []byte) (box []byte, ok bool) {
	return Seal(msg, k)
}

// EncryptString uses the key to encrypt the given message and returns the encrypted data and a boolean indicating if the
// operation succeeded
func (k Key) EncryptString(msg []byte) (box []byte, ok bool) {
	return Seal(msg, k)
}

// DecryptString uses the key to encrypt the given message and returns the decrypted data and a boolean indicating if the
// operation succeeded
func (k Key) DecryptString(box string) (msg []byte, ok bool) {
	if len(box) == 0 {
		return nil, false
	}
	b, err := hex.DecodeString(box)
	if err != nil {
		return nil, false
	}
	msg, ok = Open(b, k)
	return
}

// Decrypt uses the key to encrypt the given message and returns the decrypted data and a boolean indicating if the
// operation succeeded
func (k Key) Decrypt(msg []byte) (box []byte, ok bool) {
	return Open(msg, k)
}

// KeyFromString parses a hex encoded string and verifies if it can be used as a Key
func KeyFromString(val string) (Key, bool) {
	if val == "" {
		return nil, false
	}
	b, err := hex.DecodeString(val)
	if err != nil {
		return nil, false
	}
	key := Key(b)
	if KeyIsSuitable(key) {
		return key, true
	}
	return nil, false
}

func KeyFromUUID(val uuid.UUID) (Key, bool) {
	hash := hashing.NewSHA512Hasher().Bytes(val[:])
	k := Key(hash[:KeySize])
	if !KeyIsSuitable(k) {
		return nil, false
	}
	return k, true
}

// GenerateKey returns a key suitable for sealing and opening boxes, and
// a boolean indicating success. If the boolean returns false, the Key
// value must be discarded.
func GenerateKey() (Key, bool) {
	var key Key = make([]byte, KeySize)
	_, err := io.ReadFull(pRNG, key)
	return key, err == nil
}

func generateNonce() (nonce, error) {
	var n nonce = make([]byte, aes.BlockSize)

	_, err := io.ReadFull(pRNG, n)
	return n, err
}

func encrypt(key []byte, in []byte) (out []byte, err error) {
	var iv nonce
	if iv, err = generateNonce(); err != nil {
		return
	}

	out = make([]byte, len(in)+aes.BlockSize)
	for i := 0; i < aes.BlockSize; i++ {
		out[i] = iv[i]
		iv[i] = 0
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	ctr := cipher.NewCTR(c, out[:aes.BlockSize])
	ctr.XORKeyStream(out[aes.BlockSize:], in)
	return
}

func computeTag(key []byte, in []byte) (tag []byte) {
	h := hmac.New(sha512.New, key)
	h.Write(in)
	return h.Sum(nil)
}

func checkTag(key, in []byte) bool {
	ctlen := len(in) - sha512.Size
	tag := in[ctlen:]
	ct := in[:ctlen]
	actualTag := computeTag(key, ct)
	return subtle.ConstantTimeCompare(tag, actualTag) == 1
}

func decrypt(key []byte, in []byte) (out []byte, err error) {
	if len(in) < aes.BlockSize {
		return nil, errInvalidCipherText
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	iv := in[:aes.BlockSize]
	ct := in[aes.BlockSize:]
	ctr := cipher.NewCTR(c, iv)
	out = make([]byte, len(ct))
	ctr.XORKeyStream(out, ct)
	return
}

// Seal returns an authenticated and encrypted message, and a boolean
// indicating whether the sealing operation was successful. If it returns
// true, the message was successfully sealed. The box will be Overhead
// bytes longer than the message.
func Seal(message []byte, key Key) (box []byte, ok bool) {
	if !KeyIsSuitable(key) {
		return
	}

	ct, err := encrypt(key[:cryptKeySize], message)
	if err != nil {
		return
	}
	tag := computeTag(key[cryptKeySize:], ct)
	box = append(ct, tag...)
	ok = true
	return
}

// Open authenticates and decrypts a sealed message, also returning
// whether the message was successfully opened. If this is false, the
// message must be discarded. The returned message will be Overhead
// bytes shorter than the box.
func Open(box []byte, key Key) (message []byte, ok bool) {
	if !KeyIsSuitable(key) {
		return
	} else if box == nil {
		return
	} else if len(box) <= Overhead {
		return
	}

	msgLen := len(box) - sha512.Size
	if !checkTag(key[cryptKeySize:], box) {
		return nil, ok
	}
	message, err := decrypt(key[:cryptKeySize], box[:msgLen])
	ok = err == nil
	return
}

// KeyIsSuitable returns true if the byte slice represents a valid
// key.
func KeyIsSuitable(key []byte) bool {
	return subtle.ConstantTimeEq(int32(len(key)), int32(KeySize)) == 1
}

// Sign a given string with the app-configured secret key.
// If no secret key is set, returns the empty string.
// Return the signature in base64 (URLEncoding).
func Sign(message string, secretKey Key) string {
	mac := hmac.New(sha1.New, secretKey)
	mac.Write([]byte(message))
	signature := mac.Sum(nil)
	return hex.EncodeToString(signature)
}

// Verify returns true if the given signature is correct for the given message.
// e.g. it matches what we generate with Sign()
func Verify(message, signature string, secretKey Key) bool {
	decSignature, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	decoded, _ := hex.DecodeString(Sign(message, secretKey))

	return hmac.Equal(decSignature, decoded)
}

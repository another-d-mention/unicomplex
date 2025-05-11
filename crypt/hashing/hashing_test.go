package hashing

import (
	"testing"
)

func TestHashing(t *testing.T) {
	data := []byte("Hello, world!")

	if hash := NewSHA1Hasher().Bytes(data).String(); hash != "943a702d06f34599aee1f8da8ef9f7296031d699" {
		t.Error("SHA1 hash incorrect")
	}

	if hash := NewSHA256Hasher().Bytes(data).String(); hash != "315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3" {
		t.Error("SHA256 hash incorrect")
	}

	if hash := NewSHA384Hasher().Bytes(data).String(); hash != "55bc556b0d2fe0fce582ba5fe07baafff035653638c7ac0d5494c2a64c0bea1cc57331c7c12a45cdbca7f4c34a089eeb" {
		t.Error("SHA384 hash incorrect")
	}

	if hash := NewSHA512Hasher().Bytes(data).String(); hash != "c1527cd893c124773d811911970c8fe6e857d6df5dc9226bd8a160614c0cd963a4ddea2b94bb7d36021ef9d865d5cea294a82dd49a0bb269f51f6e7a57f79421" {
		t.Error("SHA512 hash incorrect")
	}

	if hash := NewMD5Hasher().Bytes(data).String(); hash != "6cd3556deb0da54bca060b4c39479839" {
		t.Error("MD5 hash incorrect")
	}

	if hash := NewCRC32Hasher().Bytes(data).String(); hash != "ebe6c6e6" {
		t.Error("CRC32 hash incorrect")
	}

	if hash := NewCRC64Hasher().Bytes(data).String(); hash != "d176c45c2611d9c2" {
		t.Error("CRC64 hash incorrect")
	}

	if hash := NewCRC64ECMAHasher().Bytes(data).String(); hash != "8e59e143665877c4" {
		t.Error("CRC64ECMA hash incorrect")
	}

	if hash := NewAdler32Hasher().Bytes(data).String(); hash != "205e048a" {
		t.Error("Adler32 hash incorrect")
	}

	if hash := NewMurmur128Hasher().Bytes(data).String(); hash != "f1512dd1d2d665df2c326650a8f3c564" {
		t.Error("MurmurHash incorrect", hash)
	}
	if hash := NewMurmur256Hasher().Bytes(data).String(); hash != "f1512dd1d2d665df2c326650a8f3c5642c50f72fb1e979de3e5e04cbb3d65acd" {
		t.Error("MurmurHash incorrect", hash)
	}
}

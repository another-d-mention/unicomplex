package bloomfilter

import (
	"bytes"
	"fmt"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := baseHashes([]byte("Hello, world!"))
	if hash[0] != 17388730015462876639 || hash[1] != 3184720383122326884 ||
		hash[2] != 3193323920026073566 || hash[3] != 4494034751104375501 {
		t.Error("Hashing incorrect")
	}
}

func TestBasic(t *testing.T) {
	f := New(1000, 4)
	n1 := []byte("Jane")
	n2 := []byte("Joe")
	n3 := []byte("Doe")

	f.Add(n1)
	f.Add(n3)

	if f.Test(n1) == false {
		t.Errorf("%s should be in.", n1)
	}
	if f.Test(n2) == true {
		t.Errorf("%s should not be in.", n2)
	}
	if f.ApproximatedSize() != 2 {
		t.Error("Approximated size incorrect")
	}
	d, e := f.MarshalJSON()
	if e != nil {
		t.Error(e)
	}
	fmt.Println(string(d))
	d, e = f.MarshalBinary()
	if e != nil {
		t.Error(e)
	}
	fmt.Println(d)

	f.Reset()
	if f.ApproximatedSize() != 0 {
		t.Error("Approximated size incorrect after reset")
	}
	_, _ = f.ReadFrom(bytes.NewReader(d))
	if f.ApproximatedSize() != 2 {
		t.Error("Approximated size incorrect after deserialization")
	}
}

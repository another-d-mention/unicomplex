package bitset

import (
	"bytes"
	"testing"
)

func TestBitset(t *testing.T) {
	v := New(1)
	for i := uint(0); i < 10; i++ {
		v.Set(i)
	}
	if v.String() != "{0,1,2,3,4,5,6,7,8,9}" {
		t.Error("bad string output")
	}
	if !v.Test(5) {
		t.Error("Test failed")
	}
	v.Clear(5)
	if v.Test(5) {
		t.Error("Test failed")
	}
	v.Flip(5)
	if !v.Test(5) {
		t.Error("Test failed")
	}
	data, err := v.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	if string(data) != `"AAAAAAAAAAoAAAAAAAAD_w=="` {
		t.Error("bad JSON output")
	}
	data, err = v.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(data, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xFF}) {
		t.Error("bad binary output")
	}
}

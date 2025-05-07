package syncmap

import (
	"testing"
)

func TestNewSyncMap(t *testing.T) {
	m := New[string, int]()
	m.Set("key", 1)
	if v, ok := m.Get("key"); !ok || v != 1 {
		t.Error("expected value to be 1")
	} else {
		t.Log("value set successfully")
	}

	m.Delete("key")
	if v, ok := m.Get("key"); ok {
		t.Error("expected value to be deleted")
	} else {
		t.Log("value deleted successfully", v)
	}
}

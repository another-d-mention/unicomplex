package mmap

import (
	"os"
	"testing"
)

func TestMmap(t *testing.T) {
	f, err := os.OpenFile("testdata/file.txt", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()
	_ = f.Truncate(1024)

	data, err := MemMapFile(f, 0, 1024, PROT_READ|PROT_WRITE, MAP_SHARED)

	data[20] = 'H'

	f.Seek(20, 0)
	f.Write([]byte{'H'})

	Unmap(data)
}

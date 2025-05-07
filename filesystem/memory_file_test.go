package filesystem

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
)

func TestMemoryFile_ReadAt(t *testing.T) {
	f := &memoryFile{
		data: &memoryFileData{
			data: []byte("Hello, World!"),
			size: 13,
		},
	}
	p := make([]byte, 5)
	n, err := f.ReadAt(p, 0)
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Error("expected 5 bytes read, got", n)
	}
	if string(p) != "Hello" {
		t.Error("expected 'Hello' but got", string(p))
	}

	p = make([]byte, 100)
	n, err = f.ReadAt(p, 7)
	if err != nil {
		t.Error(err)
	}
	if n != 6 {
		t.Error("expected 6 bytes read, got", n)
	}
	if string(p[:n]) != "World!" {
		t.Error("expected 'World!' but got", string(p[:n]))
	}

	n, err = f.ReadAt(p, 53)
	if err == nil {
		t.Error("expected EOF, got", n)
	}
}

func TestMemoryFile_Read(t *testing.T) {
	f := &memoryFile{
		data: &memoryFileData{
			data: []byte("Hello, World!"),
			size: 13,
		},
	}
	p := make([]byte, 5)
	n, err := f.Read(p)
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Error("expected 5 bytes read, got", n)
	}
	if string(p) != "Hello" {
		t.Error("expected 'Hello' but got", string(p))
	}

	p = make([]byte, 100)
	n, err = f.Read(p)
	if string(p[:n]) != ", World!" {
		t.Error("expected ', World!' but got", string(p[:n]))
	}
	if err != nil {
		t.Error(err)
	}
}

func TestFileSystem(t *testing.T) {
	fs := NewMemoryFilesystem()
	err := fs.CreateDir("/test/asd/xx/yy")
	if err != nil {
		t.Error(err)
	}
	err = fs.CreateDir("test/dsa")
	if err != nil {
		t.Error(err)
	}

	err = fs.CreateDir("test/dsa")
	if err == nil {
		t.Error("expected error when creating existing directory")
	}

	printTree(fs)

	entries, err := fs.ReadDir("/test", false)
	if len(entries) != 2 {
		t.Error("expected 1 directory entry, got", len(entries))
	}

	err = fs.Remove("/missing")
	if err == nil {
		t.Error("expected error when renaming non-existent file")
	}

	err = fs.Remove("/test/asd/xx")
	if err != nil {
		t.Error(err)
	}
	_ = fs.CreateDir("test/dsa/new")

	printTree(fs)

	err = fs.Rename("missing", "asd")
	if err == nil {
		t.Error("expected error when renaming non-existent file")
	}

	err = fs.Rename("test/dsa", "test/cde")
	if err != nil {
		t.Error(err)
	}

	err = fs.Rename("test/cde", "test/cde/new")
	if err == nil {
		t.Error("Should have failed to rename directory into one of it's children")
	}

	printTree(fs)

	stat, err := fs.Stat("/test/cde/new")
	if err != nil || !stat.IsDir() {
		t.Error("expected directory stat, got", err)
	}

	f, _ := fs.Open("test/cde/new/file.txt", os.O_CREATE|os.O_WRONLY, 0644)
	_, _ = f.Write([]byte("Hello, World!"))

	sum, err := fs.FileHash("test/cde/new/file.txt", hashing.NewMD5Hasher())
	if err != nil {
		t.Error(err)
	}

	if sum.HexEncoded() != "65a8e27d8879283831b664bd8b7f0ad4" {
		t.Error("expected MD5 hash, got", []byte(sum), sum.String())
	}

	err = fs.Copy("test/cde/new/file.txt", "test/cde/new/copied.txt")
	if err != nil {
		t.Error(err)
	}
	dest, _ := fs.ReadFile("test/cde/new/copied.txt")
	if !bytes.Equal(dest, []byte("Hello, World!")) {
		t.Error("expected copied file content to match original")
	}

	tmp := []byte("Hello, Unicomplex!")
	err = fs.WriteFile("test.txt", tmp)
	if err != nil {
		t.Error(err)
	}
	dest, _ = fs.ReadFile("test.txt")
	if !bytes.Equal(dest, tmp) {
		t.Error("expected copied file content to match original")
	}
}

func TestSub(t *testing.T) {
	fs := NewMemoryFilesystem()
	err := fs.CreateDir("/test/asd/xx/yy")
	if err != nil {
		t.Error(err)
	}

	nfs, err := fs.Sub("/test/as")
	if err == nil {
		t.Error("expected error when accessing non-existent sub-directory")
	}

	entries, _ := fs.ReadDir("/", false)
	if len(entries) != 1 {
		t.Error("expected 1 directory entry, got", len(entries))
	}

	nfs, err = fs.Sub("/test/asd")
	if err != nil {
		t.Error(err)
	}
	entries, _ = nfs.ReadDir("/", false)
	if len(entries) != 1 {
		t.Error("expected 1 directory entry, got", len(entries))
	}

	err = nfs.WriteFile("test.txt", []byte("hi"))

	printTree(fs)

	entries, _ = nfs.ReadDir("/", false)
	if len(entries) != 2 {
		t.Error("expected 2 directory entry, got", len(entries))
	}

	entries, _ = nfs.ReadDir("/", true)
	if len(entries) != 3 {
		t.Error("expected 3 directory entry, got", len(entries))
	}
}

func printTree(fs FileSystem) {
	files := fs.(*MemoryFilesystem).files.Keys()
	sort.Strings(files)
	for _, file := range files {
		fmt.Println(file)
	}
	fmt.Printf("\n\n")
}

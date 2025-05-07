package filesystem

import (
	"fmt"
	"os"
	"testing"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
)

func TestDiskFileSystem(t *testing.T) {
	fs := NewDiskFilesystem()

	sfs, err := fs.Sub("media")
	if err != nil {
		t.Error(err)
		return
	}

	entries, err := sfs.ReadDir("/", false)
	if err != nil {
		t.Error(err)
	}

	if !fs.Exists("~/.gitignore") {
		t.Error("expected file '~/.gitignore' to exist")
		return
	}

	sum, err := fs.FileHash("~/.gitignore", hashing.NewMD5Hasher())
	if err != nil {
		t.Error(err)
	}
	fmt.Println("MD5 hash of ~/.gitignore:", sum.HexEncoded())
	printDiskTree(entries)

	err = fs.Copy("~/.gitignore", "~/TEST/.gitignore")
	if err != nil {
		t.Error(err)
		return
	}
	err = fs.Rename("~/TEST/.gitignore", "~/TEST/new/.gitignore")
	if err != nil {
		t.Error(err)
		return
	}
	data, err := fs.ReadFile("~/TEST/new/.gitignore")
	if err != nil {
		t.Error(err)
		return
	}
	err = fs.WriteFile("~/TEST/iacob/test.txt", []byte("Hello, Unicomplex!"))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("Content of copied and renamed file:", string(data))
}

func printDiskTree(entries []os.DirEntry) {
	for _, file := range entries {
		info, _ := file.Info()
		fmt.Println(file.IsDir(), file.Name(), info.Size())
	}
	fmt.Printf("\n\n")
}

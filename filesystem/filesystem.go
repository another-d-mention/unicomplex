package filesystem

import (
	"io"
	"os"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
	"github.com/fsnotify/fsnotify"
)

// File is the interface compatible with os.File.
type File interface {
	io.Reader
	io.ReaderAt
	io.Writer
	io.WriterAt
	io.Seeker
	io.Closer

	Slice(start int64, end int64) ([]byte, error)
	Truncate(size int64) error
	Stat() (os.FileInfo, error)
	Sync() error
}

type Event = fsnotify.Event

type FileSystem interface {
	// Open opens the named file.
	Open(name string, flags int, perm os.FileMode) (File, error)
	// ReadDir returns a slice of directory entries in the named directory.
	ReadDir(path string, recursive bool) ([]os.DirEntry, error)
	// ReadFile reads and returns the content of the named file.
	ReadFile(path string) ([]byte, error)
	// WriteFile writes the content to the named file.
	WriteFile(path string, data []byte) error
	// CreateDir creates a new directory with the given name.
	CreateDir(path string) error
	// Remove removes the named file or directory.
	Remove(path string) error
	// Rename renames the named file or directory to newPath.
	Rename(oldPath, newPath string) error
	// Copy copies source file to destination
	Copy(source, destination string) error
	// Stat returns the FileInfo for the named file or directory.
	Stat(path string) (os.FileInfo, error)
	// Exists checks if the named file or directory exists.
	Exists(path string) bool
	// FileHash hashes the content of the named file with the given hash function.
	FileHash(path string, hasher hashing.Hasher) (hashing.Sum, error)
	// Sub returns a new FileSystem rooted at dir
	Sub(dir string) (FileSystem, error)
	// Watch a file or directory for changes.
	Watch(path string, callback chan Event) error
	// Unwatch removes a watch on a file or directory.
	Unwatch(path string, callback chan Event) error
	// RootDir Returns the root directory of the filesystem.
	RootDir() string
}

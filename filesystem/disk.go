package filesystem

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
)

type DiskFilesystem struct {
	rootDir string
	root    *os.Root
}

func NewDiskFilesystem() FileSystem {
	root, _ := os.OpenRoot("/")
	return DiskFilesystem{
		rootDir: "/",
		root:    root,
	}
}

func (d DiskFilesystem) Open(name string, flags int, perm os.FileMode) (File, error) {
	f, e := d.root.OpenFile(name, flags, perm)
	if e != nil {
		return nil, e
	}
	return &diskFile{File: f}, nil
}

func (d DiskFilesystem) ReadDir(path string, recursive bool) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(d.getPath(path))
	if err != nil {
		return nil, err
	}
	if !recursive {
		return entries, nil
	}
	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(path, entry.Name())
			subEntries, _ := d.ReadDir(subPath, recursive)
			for _, subEntry := range subEntries {
				entries = append(entries, subEntry)
			}
		}
	}
	return entries, nil
}

func (d DiskFilesystem) ReadFile(path string) ([]byte, error) {
	f, e := d.root.OpenFile(d.resolve(path), os.O_RDONLY, 0)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (d DiskFilesystem) WriteFile(path string, data []byte) error {
	path = d.resolve(path)
	if !d.Exists(filepath.Dir(path)) {
		if err := d.CreateDir(filepath.Dir(path)); err != nil {
			return err
		}
	}
	f, e := d.root.OpenFile(d.resolve(path), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if e != nil {
		return e
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func (d DiskFilesystem) CreateDir(path string) error {
	path = d.getPath(path)
	return os.MkdirAll(path, 0755)
}

func (d DiskFilesystem) Remove(path string) error {
	path = d.getPath(path)
	return os.RemoveAll(path)
}

func (d DiskFilesystem) Rename(oldPath, newPath string) error {
	newPath = d.getPath(d.resolve(newPath))
	if !d.Exists(filepath.Dir(newPath)) {
		if err := d.CreateDir(filepath.Dir(newPath)); err != nil {
			return err
		}
	}
	if !d.Exists(oldPath) {
		return os.ErrNotExist
	}
	return os.Rename(d.getPath(d.resolve(oldPath)), newPath)
}

func (d DiskFilesystem) Copy(source, destination string) error {
	destination = d.resolve(destination)
	stats, err := d.Stat(source)
	if err != nil {
		return err
	}

	source = d.resolve(source)
	src, err := d.Open(source, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer src.Close()

	if !d.Exists(filepath.Dir(destination)) {
		if err := d.CreateDir(filepath.Dir(destination)); err != nil {
			return err
		}
	}

	dst, err := d.Open(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, stats.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func (d DiskFilesystem) Stat(path string) (os.FileInfo, error) {
	path = d.resolve(path)
	return d.root.Stat(path)
}

func (d DiskFilesystem) Exists(path string) bool {
	if _, err := d.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (d DiskFilesystem) FileHash(path string, hasher hashing.Hasher) (hashing.Sum, error) {
	f, er := d.Open(d.resolve(path), os.O_RDONLY, 0)
	if er != nil {
		return nil, er
	}
	defer f.Close()
	return hasher.Reader(f), nil
}

func (d DiskFilesystem) Sub(dir string) (FileSystem, error) {
	dir = filepath.Clean(dir)
	root, err := d.root.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	return DiskFilesystem{
		rootDir: d.getPath(dir),
		root:    root,
	}, nil
}

func (d DiskFilesystem) resolve(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return strings.TrimPrefix(filepath.Join(home, path[1:]), d.rootDir)
	}
	if filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return path
		}
		return strings.TrimPrefix(abs, d.rootDir)
	}
	return path
}

func (d DiskFilesystem) getPath(path string) string {
	return filepath.Clean(filepath.Join(d.rootDir, filepath.Clean(path)))
}

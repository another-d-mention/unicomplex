package filesystem

import (
	"io"
	"os"
	"path/filepath"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
)

type DiskFilesystem struct {
	*base
}

func NewDiskFilesystem() FileSystem {
	return &DiskFilesystem{
		base: &base{
			userRoot: "/",
			rootDir:  "/",
		},
	}
}

func (d *DiskFilesystem) Open(name string, flags int, perm os.FileMode) (File, error) {
	if !d.Exists(filepath.Dir(name)) {
		if flags&os.O_CREATE != 0 {
			if err := d.CreateDir(filepath.Dir(name)); err != nil {
				return nil, err
			}
		}
	}
	f, e := os.OpenFile(absolutePath(d.rootDir, name), flags, perm)
	if e != nil {
		return nil, e
	}
	return &diskFile{File: f}, nil
}

func (d *DiskFilesystem) ReadDir(path string, recursive bool) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(absolutePath(d.rootDir, path))
	if err != nil {
		return nil, err
	}

	if !recursive {
		return entries, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subEntries, _ := d.ReadDir(filepath.Join(path, entry.Name()), recursive)
			entries = append(entries, subEntries...)
		}
	}

	return entries, nil
}

func (d *DiskFilesystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(absolutePath(d.rootDir, path))
}

func (d *DiskFilesystem) WriteFile(path string, data []byte) error {
	if !d.Exists(filepath.Dir(path)) {
		if err := d.CreateDir(filepath.Dir(path)); err != nil {
			return err
		}
	}
	return os.WriteFile(absolutePath(d.rootDir, path), data, 0666)
}

func (d *DiskFilesystem) CreateDir(path string) error {
	return os.MkdirAll(absolutePath(d.rootDir, path), 0755)
}

func (d *DiskFilesystem) Remove(path string) error {
	return os.RemoveAll(absolutePath(d.rootDir, path))
}

func (d *DiskFilesystem) Rename(oldPath, newPath string) error {
	if !d.Exists(filepath.Dir(newPath)) {
		if err := d.CreateDir(filepath.Dir(newPath)); err != nil {
			return err
		}
	}

	return os.Rename(absolutePath(d.rootDir, oldPath), absolutePath(d.rootDir, newPath))
}

func (d *DiskFilesystem) Copy(source, destination string) error {
	src, err := d.Open(source, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer src.Close()

	stats, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := d.Open(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, stats.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (d *DiskFilesystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(absolutePath(d.rootDir, path))
}

func (d *DiskFilesystem) Exists(path string) bool {
	if _, err := d.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (d *DiskFilesystem) FileHash(path string, hasher hashing.Hasher) (hashing.Sum, error) {
	f, er := d.Open(path, os.O_RDONLY, 0)
	if er != nil {
		return nil, er
	}
	defer f.Close()
	return hasher.Reader(f), nil
}

func (d *DiskFilesystem) Sub(dir string) (FileSystem, error) {
	return &DiskFilesystem{
		base: &base{
			userRoot: dir,
			rootDir:  absolutePath(d.rootDir, dir),
		},
	}, nil
}

func (d *DiskFilesystem) resolve(name string) string {
	return absolutePath(d.userRoot, name)
}

package filesystem

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
	"github.com/another-d-mention/unicomplex/datastruct/syncmap"
)

type MemoryFilesystem struct {
	rootDir string
	files   *syncmap.Map[string, *memoryFileData]
}

func NewMemoryFilesystem() FileSystem {
	return &MemoryFilesystem{
		rootDir: "/",
		files:   syncmap.New[string, *memoryFileData](),
	}
}

func (m *MemoryFilesystem) Open(name string, flags int, perm os.FileMode) (File, error) {
	oname := name
	name = m.getPath(name)
	f, ok := m.files.Get(name)
	if !ok {
		if (flags & os.O_CREATE) == 0 { // no create flag
			return nil, os.ErrNotExist
		}
		f = &memoryFileData{
			name: name,
			perm: perm,
			data: make([]byte, 0),
			mod:  time.Now(),
			size: 0,
		}
		_ = m.CreateDir(filepath.Dir(oname))
		m.files.Set(name, f)
	}

	if (flags & os.O_TRUNC) != 0 { // truncate flag
		f.size = 0
		f.data = make([]byte, 0)
	}

	return &memoryFile{data: f}, nil
}

func (m *MemoryFilesystem) ReadDir(dir string, recursive bool) ([]os.DirEntry, error) {
	dir = m.getPath(dir)
	var entries []os.DirEntry
	m.files.Range(func(key string, value *memoryFileData) bool {
		if key == dir {
			return true
		}
		if recursive && strings.HasPrefix(key, dir) {
			entries = append(entries, value)
		} else if !recursive && filepath.Dir(key) == dir {
			entries = append(entries, value)
		}
		return true
	})
	return entries, nil
}

func (m *MemoryFilesystem) ReadFile(path string) ([]byte, error) {
	path = m.getPath(path)
	if !m.Exists(path) {
		return nil, os.ErrNotExist
	}
	f, _ := m.files.Get(path)
	data := make([]byte, f.size)
	copy(data, f.data)
	return data, nil
}

func (m *MemoryFilesystem) WriteFile(path string, data []byte) error {
	f, err := m.Open(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(data))
	return err
}

func (m *MemoryFilesystem) CreateDir(dir string) error {
	dir = filepath.Join(m.rootDir, filepath.Clean(dir))
	_, ok := m.files.Get(dir)
	if ok {
		return os.ErrExist
	}
	m.createDirAll(dir)
	return nil
}

func (m *MemoryFilesystem) Remove(path string) error {
	path = m.getPath(path)
	if !m.Exists(path) {
		return os.ErrNotExist
	}
	f, _ := m.files.Get(path)
	if f.IsDir() {
		m.files.Range(func(key string, value *memoryFileData) bool {
			if strings.HasPrefix(filepath.Dir(key), path) {
				m.files.Delete(key)
			}
			return true
		})
	}
	m.files.Delete(path)
	return nil
}

func (m *MemoryFilesystem) Rename(oldPath, newPath string) error {
	oldPath = m.getPath(oldPath)
	if !m.Exists(oldPath) {
		return os.ErrNotExist
	}
	newPath = m.getPath(newPath)

	if oldPath == newPath { // no change
		return nil
	}

	f, _ := m.files.Get(oldPath)
	if f.IsDir() {
		if strings.HasPrefix(newPath, oldPath) { // trying to move a parent into it's children
			return errors.New("cannot rename directory to itself")
		}

		m.files.Range(func(key string, value *memoryFileData) bool {
			// rename all prefixes for all children
			if strings.HasPrefix(filepath.Dir(key), oldPath) || key == oldPath {
				m.files.Delete(key)
				m.files.Set(strings.Replace(key, oldPath, newPath, 1), value)
			} else {
				fmt.Println("Skipping:", filepath.Dir(key), oldPath)
			}
			return true
		})
	} else {
		// it's a file
		m.files.Delete(oldPath)
		m.files.Set(newPath, f)
	}
	return nil
}

func (m *MemoryFilesystem) Copy(source, destination string) error {
	destination = m.getPath(destination)
	source = m.getPath(source)
	if !m.Exists(source) {
		return os.ErrNotExist
	}

	sourceF, _ := m.files.Get(source)
	if sourceF.IsDir() {
		// copy dir and all its contents
		return nil
	}

	src, _ := m.Open(source, os.O_RDONLY, 0)
	stat, _ := src.Stat()
	dst, err := m.Open(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, stat.Mode())

	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}

func (m *MemoryFilesystem) Stat(path string) (os.FileInfo, error) {
	path = m.getPath(path)
	f, ok := m.files.Get(path)
	if !ok {
		return nil, os.ErrNotExist
	}
	return f.Stats()
}

func (m *MemoryFilesystem) Exists(path string) bool {
	return m.files.Contains(m.getPath(path))
}

func (m *MemoryFilesystem) FileHash(path string, hasher hashing.Hasher) (hashing.Sum, error) {
	f, ok := m.files.Get(m.getPath(path))
	if !ok {
		return nil, os.ErrNotExist
	}
	if f.IsDir() {
		return nil, errors.New("cannot hash a directory")
	}
	return hasher.Bytes(f.data), nil
}

func (m *MemoryFilesystem) Sub(dir string) (FileSystem, error) {
	dir = m.getPath(dir)
	if !m.Exists(dir) {
		return nil, os.ErrNotExist
	}

	return &MemoryFilesystem{
		rootDir: dir,
		files:   m.files,
	}, nil
}

func (m *MemoryFilesystem) getPath(path string) string {
	return filepath.Clean(filepath.Join(m.rootDir, filepath.Clean(path)))
}

func (m *MemoryFilesystem) createDirAll(path string) {
	path = filepath.Clean(path)
	parents := strings.Split(path, "/")
	for i := 1; i < len(parents); i++ {
		dir := m.getPath(strings.Join(parents[:i+1], "/"))
		_, ok := m.files.Get(dir)
		if !ok {
			m.files.Set(dir, &memoryFileData{
				name: dir,
				perm: os.ModePerm | os.ModeDir,
				mod:  time.Now(),
			})
		}
	}
}

package filesystem

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
	"github.com/another-d-mention/unicomplex/datastruct/syncmap"
)

type MemoryFilesystem struct {
	*base
	files *syncmap.Map[string, *memoryFileData]
}

func NewMemoryFilesystem() FileSystem {
	return &MemoryFilesystem{
		base: &base{
			userRoot: "/",
			rootDir:  "/",
		},
		files: syncmap.New[string, *memoryFileData](),
	}
}

func (m *MemoryFilesystem) Open(name string, flags int, perm os.FileMode) (File, error) {
	originalName := name
	name = absolutePath(m.rootDir, name)

	f, ok := m.files.Get(name)
	if !ok {
		if (flags & os.O_CREATE) == 0 { // no create flag
			return nil, os.ErrNotExist
		}
		_ = m.CreateDir(filepath.Dir(originalName))

		f = &memoryFileData{
			name: name,
			perm: perm,
			data: make([]byte, 0),
			mod:  time.Now(),
			size: 0,
		}
		m.files.Set(name, f)
	}

	if (flags & os.O_TRUNC) != 0 { // truncate flag
		f.size = 0
		f.data = make([]byte, 0)
	}

	return &memoryFile{data: f}, nil
}

func (m *MemoryFilesystem) ReadDir(dir string, recursive bool) ([]os.DirEntry, error) {
	dir = absolutePath(m.rootDir, dir)

	var entries []os.DirEntry
	m.files.Range(func(key string, value *memoryFileData) bool {
		if key == dir { // should not return the folder itself
			return true
		}

		if recursive && strings.HasPrefix(key, dir) { // dir must be a prefix of key
			entries = append(entries, value)
		} else if !recursive && filepath.Dir(key) == dir { // dir must be exactly the dir
			entries = append(entries, value)
		}

		return true
	})
	return entries, nil
}

func (m *MemoryFilesystem) ReadFile(path string) ([]byte, error) {
	if !m.Exists(path) {
		return nil, os.ErrNotExist
	}

	f, _ := m.files.Get(absolutePath(m.rootDir, path))
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
	exists := m.files.Contains(absolutePath(m.rootDir, dir))
	if exists {
		return os.ErrExist
	}

	// yeah, we remove add then remove the rootDir, but the absolutePath also resolves relative paths
	// so we use it for that purpose at the moment
	path := strings.TrimPrefix(m.rootDir, absolutePath(m.rootDir, dir))

	parents := strings.Split(path, "/")

	for i := 1; i < len(parents); i++ {
		target := absolutePath(m.rootDir, strings.Join(parents[:i+1], "/"))
		if _, ok := m.files.Get(target); !ok {
			m.files.Set(target, &memoryFileData{
				name: target,
				perm: os.ModePerm | os.ModeDir,
				mod:  time.Now(),
			})
		}
	}

	return nil
}

func (m *MemoryFilesystem) Remove(path string) error {
	if !m.Exists(path) {
		return os.ErrNotExist
	}

	path = absolutePath(m.rootDir, path)
	f, _ := m.files.Get(path)

	if f.IsDir() {
		m.files.Range(func(key string, value *memoryFileData) bool {
			if strings.HasPrefix(filepath.Dir(key), path) { // a child
				m.files.Delete(key)
			}
			return true
		})
	}

	m.files.Delete(path)
	return nil
}

func (m *MemoryFilesystem) Rename(oldPath, newPath string) error {
	if !m.Exists(oldPath) {
		return os.ErrNotExist
	}

	oldPath = absolutePath(m.rootDir, oldPath)
	newPath = absolutePath(m.rootDir, newPath)

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
				value.name = strings.Replace(key, oldPath, newPath, 1)
				m.files.Set(value.name, value)
			}

			return true
		})
	} else {
		f.name = newPath
		m.files.Delete(oldPath)
		m.files.Set(newPath, f)
	}
	return nil
}

func (m *MemoryFilesystem) Copy(source, destination string) error {
	if !m.Exists(source) {
		return os.ErrNotExist
	}

	stat, _ := m.Stat(absolutePath(m.rootDir, source))
	if stat.IsDir() {

		absSource := absolutePath(m.rootDir, source)

		filesToCopy, _ := m.ReadDir(source, true)
		for _, file := range filesToCopy {
			mSrc := file.(*memoryFileData)
			sourcePath := strings.TrimPrefix(mSrc.name, absSource)
			src, _ := m.Open(mSrc.name, os.O_RDONLY, 0)

			dst, err := m.Open(filepath.Join(destination, sourcePath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, stat.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(dst, src)
		}

		return nil
	}

	src, _ := m.Open(source, os.O_RDONLY, 0)
	dst, err := m.Open(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, stat.Mode())
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	return err
}

func (m *MemoryFilesystem) Stat(path string) (os.FileInfo, error) {
	f, ok := m.files.Get(absolutePath(m.rootDir, path))
	if !ok {
		return nil, os.ErrNotExist
	}
	return f.Stats()
}

func (m *MemoryFilesystem) Exists(path string) bool {
	return m.files.Contains(absolutePath(m.rootDir, path))
}

func (m *MemoryFilesystem) FileHash(path string, hasher hashing.Hasher) (hashing.Sum, error) {
	f, ok := m.files.Get(absolutePath(m.rootDir, path))
	if !ok {
		return nil, os.ErrNotExist
	}
	if f.IsDir() {
		return nil, errors.New("cannot hash a directory")
	}
	return hasher.Bytes(f.data), nil
}

func (m *MemoryFilesystem) Sub(dir string) (FileSystem, error) {
	if !m.Exists(dir) {
		return nil, os.ErrNotExist
	}

	return &MemoryFilesystem{
		base: &base{
			userRoot: dir,
			rootDir:  absolutePath(m.rootDir, dir),
		},
		files: m.files,
	}, nil
}

func (m *MemoryFilesystem) Watch(path string, callback chan Event) error {
	return nil
}

func (m *MemoryFilesystem) Unwatch(path string, callback chan Event) error {
	return nil
}

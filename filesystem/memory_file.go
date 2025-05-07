package filesystem

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type memoryFileData struct {
	name string
	perm os.FileMode
	data []byte
	mod  time.Time
	size int64
	lock sync.RWMutex
}

func (f *memoryFileData) Name() string {
	_, name := filepath.Split(f.name)
	return name
}

func (f *memoryFileData) Size() int64 {
	return f.size
}

func (f *memoryFileData) Mode() os.FileMode {
	return f.perm
}

func (f *memoryFileData) Type() os.FileMode {
	return f.perm
}

func (f *memoryFileData) Stats() (os.FileInfo, error) {
	return f, nil
}

func (f *memoryFileData) Info() (os.FileInfo, error) {
	return f, nil
}

func (f *memoryFileData) ModTime() time.Time {
	return f.mod
}

func (f *memoryFileData) IsDir() bool {
	return f.perm.IsDir()
}

func (f *memoryFileData) Sys() interface{} {
	return nil
}

type memoryFile struct {
	data   *memoryFileData
	offset int64
}

func (f *memoryFile) Read(p []byte) (int, error) {
	n, err := f.ReadAt(p, f.offset)
	if err != nil {
		return n, err
	}
	f.offset += int64(n)
	return n, nil
}

func (f *memoryFile) ReadAt(p []byte, offset int64) (int, error) {
	if f.data == nil {
		return 0, os.ErrClosed
	}
	if offset >= f.data.size {
		return 0, io.EOF
	}

	f.data.lock.RLock()
	defer f.data.lock.RUnlock()

	size := int64(len(p))
	// buffer length is bigger than the available data
	if offset+size > f.data.size {
		size = f.data.size - offset
		copy(p, f.data.data[offset:offset+size])
		return int(size), nil
	}
	copy(p, f.data.data[offset:])
	return int(size), nil
}

func (f *memoryFile) Write(p []byte) (int, error) {
	n, err := f.WriteAt(p, f.offset)
	if err != nil {
		return n, err
	}
	f.offset += int64(n)
	return n, nil
}

func (f *memoryFile) WriteAt(p []byte, offset int64) (int, error) {
	if f.data == nil {
		return 0, os.ErrClosed
	}

	n := int64(len(p))
	if offset+n > f.data.size {
		if err := f.Truncate(offset + n); err != nil {
			return 0, err
		}
	}

	f.data.lock.Lock()
	defer f.data.lock.Unlock()

	copy(f.data.data[offset:offset+n], p)
	return int(n), nil
}

func (f *memoryFile) Seek(offset int64, whence int) (int64, error) {
	if f.data == nil {
		return 0, os.ErrClosed
	}
	switch whence {
	case io.SeekStart:
		f.offset = offset
	case io.SeekCurrent:
		f.offset += offset
	case io.SeekEnd:
		f.offset = f.data.size + offset
	default:
		return 0, errors.New("invalid whence")
	}
	return f.offset, nil
}

func (f *memoryFile) Close() error {
	if f.data == nil {
		return os.ErrClosed
	}
	return nil
}

func (f *memoryFile) Slice(start, end int64) ([]byte, error) {
	if f.data == nil {
		return nil, os.ErrClosed
	}
	if start < 0 || start > f.data.size || end < start || end > f.data.size {
		return nil, errors.New("slice range out of bounds")
	}

	f.data.lock.RLock()
	defer f.data.lock.RUnlock()

	return f.data.data[start:end], nil
}

func (f *memoryFile) Truncate(size int64) error {
	if f.data == nil {
		return os.ErrClosed
	}
	if size < 0 {
		return errors.New("truncate size must not be negative")
	}

	f.data.lock.Lock()
	defer f.data.lock.Unlock()

	if size > f.data.size {
		diff := int(size - f.data.size)
		f.data.data = append(f.data.data, make([]byte, diff)...) // grow the data
	} else {
		f.data.data = f.data.data[:size] // shrink the data
	}

	f.data.size = size
	return nil
}

func (f *memoryFile) Stat() (os.FileInfo, error) {
	if f.data == nil {
		return nil, os.ErrClosed
	}
	return f.data, nil
}

func (f *memoryFile) Sync() error {
	if f.data == nil {
		return os.ErrClosed
	}
	return nil
}

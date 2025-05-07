package filesystem

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/another-d-mention/unicomplex/os/mmap"
)

type mmapFile struct {
	file      *os.File
	data      []byte
	openFlags int
	offset    int64
	size      int64
	lock      sync.RWMutex
	options   MMapOptions
}

func (m *mmapFile) Read(p []byte) (int, error) {
	n, err := m.ReadAt(p, m.offset)
	if err != nil {
		return n, err
	}
	m.offset += int64(n)
	return n, nil
}

func (m *mmapFile) ReadAt(p []byte, offset int64) (n int, err error) {
	if m.data == nil {
		return 0, os.ErrClosed
	}
	if offset >= m.size {
		return 0, io.EOF
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	size := int64(len(p))
	// buffer length is bigger than the available data
	if offset+size > m.size {
		size = m.size - offset
		copy(p, m.data[offset:offset+size])
		return int(size), nil
	}
	copy(p, m.data[offset:])
	return int(size), nil
}

func (m *mmapFile) Write(p []byte) (int, error) {
	n, err := m.WriteAt(p, m.offset)
	if err != nil {
		return n, err
	}
	m.offset += int64(n)
	return n, nil
}

func (m *mmapFile) WriteAt(p []byte, offset int64) (int, error) {
	if m.file == nil || m.data == nil {
		return 0, os.ErrClosed
	}

	n := int64(len(p))
	if offset+n > m.size {
		if err := m.Truncate(offset + n); err != nil {
			return 0, err
		}
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	copy(m.data[offset:offset+n], p)

	if m.options.Sync == SyncFull {
		if err := m.Sync(); err != nil {
			return 0, err
		}
	}

	return int(n), nil
}

func (m *mmapFile) Seek(offset int64, whence int) (int64, error) {
	if m.file == nil || m.data == nil {
		return 0, os.ErrClosed
	}
	switch whence {
	case io.SeekStart:
		m.offset = offset
	case io.SeekCurrent:
		m.offset += offset
	case io.SeekEnd:
		m.offset = m.size + offset
	default:
		return 0, errors.New("invalid whence")
	}
	return m.offset, nil
}

func (m *mmapFile) Close() (err error) {
	if m.file == nil || m.data == nil {
		return os.ErrClosed
	}
	if err = mmap.Unmap(m.data); err != nil {
		return
	}
	m.data = nil

	if err = m.file.Close(); err != nil {
		return
	}
	m.file = nil
	return nil
}

func (m *mmapFile) Slice(start int64, end int64) ([]byte, error) {
	if m.file == nil || m.data == nil {
		return nil, os.ErrClosed
	}
	if start < 0 || start > m.size || end < start || end > m.size {
		return nil, errors.New("slice range out of bounds")
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.data[start:end], nil
}

func (m *mmapFile) Truncate(size int64) (err error) {
	if m.file == nil || m.data == nil {
		return os.ErrClosed
	}
	if err = m.file.Truncate(size); err != nil {
		return err
	}
	m.size = size
	return m.reMapFile()
}

func (m *mmapFile) Stat() (os.FileInfo, error) {
	if m.file == nil || m.data == nil {
		return nil, os.ErrClosed
	}
	if err := m.Sync(); err != nil {
		return nil, err
	}
	return m.file.Stat()
}

func (m *mmapFile) Sync() error {
	if m.file == nil || m.data == nil {
		return os.ErrClosed
	}

	if !m.isWritable() {
		return nil
	}

	if m.options.Mode == ReadOnlyMode {
		// Even if your mapping is PROT_READ, the underlying page cache can still be dirty
		// (e.g., modified by another process or by the OS during lazy writes).
		// Calling msync() ensures those pages get flushed to disk.
		if m.options.Access == SharedAccess {
			return mmap.SyncInvalidate(m.data)
		}
		// Private mapping; no backing file writes
		return nil
	}

	if m.options.Access == SharedAccess && m.options.Sync == SyncFull {
		return mmap.SyncInvalidate(m.data)
	}

	return mmap.Sync(m.data)
}

func (m *mmapFile) isWritable() bool {
	return m.openFlags&os.O_RDWR != 0 || m.openFlags&os.O_WRONLY != 0
}

func (m *mmapFile) mapFile() (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data, err = mmap.MemMapFile(m.file, 0, int(m.size), mmap.Prot(m.options.Mode), mmap.MapType(m.options.Access))
	return
}

func (m *mmapFile) reMapFile() (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.data != nil {
		err = mmap.Unmap(m.data)
		m.data = nil
	}
	m.data, err = mmap.MemMapFile(m.file, 0, int(m.size), mmap.Prot(m.options.Mode), mmap.MapType(m.options.Access))
	return
}

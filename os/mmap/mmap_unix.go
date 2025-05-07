//go:build !windows

package mmap

import (
	"os"

	"golang.org/x/sys/unix"
)

// MemMapFile creates a memory-mapped region of the given file.
// It returns the memory-mapped region as a byte slice.
// offset - is the offset in bytes from the beginning of the file where the mapping starts.
// length - is the number of bytes to map.
// access - is the protection flags (PROT_READ, PROT_WRITE, PROT_EXEC).
// mapType - is the mapping type (MAP_SHARED, MAP_PRIVATE).
func MemMapFile(f *os.File, offset int64, length int, access Prot, mapType MapType) ([]byte, error) {
	return unix.Mmap(int(f.Fd()), 0, length, int(access), int(mapType))
}

// Sync ensures that all changes made to the memory-mapped region are written
// back to the file and visible to other processes.
// Use when you want strong durability/data consistency.
func Sync(data []byte) error {
	return unix.Msync(data, unix.MS_SYNC)
}

// ASync is a non-blocking version of Sync.
// Use when you want performance-sensitive flushing (non-blocking) in detriment to data consistency.
func ASync(data []byte) error {
	return unix.Msync(data, unix.MS_ASYNC)
}

// SyncInvalidate syncs the changes made to the memory-mapped region, and invalidates the cache copies of other processes
// forcing them to re-read the data from the disk.
// Use in multi-mapping situations to force cache reloads in detriment of performance.
func SyncInvalidate(data []byte) error {
	return unix.Msync(data, unix.MS_SYNC|unix.MS_INVALIDATE)
}

// Unmap calls SyncInvalidate and releases the memory-mapped region.
func Unmap(data []byte) error {
	if err := SyncInvalidate(data); err != nil {
		return err
	}
	return unix.Munmap(data)
}

package filesystem

import (
	"os"

	"github.com/another-d-mention/unicomplex/os/mmap"
)

type MMapFilesystem struct {
	FileSystem
	rootDir string
	options *MMapOptions
}

type MMapAccess mmap.MapType

const (
	SharedAccess  = MMapAccess(mmap.MAP_SHARED)
	PrivateAccess = MMapAccess(mmap.MAP_PRIVATE)
)

type SyncType int

const (
	// SyncNormal will sync all changes when the file is closed or a call to Sync() is made or
	// when the OS decides to flush the data to disk.
	SyncNormal SyncType = iota
	// SyncFull will call Sync() whenever a call to Write or WriteAt is made
	SyncFull
)

type MMapMode mmap.Prot

const (
	ReadOnlyMode  = MMapMode(mmap.PROT_READ)
	WriteOnlyMode = MMapMode(mmap.PROT_WRITE)
	ReadWriteMode = MMapMode(mmap.PROT_READ | mmap.PROT_WRITE)
)

type MMapOptions struct {
	Access MMapAccess
	Mode   MMapMode
	Sync   SyncType
}

var defaultMMapOptions = MMapOptions{
	Access: SharedAccess,
	Mode:   ReadWriteMode,
	Sync:   SyncNormal,
}

func NewMMapFilesystem(options *MMapOptions) FileSystem {
	if options == nil {
		options = &defaultMMapOptions
	}
	return &MMapFilesystem{
		FileSystem: NewDiskFilesystem(),
		rootDir:    "/",
		options:    options,
	}
}

func (m *MMapFilesystem) Open(name string, flags int, perm os.FileMode) (File, error) {
	f, err := m.FileSystem.Open(name, flags, perm)
	if err != nil {
		return nil, err
	}
	return &mmapFile{
		file:      f.(*diskFile).File,
		offset:    0,
		data:      nil,
		size:      0,
		options:   *m.options,
		openFlags: flags,
	}, nil
}

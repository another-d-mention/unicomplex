package filesystem

import (
	"os"
)

type diskFile struct {
	*os.File
}

func (d *diskFile) Slice(start, end int64) ([]byte, error) {
	if d.File == nil {
		return nil, os.ErrClosed
	}
	buf := make([]byte, end-start)
	_, err := d.ReadAt(buf, start)
	return buf, err
}

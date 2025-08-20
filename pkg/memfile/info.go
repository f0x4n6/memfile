package memfile

import (
	"io/fs"
	"time"
)

// Info provides a fs.FileInfo compatible structure.
type Info struct {
	data *Data // file data
}

// Name returns the file name.
func (i *Info) Name() string {
	return i.data.Name()
}

// Size returns the file size in bytes.
func (i *Info) Size() int64 {
	i.data.RLock()
	defer i.data.RUnlock()
	return int64(len(i.data.buf))
}

// Mode always returns 0 (regular).
func (i *Info) Mode() fs.FileMode {
	return 0 // regular
}

// ModTime returns the files last modified time.
func (i *Info) ModTime() time.Time {
	i.data.RLock()
	defer i.data.RUnlock()
	return i.data.mod
}

// IsDir always returns false.
func (i *Info) IsDir() bool {
	return false
}

// Sys always returns false.
func (i *Info) Sys() any {
	return nil
}

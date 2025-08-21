// Package memfile is an in-memory, thread-safe, dependency-free file abstraction.
package memfile

import (
	"errors"
	"io"
	"io/fs"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrorFileClosed indicates an already closed file
	ErrorFileClosed = errors.New("file closed")
	// ErrorInvalidOffset indicates an invalid offset
	ErrorInvalidOffset = errors.New("invalid offset")
)

// Fileable supports a wide range of io interfaces.
type Fileable interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.ReaderFrom
	io.ReadCloser
	io.ReadSeeker
	io.ReadSeekCloser
	io.ReadWriteCloser
	io.ReadWriteSeeker
	io.ReadWriter
	io.Seeker
	io.StringWriter
	io.Writer
	io.WriterAt
	io.WriterTo
	io.WriteCloser
	io.WriteSeeker

	Name() string
	Stat() (fs.FileInfo, error)
	Truncate(size int64) error
}

// Notify callback for file changes
type Notify func(name string)

// File provides an in-memory file structure.
type File struct {
	sync.RWMutex
	// Given filename.
	name string
	// If file is open.
	open bool
	// File buffer.
	buf []byte
	// File offset.
	off atomic.Int64
	// File last modified time.
	mod time.Time
	// File notification callback.
	fn Notify
}

// FileInfo provides a fs.FileInfo compatible structure.
type FileInfo struct {
	f *File // file data
}

// New returns a new os.File like structure.
func New(name string) Fileable {
	return &File{name: name, open: true}
}

// Name returns the file name.
func (f *File) Name() string {
	return f.name
}

// Close the file for read and write.
func (f *File) Close() error {
	f.Lock()
	defer f.Unlock()
	f.open = false
	return nil
}

// Read the current file offset.
func (f *File) Read(b []byte) (n int, err error) {
	n, err = f.ReadAt(b, f.off.Load())
	f.off.Add(int64(n))
	return
}

// ReadAt the given file offset.
func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	f.RLock()
	defer f.RUnlock()

	if !f.open {
		return 0, ErrorFileClosed
	}

	if off < 0 || int64(int(off)) < off {
		return 0, ErrorInvalidOffset
	}

	if off > int64(len(f.buf)) {
		return 0, io.EOF
	}

	n = copy(b, f.buf[off:])

	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// ReadFrom the given io.Reader.
func (f *File) ReadFrom(r io.Reader) (n int64, err error) {
	b, err := io.ReadAll(r)

	if err != nil {
		return 0, err
	}

	i, err := f.Write(b)

	return int64(i), err
}

// Seek sets the current file offset.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.RLock()
	defer f.RUnlock()

	if !f.open {
		return 0, ErrorFileClosed
	}

	var abs int64

	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.off.Load() + offset
	case io.SeekEnd:
		abs = int64(len(f.buf)) + offset
	default:
		return 0, ErrorInvalidOffset
	}

	if abs < 0 {
		return 0, ErrorInvalidOffset
	}

	f.off.Store(abs)

	return abs, nil
}

// Stat returns fs.FileInfo like stats.
func (f *File) Stat() (fs.FileInfo, error) {
	return &FileInfo{f}, nil
}

// Truncate the file to the given size.
func (f *File) Truncate(size int64) error {
	f.Lock()
	defer f.Unlock()

	if !f.open {
		return ErrorFileClosed
	}

	switch {
	case size < 0 || int64(int(size)) < size:
		return ErrorInvalidOffset
	case size <= int64(len(f.buf)):
		f.buf = f.buf[:size]
	default:
		f.buf = append(f.buf, make([]byte, int(size)-len(f.buf))...)
	}

	f.mod = time.Now()

	if f.fn != nil {
		f.fn(f.name)
	}

	return nil
}

// Write the given bytes at the current file offset.
func (f *File) Write(b []byte) (n int, err error) {
	n, err = f.WriteAt(b, f.off.Load())
	f.off.Add(int64(n))
	return
}

// WriteAt the given bytes and the given file offset.
func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	f.Lock()
	defer f.Unlock()

	if !f.open {
		return 0, ErrorFileClosed
	}

	if off < 0 || int64(int(off)) < off {
		return 0, ErrorInvalidOffset
	}

	if off > int64(len(f.buf)) {
		_ = f.Truncate(off)
	}

	n = copy(f.buf[off:], b)

	f.buf = append(f.buf, b[n:]...)
	f.mod = time.Now()

	if f.fn != nil {
		f.fn(f.name)
	}

	return len(b), nil
}

// WriteTo the given io.Writer.
func (f *File) WriteTo(w io.Writer) (n int64, err error) {
	f.RLock()
	defer f.RUnlock()
	i, err := w.Write(f.buf)
	return int64(i), err
}

// WriteString at the current file offset.
func (f *File) WriteString(s string) (n int, err error) {
	return f.Write([]byte(s))
}

// SetNotify sets the file notification callback.
func (f *File) SetNotify(fn Notify) {
	f.Lock()
	f.fn = fn
	f.Unlock()
}

// Name returns the file name.
func (fi *FileInfo) Name() string {
	return fi.f.Name()
}

// Size returns the file size in bytes.
func (fi *FileInfo) Size() int64 {
	fi.f.RLock()
	defer fi.f.RUnlock()
	return int64(len(fi.f.buf))
}

// Mode always returns 0 (regular).
func (fi *FileInfo) Mode() fs.FileMode {
	return 0 // regular
}

// ModTime returns the files last modified time.
func (fi *FileInfo) ModTime() time.Time {
	fi.f.RLock()
	defer fi.f.RUnlock()
	return fi.f.mod
}

// IsDir always returns false.
func (fi *FileInfo) IsDir() bool {
	return false
}

// Sys always returns false.
func (fi *FileInfo) Sys() any {
	return nil
}

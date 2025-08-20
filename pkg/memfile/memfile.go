// Memfile is an in-memory thread-safe file abstraction.

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
	// ErrorInvalidOffset indicates an invalid offset
	ErrorInvalidOffset = errors.New("invalid offset")
)

// File supports a wide range of io interfaces.
type File interface {
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

// FileData provides an in-memory file data structure.
type FileData struct {
	sync.RWMutex
	// Given filename.
	name string
	// File buffer.
	buf []byte
	// File offset.
	off atomic.Int64
	// File last modified time.
	mod time.Time
	// File notification channel.
	ch chan string
}

// FileInfo provides a fs.FileInfo compatible structure.
type FileInfo struct {
	fd *FileData // file data
}

// New returns a new file like structure.
func New(name string) File {
	return &FileData{name: name}
}

// Name returns the file name.
func (fd *FileData) Name() string {
	return fd.name
}

// Bytes returns the file data.
func (fd *FileData) Bytes() []byte {
	fd.RLock()
	defer fd.RUnlock()
	return fd.buf
}

// Notify sets the file notification channel.
func (fd *FileData) Notify(ch chan string) {
	fd.Lock()
	fd.ch = ch
	fd.Unlock()
}

// Close resets the file offset.
func (fd *FileData) Close() error {
	fd.off.Store(0)
	return nil
}

// Read the current file offset.
func (fd *FileData) Read(b []byte) (n int, err error) {
	n, err = fd.ReadAt(b, fd.off.Load())
	fd.off.Add(int64(n))
	return
}

// ReadAt the given file offset.
func (fd *FileData) ReadAt(b []byte, off int64) (n int, err error) {
	fd.RLock()
	defer fd.RUnlock()

	if off < 0 || int64(int(off)) < off {
		return 0, ErrorInvalidOffset
	}

	if off > int64(len(fd.buf)) {
		return 0, io.EOF
	}

	n = copy(b, fd.buf[off:])

	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// ReadFrom the given io.Reader.
func (fd *FileData) ReadFrom(r io.Reader) (n int64, err error) {
	b, err := io.ReadAll(r)

	if err != nil {
		return 0, err
	}

	i, err := fd.Write(b)

	return int64(i), err
}

// Seek sets the current file offset.
func (fd *FileData) Seek(offset int64, whence int) (int64, error) {
	fd.RLock()
	defer fd.RUnlock()

	var abs int64

	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = fd.off.Load() + offset
	case io.SeekEnd:
		abs = int64(len(fd.buf)) + offset
	default:
		return 0, ErrorInvalidOffset
	}

	if abs < 0 {
		return 0, ErrorInvalidOffset
	}

	fd.off.Store(abs)

	return abs, nil
}

// Stat returns fs.FileInfo like stats.
func (fd *FileData) Stat() (fs.FileInfo, error) {
	return &FileInfo{fd}, nil
}

// Truncate the file to the given size.
func (fd *FileData) Truncate(size int64) error {
	fd.Lock()
	defer fd.Unlock()

	switch {
	case size < 0 || int64(int(size)) < size:
		return ErrorInvalidOffset
	case size <= int64(len(fd.buf)):
		fd.buf = fd.buf[:size]
	default:
		fd.buf = append(fd.buf, make([]byte, int(size)-len(fd.buf))...)
	}

	fd.mod = time.Now()

	if fd.ch != nil {
		fd.ch <- fd.name
	}

	return nil
}

// Write the given bytes at the current file offset.
func (fd *FileData) Write(b []byte) (n int, err error) {
	n, err = fd.WriteAt(b, fd.off.Load())
	fd.off.Add(int64(n))
	return
}

// WriteAt the given bytes and the given file offset.
func (fd *FileData) WriteAt(b []byte, off int64) (n int, err error) {
	fd.Lock()
	defer fd.Unlock()

	if off < 0 || int64(int(off)) < off {
		return 0, ErrorInvalidOffset
	}

	if off > int64(len(fd.buf)) {
		_ = fd.Truncate(off)
	}

	n = copy(fd.buf[off:], b)

	fd.buf = append(fd.buf, b[n:]...)
	fd.mod = time.Now()

	if fd.ch != nil {
		fd.ch <- fd.name
	}

	return len(b), nil
}

// WriteTo the given io.Writer.
func (fd *FileData) WriteTo(w io.Writer) (n int64, err error) {
	fd.RLock()
	defer fd.RUnlock()
	i, err := w.Write(fd.buf)
	return int64(i), err
}

// WriteString at the current file offset.
func (fd *FileData) WriteString(s string) (n int, err error) {
	return fd.Write([]byte(s))
}

// Name returns the file name.
func (fi *FileInfo) Name() string {
	return fi.fd.Name()
}

// Size returns the file size in bytes.
func (fi *FileInfo) Size() int64 {
	fi.fd.RLock()
	defer fi.fd.RUnlock()
	return int64(len(fi.fd.buf))
}

// Mode always returns 0 (regular).
func (fi *FileInfo) Mode() fs.FileMode {
	return 0 // regular
}

// ModTime returns the files last modified time.
func (fi *FileInfo) ModTime() time.Time {
	fi.fd.RLock()
	defer fi.fd.RUnlock()
	return fi.fd.mod
}

// IsDir always returns false.
func (fi *FileInfo) IsDir() bool {
	return false
}

// Sys always returns false.
func (fi *FileInfo) Sys() any {
	return nil
}

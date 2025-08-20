package memfile

import (
	"errors"
	"io"
	"io/fs"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Data provides an in-memory file data structure.
type Data struct {
	sync.RWMutex
	// Given filename.
	name string
	// File buffer.
	buf []byte
	// File offset.
	off atomic.Int64
	// File last modified time.
	mod time.Time
	// File watcher channel.
	evt chan fsnotify.Event
}

var (
	// ErrorInvalidOffset indicates an invalid offset
	ErrorInvalidOffset = errors.New("invalid offset")
)

// New returns a new file structure.
func New(name string) File {
	return &Data{name: name}
}

// Name returns the file name.
func (d *Data) Name() string {
	return d.name
}

// Bytes returns the file contents.
func (d *Data) Bytes() []byte {
	d.RLock()
	defer d.RUnlock()
	return d.buf
}

// Watch sets the used file watcher channel.
func (d *Data) Watch(ch chan fsnotify.Event) {
	d.Lock()
	d.evt = ch
	d.Unlock()
}

// Close resets the file offset.
func (d *Data) Close() error {
	d.off.Store(0)
	return nil
}

// Read the current file offset.
func (d *Data) Read(b []byte) (n int, err error) {
	n, err = d.ReadAt(b, d.off.Load())
	d.off.Add(int64(n))
	return
}

// ReadAt the given file offset.
func (d *Data) ReadAt(b []byte, off int64) (n int, err error) {
	d.RLock()
	defer d.RUnlock()

	if off < 0 || int64(int(off)) < off {
		return 0, ErrorInvalidOffset
	}

	if off > int64(len(d.buf)) {
		return 0, io.EOF
	}

	n = copy(b, d.buf[off:])

	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// ReadFrom the given io.Reader.
func (d *Data) ReadFrom(r io.Reader) (n int64, err error) {
	b, err := io.ReadAll(r)

	if err != nil {
		return 0, err
	}

	i, err := d.Write(b)

	return int64(i), err
}

// Seek sets the current file offset.
func (d *Data) Seek(offset int64, whence int) (int64, error) {
	d.RLock()
	defer d.RUnlock()

	var abs int64

	switch whence {
	case io.SeekStart:
		abs = offset

	case io.SeekCurrent:
		abs = d.off.Load() + offset

	case io.SeekEnd:
		abs = int64(len(d.buf)) + offset

	default:
		return 0, ErrorInvalidOffset
	}

	if abs < 0 {
		return 0, ErrorInvalidOffset
	}

	d.off.Store(abs)

	return abs, nil
}

// Stat returns the file statistics.
func (d *Data) Stat() (fs.FileInfo, error) {
	return &Info{data: d}, nil
}

// Truncate the file to the given size.
func (d *Data) Truncate(size int64) error {
	d.Lock()
	defer d.Unlock()

	switch {
	case size < 0 || int64(int(size)) < size:
		return ErrorInvalidOffset

	case size <= int64(len(d.buf)):
		d.buf = d.buf[:size]

	default:
		d.buf = append(d.buf, make([]byte, int(size)-len(d.buf))...)
	}

	d.mod = time.Now()

	d.notify()

	return nil
}

// Write the given bytes at the current file offset.
func (d *Data) Write(b []byte) (n int, err error) {
	n, err = d.WriteAt(b, d.off.Load())
	d.off.Add(int64(n))
	return
}

// WriteAt the given bytes and the given file offset.
func (d *Data) WriteAt(b []byte, off int64) (n int, err error) {
	d.Lock()
	defer d.Unlock()

	if off < 0 || int64(int(off)) < off {
		return 0, ErrorInvalidOffset
	}

	if off > int64(len(d.buf)) {
		_ = d.Truncate(off)
	}

	n = copy(d.buf[off:], b)

	d.buf = append(d.buf, b[n:]...)
	d.mod = time.Now()

	d.notify()

	return len(b), nil
}

// WriteTo the given io.Writer.
func (d *Data) WriteTo(w io.Writer) (n int64, err error) {
	d.RLock()
	defer d.RUnlock()
	i, err := w.Write(d.buf)
	return int64(i), err
}

// WriteString at the current file offset.
func (d *Data) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// notify the file watcher about changes.
func (d *Data) notify() {
	if d.evt != nil {
		d.evt <- fsnotify.Event{
			Name: d.name,
			Op:   fsnotify.Write,
		}
	}
}

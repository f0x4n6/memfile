package memfile

import (
	"io"
	"io/fs"
)

// File is an os.file compatible interface
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

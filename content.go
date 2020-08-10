package memphis

import (
	"io"
	"os"
	"sync"
)

// FileContent represents the actual data of a file.
type FileContent interface {
	Size() int64
	io.ReaderAt
	io.WriterAt
}

// TruncatableContents is an optional FileContent interface for efficiency
type TruncatableContents interface {
	Truncate(size int64) error
}

type memoryContents struct {
	bytes []byte
}

// NewEmptyFileContents creates a new memoryContents buffer
func NewEmptyFileContents() FileContent {
	m := memoryContents{bytes: []byte{}}
	return &m
}

func (m *memoryContents) Size() int64 {
	return int64(len(m.bytes))
}

func (m *memoryContents) WriteAt(p []byte, offset int64) (int, error) {
	if offset < 0 {
		return 0, os.ErrInvalid
	}

	prev := len(m.bytes)

	padding := int(offset) - prev
	if padding > 0 {
		m.bytes = append(m.bytes, make([]byte, padding)...)
	}

	m.bytes = append(m.bytes[:offset], p...)
	if len(m.bytes) < prev {
		m.bytes = m.bytes[:prev]
	}

	return len(p), nil
}

func (m *memoryContents) ReadAt(buf []byte, offset int64) (n int, err error) {
	if offset < 0 {
		return 0, os.ErrInvalid
	}

	size := int64(len(m.bytes))
	if offset >= size {
		return 0, io.EOF
	}

	l := int64(len(buf))
	if offset+l > size {
		l = size - offset
	}

	btr := m.bytes[offset : offset+l]
	if len(btr) < len(buf) {
		err = io.EOF
	}
	n = copy(buf, btr)

	return
}

func (m *memoryContents) Truncate(size int64) error {
	if int(size) == len(m.bytes) {
		return nil
	}
	if int(size) < len(m.bytes) {
		m.bytes = m.bytes[0:size]
	}

	extra := int(size) - len(m.bytes)
	ap := make([]byte, extra)
	m.bytes = append(m.bytes, ap...)

	return nil
}

// MemBufferFrom copies a file contents into a memoryContents format where it can be truncated
func MemBufferFrom(fc FileContent) FileContent {
	d := make([]byte, fc.Size())
	n := 0
	for n < len(d) {
		l, e := fc.ReadAt(d[n:], int64(n))
		n += l
		if e != nil {
			break
		}
	}
	return &memoryContents{d}
}

// copyOnWrite represents a FileContent that should on the first 'write' be copied via 'MemBufferFrom'
type copyOnWrite struct {
	FileContent
	sync.Once
}

func (c *copyOnWrite) WriteAt(p []byte, offset int64) (int, error) {
	c.Do(func() {
		c.FileContent = MemBufferFrom(c.FileContent)
	})
	return c.WriteAt(p, offset)
}

// osFileContent is a File Content backed by an on-disk file.
type osFileContent struct {
	path string
	size int64
}

func (o *osFileContent) Size() int64 {
	return o.size
}

func (o *osFileContent) ReadAt(buf []byte, offset int64) (n int, err error) {
	fp, err := os.Open(o.path)
	if err != nil {
		return 0, err
	}
	defer fp.Close()
	return fp.ReadAt(buf, offset)
}

func (o *osFileContent) WriteAt(p []byte, offset int64) (int, error) {
	fp, err := os.Open(o.path)
	if err != nil {
		return 0, err
	}
	defer fp.Close()
	return fp.WriteAt(p, offset)
}

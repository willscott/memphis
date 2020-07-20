package memphis

import (
	"io"

	"github.com/go-git/go-billy/v5"
)

// BillyFile is a wrapper to file contents implementing the implicit position cursor for read/write
type BillyFile struct {
	*File
	position int64
}

var _ billy.File = (*BillyFile)(nil)

// Lock is not used in this implementation
func (bf *BillyFile) Lock() error {
	return nil
}

// Unlock is not used in this implementation
func (bf *BillyFile) Unlock() error {
	return nil
}

// Truncate changes the size of the file contents
func (bf *BillyFile) Truncate(size int64) error {
	if size == 0 {
		bf.contents = NewEmptyFileContents()
		return nil
	}

	var truncatable TruncatableContents
	var ok bool
	if truncatable, ok = bf.contents.(TruncatableContents); !ok {
		bf.contents = MemBufferFrom(bf.contents)
		truncatable = bf.contents.(TruncatableContents)
	}
	truncatable.Truncate(size)

	return nil
}

// Close closes the file - not relevant in this implementation.
func (bf *BillyFile) Close() error {
	return nil
}

// ReadAt is as passthrough.
func (bf *BillyFile) ReadAt(buf []byte, offset int64) (n int, err error) {
	return bf.contents.ReadAt(buf, offset)
}

// Read is a more common implementation implemented by ReadAt
func (bf *BillyFile) Read(buf []byte) (n int, err error) {
	n, err = bf.ReadAt(buf, bf.position)
	bf.position += int64(n)
	return
}

// Write modifies file contents
func (bf *BillyFile) Write(buf []byte) (n int, err error) {
	n, err = bf.contents.WriteAt(buf, bf.position)
	bf.position += int64(n)
	return
}

// Seek changes file position
func (bf *BillyFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		bf.position += offset
	case io.SeekStart:
		bf.position = offset
	case io.SeekEnd:
		bf.position = bf.contents.Size() - offset
	}
	// validate
	if bf.position < 0 {
		return -1, io.EOF
	}

	return bf.position, nil
}

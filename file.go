package memphis

import (
	"io"
	"os"
	"time"
)

// File holds the metadata of a FS object
type File struct {
	name       string
	mode       os.FileMode
	uid        uint32
	gid        uint32
	createTime time.Time
	modTime    time.Time
	contents   FileContent
}

// Name returns the file name
func (f *File) Name() string {
	return f.name
}

// Size returns the file's size
func (f *File) Size() int64 {
	return f.contents.Size()
}

// Mode returns the file's mode
func (f *File) Mode() os.FileMode {
	return f.mode
}

// ModTime returns when the file was modified
func (f *File) ModTime() time.Time {
	return f.modTime
}

// IsDir returns if the file is a directory (no)
func (f *File) IsDir() bool {
	return false
}

// Sys is a wildcard in the OS interface
func (f *File) Sys() interface{} {
	return nil
}

// Bytes returns a direct buffer of the contents of the file
func (f *File) Bytes() []byte {
	if f.contents == nil {
		return []byte{}
	}
	n := int64(0)
	l := f.contents.Size()
	buf := make([]byte, l)
	for n < l {
		a, e := f.contents.ReadAt(buf[n:], n)
		n += int64(a)
		if n == l || e == io.EOF {
			return buf
		}
		if e != nil {
			return []byte{}
		}
		if a == 0 {
			return []byte{}
		}
	}
	return buf
}

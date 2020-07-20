package memphis

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
)

// Billy wraps a filesystem subtree in the billy filesystem interface
type Billy struct {
	euid uint32
	egid uint32
	root *Tree
}

// Create makes a new empty file
func (b *Billy) Create(filename string) (billy.File, error) {
	dir, name := path.Split(filename)
	treeRef := b.root.WalkDir(strings.Split(dir, string(os.PathSeparator)))
	if treeRef == nil {
		return nil, ErrNotDir
	}

	if _, ok := treeRef.files[name]; ok {
		return nil, ErrExists
	}

	return nil, nil
}

// Open is a shortcut to openfile
func (b *Billy) Open(filename string) (billy.File, error) {
	return b.OpenFile(filename, os.O_RDONLY, 0666)
}

// OpenFile opens a file for access
func (b *Billy) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	dir, name := path.Split(filename)
	parent := b.root.WalkDir(strings.Split(dir, string(os.PathSeparator)))
	if parent == nil {
		return nil, os.ErrNotExist
	}
	if _, dok := parent.directories[name]; dok {
		return nil, os.ErrExist
	}

	// If create mode - need parent but not file.
	if (flag & os.O_CREATE) != 0 {
		if _, fok := parent.files[name]; fok {
			return nil, os.ErrExist
		}
		return b.Create(filename)
	}

	f, err := b.root.Get(strings.Split(filename, string(os.PathSeparator)))
	if err != nil {
		return nil, err
	}

	return &BillyFile{f, 0}, nil
}

// Stat returns file metadata
func (b *Billy) Stat(filename string) (os.FileInfo, error) {
	dir, name := path.Split(filename)
	parent := b.root.WalkDir(strings.Split(dir, string(os.PathSeparator)))
	if parent == nil {
		return nil, os.ErrNotExist
	}
	if f, ok := parent.files[name]; ok {
		return f, nil
	}
	if d, ok := parent.directories[name]; ok {
		return &DirMeta{name, &d}, nil
	}
	return nil, os.ErrNotExist
}

// Rename a file
func (b *Billy) Rename(oldpath, newpath string) error {
	return nil
}

// Remove deletes a file
func (b *Billy) Remove(filename string) error {
	return nil
}

// Join constructs a path
func (b *Billy) Join(elem ...string) string {
	return path.Join(elem...)
}

// TempFile create an empty tempfile
func (b *Billy) TempFile(dir, prefix string) (billy.File, error) {
	return nil, nil
}

// ReadDir lists directory contents
func (b *Billy) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, nil
}

// MkdirAll creates a new directory
func (b *Billy) MkdirAll(filename string, perm os.FileMode) error {
	return nil
}

// Lstat provides symlink info
func (b *Billy) Lstat(filename string) (os.FileInfo, error) {
	return nil, nil
}

// Symlink creates a symbolic link
func (b *Billy) Symlink(target, link string) error {
	return nil
}

// Readlink returns symlink contents
func (b *Billy) Readlink(link string) (string, error) {
	return "", nil
}

// Chmod changes file permissions
func (b *Billy) Chmod(name string, mode os.FileMode) error {
	return nil
}

// Lchown changes symlink ownership
func (b *Billy) Lchown(name string, uid, gid int) error {
	return nil
}

// Chown chagnes file ownership
func (b *Billy) Chown(name string, uid, gid int) error {
	return nil
}

// Chtimes changes file access time
func (b *Billy) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return nil
}

// Chroot returns a subtree of the filesystem
func (b *Billy) Chroot(path string) (billy.Filesystem, error) {
	dir := b.root.WalkDir(strings.Split(path, string(os.PathSeparator)))
	if dir == nil {
		return nil, os.ErrNotExist
	}
	return &Billy{b.euid, b.egid, dir}, nil
}

// Root prints the path of the current fs root
func (b *Billy) Root() string {
	return string(os.PathSeparator)
}

// Capabilities tells billy what this FS can do
func (b *Billy) Capabilities() billy.Capability {
	return billy.AllCapabilities
}

var _ billy.Filesystem = (*Billy)(nil)

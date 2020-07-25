package memphis

import (
	"fmt"
	"math/rand"
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

	f := treeRef.Create(name, b.euid, b.egid, 0666)

	return &BillyFile{f, 0}, nil
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
	return b.getFileInfo(filename, true)
}

// Lstat provides symlink info
func (b *Billy) Lstat(filename string) (os.FileInfo, error) {
	return b.getFileInfo(filename, false)
}

func (b *Billy) getFileInfo(filename string, followLinks bool) (os.FileInfo, error) {
	dir, name := path.Split(filename)
	parent := b.root.WalkDir(strings.Split(dir, string(os.PathSeparator)))
	if parent == nil {
		return nil, os.ErrNotExist
	}
	if f, ok := parent.files[name]; ok {
		if ((f.Mode() & os.ModeSymlink) != 0) && followLinks {
			// todo: maybe resolve circular links?
			return b.getFileInfo(string(f.Bytes()), true)
		}
		return f, nil
	}
	if d, ok := parent.directories[name]; ok {
		return &DirMeta{name, d}, nil
	}
	return nil, os.ErrNotExist

}

// Rename a file
func (b *Billy) Rename(oldpath, newpath string) error {
	oldDir, oldName := path.Split(oldpath)
	oldParent := b.root.WalkDir(strings.Split(oldDir, string(os.PathSeparator)))
	if oldParent == nil {
		return os.ErrNotExist
	}

	newDir, newName := path.Split(newpath)
	newParent := b.root.WalkDir(strings.Split(newDir, string(os.PathSeparator)))
	if newParent == nil {
		return os.ErrNotExist
	}
	if _, ok := newParent.files[newName]; ok {
		return os.ErrExist
	}
	if _, ok := newParent.directories[newName]; ok {
		return os.ErrExist
	}

	// TODO: permissions check.

	if f, ok := oldParent.files[oldName]; ok {
		// move file.
		newParent.files[newName] = f
		delete(oldParent.files, oldName)
		return nil
	} else if d, ok := oldParent.directories[oldName]; ok {
		// move dir.
		newParent.directories[newName] = d
		delete(oldParent.directories, newName)
		return nil
	}
	return os.ErrNotExist
}

// Remove deletes a file
func (b *Billy) Remove(filename string) error {
	dir, name := path.Split(filename)
	parent := b.root.WalkDir(strings.Split(dir, string(os.PathSeparator)))
	if _, ok := parent.files[name]; ok {
		// TODO: permissions.
		delete(parent.files, name)
		return nil
	}

	if d, ok := parent.directories[name]; ok {
		// TODO: permissions.

		// Directory must be empty
		if len(d.files) > 0 || len(d.directories) > 0 {
			return os.ErrExist
		}
		delete(parent.directories, name)
		return nil
	}

	return os.ErrNotExist
}

// Join constructs a path
func (b *Billy) Join(elem ...string) string {
	return path.Join(elem...)
}

// TempFile create an empty tempfile
func (b *Billy) TempFile(dir, prefix string) (billy.File, error) {
	d := b.root.WalkDir(strings.Split(dir, string(os.PathSeparator)))
	if d == nil {
		return nil, os.ErrNotExist
	}
	r := rand.Int()
	n := fmt.Sprintf("%s%d", prefix, r)
	return b.Create(path.Join(dir, n))
}

// ReadDir lists directory contents
func (b *Billy) ReadDir(path string) ([]os.FileInfo, error) {
	d := b.root.WalkDir(strings.Split(path, string(os.PathSeparator)))
	if d == nil {
		return nil, os.ErrNotExist
	}
	items := make([]os.FileInfo, 0, len(d.directories)+len(d.files))
	for name, dir := range d.directories {
		items = append(items, &DirMeta{name, dir})
	}
	for _, file := range d.files {
		items = append(items, file)
	}
	return items, nil
}

// MkdirAll creates a new directory
func (b *Billy) MkdirAll(filename string, perm os.FileMode) error {
	parts := strings.Split(filename, string(os.PathSeparator))
	cur := b.root
	for _, p := range parts {
		// todo: permissions
		if next, ok := cur.directories[p]; ok {
			cur = next
		} else {
			cur = cur.CreateDir(p, b.euid, b.egid, perm.Perm()|os.ModeDir)
		}
	}
	return nil
}

// Symlink creates a symbolic link
func (b *Billy) Symlink(target, link string) error {
	f, err := b.Create(link)
	if err != nil {
		return err
	}
	bf := f.(*BillyFile)
	bf.Write([]byte(target))
	bf.File.mode |= os.ModeSymlink
	return nil
}

// Readlink returns symlink contents
func (b *Billy) Readlink(link string) (string, error) {
	f, err := b.getFileInfo(link, false)
	if err != nil {
		return "", err
	}
	//todo: validate read permission.
	ffile, ok := f.(*File)
	if !ok {
		return "", os.ErrExist
	}
	return string(ffile.Bytes()), nil
}

// Chmod changes file permissions
func (b *Billy) Chmod(name string, mode os.FileMode) error {
	f, err := b.getFileInfo(name, true)
	if err != nil {
		return err
	}
	ffile, ok := f.(*File)
	if !ok {
		fdir, ok := f.(*DirMeta)
		if !ok {
			return os.ErrInvalid
		}
		fdir.Tree.mode = mode
		return nil
	}
	ffile.mode = mode
	return nil
}

// Lchown changes symlink ownership
func (b *Billy) Lchown(name string, uid, gid int) error {
	return b.changeOwnership(name, uid, gid, false)
}

// Chown chagnes file ownership
func (b *Billy) Chown(name string, uid, gid int) error {
	return b.changeOwnership(name, uid, gid, true)
}

func (b *Billy) changeOwnership(name string, uid, gid int, followLinks bool) error {
	if b.euid != 0 && b.egid != 0 {
		return os.ErrPermission
	}

	f, err := b.getFileInfo(name, followLinks)
	if err != nil {
		return err
	}
	ffile, ok := f.(*File)
	if !ok {
		fdir, ok := f.(*DirMeta)
		if !ok {
			return os.ErrInvalid
		}
		fdir.Tree.uid = uint32(uid)
		fdir.Tree.gid = uint32(gid)
		return nil
	}
	ffile.uid = uint32(uid)
	ffile.gid = uint32(gid)

	return nil
}

// Chtimes changes file access time
func (b *Billy) Chtimes(name string, atime time.Time, mtime time.Time) error {
	f, err := b.getFileInfo(name, true)
	if err != nil {
		return err
	}
	ffile, ok := f.(*File)
	if !ok {
		fdir, ok := f.(*DirMeta)
		if !ok {
			return os.ErrInvalid
		}
		fdir.Tree.modTime = mtime
		return nil
	}
	ffile.modTime = mtime

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

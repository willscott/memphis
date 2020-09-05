package memphis

import (
	"os"
	"strings"
	"time"

	"github.com/polydawn/rio/fs"
)

// Placer conforms a memphis directory tree to the rio FS interface
type Placer struct {
	root *Tree
}

// BasePath is the root of this FS - always '/'
func (p *Placer) BasePath() fs.AbsolutePath {
	return fs.MustAbsolutePath(string(os.PathSeparator))
}

// OpenFile attempts to open a file
func (p *Placer) OpenFile(path fs.RelPath, flag int, perms fs.Perms) (fs.File, error) {
	f, err := p.root.Get(strings.Split(path.String(), "/"))
	if err != nil {
		return nil, err
	}
	return &BillyFile{f, 0}, nil
}

func permsToOs(perms fs.Perms) (mode os.FileMode) {
	mode = os.FileMode(perms & 0777)
	if perms&04000 != 0 {
		mode |= os.ModeSetuid
	}
	if perms&02000 != 0 {
		mode |= os.ModeSetgid
	}
	if perms&01000 != 0 {
		mode |= os.ModeSticky
	}
	return mode
}

// Mkdir makes a directory at path
func (p *Placer) Mkdir(path fs.RelPath, perms fs.Perms) error {
	parent := p.root.WalkDir(strings.Split(path.Dir().String(), "/"))
	if parent == nil {
		return os.ErrNotExist
	}
	fname := path.Last()
	if _, ok := parent.files[fname]; ok {
		return os.ErrExist
	}
	if _, ok := parent.directories[fname]; ok {
		return os.ErrExist
	}
	// todo: perms

	dir := parent.CreateDir(fname, parent.uid, parent.gid, permsToOs(perms)|os.ModeDir)
	if dir == nil {
		return os.ErrInvalid
	}
	return nil
}

// Mklink makes a symlink at path
func (p *Placer) Mklink(path fs.RelPath, target string) error {
	f, err := p.Create(path)
	if err != nil {
		return err
	}
	bf := f.(*BillyFile)
	bf.Write([]byte(target))
	bf.File.mode |= os.ModeSymlink
	return nil
}

// Mkfifo makes a fifo node at path
func (p *Placer) Mkfifo(path fs.RelPath, perms fs.Perms) error {

}

// MkdevBlock makes a block device at path
func (p *Placer) MkdevBlock(path fs.RelPath, major int64, minor int64, perms fs.Perms) error {

}

// MkdevChar makes a character device at path
func (p *Placer) MkdevChar(path fs.RelPath, major int64, minor int64, perms fs.Perms) error {

}

// Lchown sets ownership of path w/o following symlinks
func (p *Placer) Lchown(path fs.RelPath, uid uint32, gid uint32) error {

}

// Chmod sets permissions of path
func (p *Placer) Chmod(path fs.RelPath, perms fs.Perms) error {

}

// SetTimesLNano sets modification/access times of path
func (p *Placer) SetTimesLNano(path fs.RelPath, mtime time.Time, atime time.Time) error {

}

// SetTimesNano sets modification/access times of path
func (p *Placer) SetTimesNano(path fs.RelPath, mtime time.Time, atime time.Time) error {

}

// Stat returns file metadata
func (p *Placer) Stat(path fs.RelPath) (*fs.Metadata, error) {

}

// LStat returns file metadata not following symlinks
func (p *Placer) LStat(path fs.RelPath) (*fs.Metadata, error) {

}

// ReadDirNames lists files in a directory
func (p *Placer) ReadDirNames(path fs.RelPath) ([]string, error) {

}

// Readlink reads a symlink
func (p *Placer) Readlink(path fs.RelPath) (target string, isSymlink bool, err error) {

}

// ResolveLink resolves a symlink
func (p *Placer) ResolveLink(symlink string, startingAt fs.RelPath) (fs.RelPath, error) {

}

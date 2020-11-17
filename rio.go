package memphis

import (
	"encoding/binary"
	"fmt"
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
	return fs.MustAbsolutePath(Separator)
}

// OpenFile attempts to open a file
func (p *Placer) OpenFile(path fs.RelPath, flag int, perms fs.Perms) (fs.File, error) {
	f, d, err := p.root.Get(strings.Split(path.String(), Separator), true)
	if err != nil {
		return nil, err
	}
	if d != nil {
		return nil, os.ErrExist
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

func modeToPerms(mode os.FileMode) (perms fs.Perms) {
	perms = fs.Perms(mode & 0777)
	if mode&os.ModeSetuid != 0 {
		perms |= 04000
	}
	if mode&os.ModeSetgid != 0 {
		perms |= 02000
	}
	if mode&os.ModeSticky != 0 {
		perms |= 01000
	}
	return perms
}

// Mkdir makes a directory at path
func (p *Placer) Mkdir(path fs.RelPath, perms fs.Perms) error {
	parent := p.root.WalkDir(strings.Split(path.Dir().String(), Separator))
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
	f, err := p.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
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
	f, err := p.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return err
	}
	bf := f.(*BillyFile)
	bf.Write([]byte{})
	bf.File.mode |= os.ModeNamedPipe
	return nil
}

// MkdevBlock makes a block device at path
func (p *Placer) MkdevBlock(path fs.RelPath, major int64, minor int64, perms fs.Perms) error {
	f, err := p.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return err
	}
	bf := f.(*BillyFile)
	buf := [16]byte{}
	binary.LittleEndian.PutUint64(buf[0:8], uint64(major))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(minor))
	bf.Write(buf[:])
	bf.File.mode |= os.ModeDevice
	return nil
}

// MkdevChar makes a character device at path
func (p *Placer) MkdevChar(path fs.RelPath, major int64, minor int64, perms fs.Perms) error {
	f, err := p.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return err
	}
	bf := f.(*BillyFile)
	buf := [16]byte{}
	binary.LittleEndian.PutUint64(buf[0:8], uint64(major))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(minor))
	bf.Write(buf[:])
	bf.File.mode |= os.ModeCharDevice
	return nil
}

// Lchown sets ownership of path w/o following symlinks
func (p *Placer) Lchown(path fs.RelPath, uid uint32, gid uint32) error {
	f, d, err := p.root.Get(strings.Split(path.String(), Separator), false)
	if err != nil {
		return err
	}

	if f != nil {
		f.uid = uid
		f.gid = gid
	} else if d != nil {
		d.uid = uid
		d.gid = gid
	}
	return nil
}

const nonPermModeBits = ^(os.ModePerm | os.ModeSetgid | os.ModeSetuid | os.ModeSticky)

// Chmod sets permissions of path
func (p *Placer) Chmod(path fs.RelPath, perms fs.Perms) error {
	f, d, err := p.root.Get(strings.Split(path.String(), Separator), true)
	if err != nil {
		return err
	}

	mode := permsToOs(perms)
	if f != nil {
		f.mode = (f.mode & nonPermModeBits) | mode
	} else if d != nil {
		d.mode = (d.mode & nonPermModeBits) | mode
	}
	return nil
}

// SetTimesLNano sets modification/access times of path
func (p *Placer) SetTimesLNano(path fs.RelPath, mtime time.Time, atime time.Time) error {
	f, d, err := p.root.Get(strings.Split(path.String(), Separator), false)
	if err != nil {
		return err
	}

	if f != nil {
		f.modTime = mtime
	} else {
		d.modTime = mtime
	}
	return nil
}

// SetTimesNano sets modification/access times of path
func (p *Placer) SetTimesNano(path fs.RelPath, mtime time.Time, atime time.Time) error {
	f, d, err := p.root.Get(strings.Split(path.String(), Separator), true)
	if err != nil {
		return err
	}

	if f != nil {
		f.modTime = mtime
	} else {
		d.modTime = mtime
	}
	return nil
}

func toMetadata(f *File) *fs.Metadata {
	md := &fs.Metadata{
		Name:  fs.MustRelPath(f.Name()),
		Type:  fs.Type_File,
		Perms: modeToPerms(f.Mode()),
		Uid:   f.uid,
		Gid:   f.gid,
		Size:  f.Size(),
		Mtime: f.ModTime(),
	}

	if f.mode&os.ModeSymlink != 0 {
		md.Type = fs.Type_Symlink
		md.Linkname = string(f.Bytes())
	}

	if f.mode&(os.ModeCharDevice|os.ModeDevice) != 0 {
		if f.mode&os.ModeCharDevice != 0 {
			md.Type = fs.Type_CharDevice
		} else {
			md.Type = fs.Type_Device
		}
		b := f.Bytes()
		md.Devmajor = int64(binary.LittleEndian.Uint64(b[0:8]))
		md.Devminor = int64(binary.LittleEndian.Uint64(b[8:16]))
	}

	if f.mode&os.ModeNamedPipe != 0 {
		md.Type = fs.Type_NamedPipe
	}

	return md
}

func dirMetadata(path fs.RelPath, d *Tree) *fs.Metadata {
	return &fs.Metadata{
		Name:  path,
		Type:  fs.Type_Dir,
		Perms: modeToPerms(d.mode),
		Uid:   d.uid,
		Gid:   d.gid,
		Size:  0,
		Mtime: d.modTime,
	}
}

// Stat returns file metadata
func (p *Placer) Stat(path fs.RelPath) (*fs.Metadata, error) {
	f, d, err := p.root.Get(strings.Split(path.String(), Separator), true)
	if err != nil {
		return nil, err
	}

	if f != nil {
		return toMetadata(f), nil
	}
	return dirMetadata(path, d), nil
}

// LStat returns file metadata not following symlinks
func (p *Placer) LStat(path fs.RelPath) (*fs.Metadata, error) {
	f, d, err := p.root.Get(strings.Split(path.String(), Separator), true)
	if err != nil {
		return nil, err
	}

	if f != nil {
		return toMetadata(f), nil
	}
	return dirMetadata(path, d), nil
}

// ReadDirNames lists files in a directory
func (p *Placer) ReadDirNames(path fs.RelPath) ([]string, error) {
	_, d, err := p.root.Get(strings.Split(path.String(), Separator), true)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, os.ErrExist
	}
	names := make([]string, 0, len(d.directories)+len(d.files))
	for dn := range d.directories {
		names = append(names, dn)
	}
	for fn := range d.files {
		names = append(names, fn)
	}
	return names, nil
}

// Readlink reads a symlink
func (p *Placer) Readlink(path fs.RelPath) (target string, isSymlink bool, err error) {
	f, _, err := p.root.Get(strings.Split(path.String(), Separator), false)
	if err != nil {
		return "", false, err
	}
	if f == nil {
		return "", false, os.ErrExist
	}
	if f.mode&os.ModeSymlink == 0 {
		return "", false, nil
	}
	return string(f.Bytes()), true, nil
}

// ResolveLink resolves a symlink
func (p *Placer) ResolveLink(symlink string, startingAt fs.RelPath) (fs.RelPath, error) {
	if startingAt.GoesUp() {
		return startingAt, fmt.Errorf("%s", fs.ErrBreakout)
	}
	return p.resolveLink(symlink, startingAt, map[fs.RelPath]struct{}{})
}

func (p *Placer) resolveLink(symlink string, startingAt fs.RelPath, seen map[fs.RelPath]struct{}) (fs.RelPath, error) {
	if _, isSeen := seen[startingAt]; isSeen {
		return startingAt, fmt.Errorf("%s", fs.ErrRecursion)
	}
	seen[startingAt] = struct{}{}
	segments := strings.Split(symlink, "/")
	path := startingAt
	if segments[0] == "" { // rooted
		path = fs.RelPath{}
		segments = segments[1:]
	} else {
		path = startingAt.Dir() // drop the link node itself
	}
	iLast := len(segments) - 1
	for i, s := range segments {
		// Identity segments can simply be skipped.
		if s == "" || s == "." {
			continue
		}
		// Excessive up segements aren't an error; they simply no-op when already at root.
		if s == ".." && path == (fs.RelPath{}) {
			continue
		}
		// Okay, join the segment and peek at it.
		path = path.Join(fs.MustRelPath(s))
		// Bail on cycles before considering recursion!
		if path == startingAt {
			return startingAt, fmt.Errorf("%s", fs.ErrRecursion)
		}
		// Check if this is a symlink; if so we must recurse on it.
		morelink, isLink, err := p.Readlink(fs.MustRelPath(p.BasePath().Join(path).String()))
		if err != nil {
			if i == iLast && os.IsNotExist(err) {
				return path, nil
			}
			return startingAt, fs.NormalizeIOError(err)
		}
		if isLink {
			path, err = p.resolveLink(morelink, path, seen)
			if err != nil {
				return startingAt, err
			}
		}
	}
	return path, nil
}

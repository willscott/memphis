package memphis

import (
	"os"
	"strings"
	"sync"
	"time"
)

// Tree represents a directory
type Tree struct {
	ready       sync.Once
	deferred    func()
	uid         uint32
	gid         uint32
	mode        os.FileMode
	directories map[string]*Tree
	files       map[string]*File
	createTime  time.Time
	modTime     time.Time
}

// Create makes a new file in the directory
func (t *Tree) Create(name string, euid, egid uint32, perm os.FileMode) *File {
	t.ready.Do(t.deferred)
	t.files[name] = &File{
		name:       name,
		mode:       perm,
		uid:        euid,
		gid:        egid,
		createTime: time.Now(),
		modTime:    time.Now(),
		contents:   NewEmptyFileContents(),
	}
	return t.files[name]
}

func noOp() {}

// CreateDir makes a new directory in the directory
func (t *Tree) CreateDir(name string, euid, egid uint32, perm os.FileMode) *Tree {
	t.directories[name] = &Tree{
		deferred:    noOp,
		uid:         euid,
		gid:         egid,
		mode:        perm,
		directories: make(map[string]*Tree),
		files:       make(map[string]*File),
		createTime:  time.Now(),
		modTime:     time.Now(),
	}
	return t.directories[name]
}

// WalkDir descends to a given sub directory
func (t *Tree) WalkDir(p []string) *Tree {
	t.ready.Do(t.deferred)

	node := t
	path := []string{}

	remaining := p
	for len(remaining) > 0 {
		part := remaining[0]
		remaining = remaining[1:]
		if part == "." || part == "" {
			continue
		}
		if part == ".." {
			if len(path) > 0 {
				path = path[:len(path)-1]
				// TODO: circular symlinks will infiniloop
				node = t.WalkDir(path)
				continue
			} else {
				return nil //directory traversal
			}
		}
		n, ok := node.directories[part]
		if ok {
			node = n
			n.ready.Do(n.deferred)
			path = append(path, part)
			continue
		}
		f, ok := node.files[part]
		if ok && f.Mode()&os.ModeSymlink != 0 {
			target := strings.Split(string(f.Bytes()), Separator)
			if target[0] == "" {
				target = target[1:]
				path = []string{}
				node = t
			}
			remaining = append(target, remaining...)
			continue
		}

		return nil
	}
	return node
}

// Get attempts to get a file at a given path.
func (t *Tree) Get(p []string, followSymlinks bool) (*File, *Tree, error) {
	t.ready.Do(t.deferred)
	if len(p) == 0 {
		return nil, nil, os.ErrNotExist
	}
	fname := p[len(p)-1]
	base := t
	if len(p) > 1 {
		base = t.WalkDir(p[:len(p)-1])
	}
	if base == nil {
		return nil, nil, os.ErrNotExist
	}

	if d, ok := base.directories[fname]; ok {
		return nil, d, nil
	}

	f, ok := base.files[fname]
	if !ok {
		return nil, nil, os.ErrNotExist
	}
	// Check permissions
	// TODO

	// Resolve symlinks
	if followSymlinks && ((f.Mode() & os.ModeSymlink) != 0) {
		target := strings.Split(string(f.Bytes()), Separator)
		if target[0] == "" {
			return t.Get(target[1:], true)
		}
		return t.Get(append(p[0:len(p)-1], target...), true)
	}
	return f, nil, nil
}

// DirMeta is a struct of metadata about a directory
type DirMeta struct {
	name string
	*Tree
}

// Name of the directory
func (d *DirMeta) Name() string {
	return d.name
}

// Size of the directory
func (d *DirMeta) Size() int64 {
	return 0
}

// Mode is the os.FileMode (permissions) for the directory
func (d *DirMeta) Mode() os.FileMode {
	return d.Tree.mode
}

// ModTime is when the directory was last modified
func (d *DirMeta) ModTime() time.Time {
	return d.Tree.modTime
}

// IsDir is true for directories
func (d *DirMeta) IsDir() bool {
	return true
}

// Sys provides os.FileInfo trapdoor down to undefined behavior
func (d *DirMeta) Sys() interface{} {
	return nil
}

package memphis

import (
	"io/ioutil"
	"os"
	"path"
)

func deferredOSDir(dir *Tree, dirPath string) func() {
	return func() {
		info, err := os.Stat(dirPath)
		if err != nil {
			return
		}
		dir.modTime = info.ModTime()
		dir.createTime = dir.modTime

		// hacky retreaval of uid/guid from os
		osStat(dir, info.Sys())

		dir.mode = info.Mode()

		files, err := ioutil.ReadDir(dirPath)
		if err != nil {
			return
		}

		for _, f := range files {
			if f.IsDir() {
				child := newTree(dir.uid, dir.gid, dir.mode)
				child.deferred = deferredOSDir(child, path.Join(dirPath, f.Name()))
				dir.directories[f.Name()] = child
			} else {
				dir.files[f.Name()] = FileFromOS(path.Join(dirPath, f.Name()), dir.uid, dir.gid, f)
			}
		}
	}
}

// FileFromOS creates a file representing an underlying OS file.
// writes will transition the file contents to a memory buffer.
func FileFromOS(path string, uid, gid uint32, info os.FileInfo) *File {
	f := File{
		name:       info.Name(),
		mode:       info.Mode(),
		uid:        uid,
		gid:        gid,
		createTime: info.ModTime(),
		modTime:    info.ModTime(),
	}

	osStatFile(&f, info.Sys())

	f.contents = &copyOnWrite{FileContent: &osFileContent{path: path, size: info.Size()}}
	return &f
}

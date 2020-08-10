package memphis

import (
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"time"
)

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func deferredOSDir(dir *Tree, dirPath string) func() {
	return func() {
		info, err := os.Stat(dirPath)
		if err != nil {
			return
		}
		dir.modTime = info.ModTime()
		dir.createTime = dir.modTime

		// hacky retreaval of uid/guid from os
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			dir.uid = stat.Uid
			dir.gid = stat.Gid
			// TODO: sadly, the name of this field is inconsistent. likely needs deferral to per-goos impls
			dir.createTime = timespecToTime(stat.Ctimespec)
		}

		dir.mode = info.Mode()

		files, err := ioutil.ReadDir(dirPath)
		if err != nil {
			return
		}

		for _, f := range files {
			if f.IsDir() {
				child := dir.CreateDir(f.Name(), dir.uid, dir.gid, dir.mode)
				child.deferred = deferredOSDir(child, path.Join(dirPath, f.Name()))
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

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		f.uid = stat.Uid
		f.gid = stat.Gid
		// TODO: sadly, the name of this field is inconsistent. likely needs deferral to per-goos impls
		f.createTime = timespecToTime(stat.Ctimespec)
	}

	f.contents = &copyOnWrite{FileContent: &osFileContent{path: path, size: info.Size()}}
	return &f
}

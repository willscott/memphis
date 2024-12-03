//go:build windows
// +build windows

package memphis

import (
	"syscall"
	"time"
)

func osStat(d *Tree, stat any) {
	winStat := stat.(*syscall.Win32FileAttributeData)
	// todo: uid/gid
	d.createTime = time.Unix(0, winStat.CreationTime.Nanoseconds())
}

func osStatFile(f *File, stat any) {
	winStat := stat.(*syscall.Win32FileAttributeData)
	// todo: uid/gid
	f.createTime = time.Unix(0, winStat.CreationTime.Nanoseconds())
}

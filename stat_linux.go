//go:build linux
// +build linux

package memphis

import (
	"syscall"
	"time"
)

func osStat(d *Tree, stat any) {
	unixStat := stat.(*syscall.Stat_t)
	d.uid = unixStat.Uid
	d.gid = unixStat.Gid
	ts := unixStat.Ctim
	d.createTime = time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func osStatFile(f *File, stat any) {
	unixStat := stat.(*syscall.Stat_t)
	f.uid = unixStat.Uid
	f.gid = unixStat.Gid
	ts := unixStat.Ctim
	f.createTime = time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

//go:build darwin || dragonfly || freebsd || netbsd || openbsd
// +build darwin dragonfly freebsd netbsd openbsd

package memphis

import (
	"syscall"
	"time"
)

func osStat(d *Tree, stat any) {
	unixStat := stat.(*syscall.Stat_t)
	d.uid = unixStat.Uid
	d.gid = unixStat.Gid
	ts := unixStat.Ctimespec
	d.createTime = time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func osStatFile(f *File, stat any) {
	unixStat := stat.(*syscall.Stat_t)
	f.uid = unixStat.Uid
	f.gid = unixStat.Gid
	ts := unixStat.Ctimespec
	f.createTime = time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

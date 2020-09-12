// +build darwin dragonfly freebsd netbsd openbsd

package memphis

import (
	"syscall"
	"time"
)

func createTime(s *syscall.Stat_t) time.Time {
	ts := s.Ctimespec
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

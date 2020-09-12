// +build linux

package memphis

import (
	"syscall"
	"time"
)

func createTime(s *syscall.Stat_t) time.Time {
	ts := s.Ctim
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

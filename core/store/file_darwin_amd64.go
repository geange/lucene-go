//go:build amd64 && darwin

package store

import (
	"os"
	"syscall"
	"time"
)

func FileTime(info os.FileInfo) (access, create, modify time.Time) {
	stat_t := info.Sys().(*syscall.Stat_t)
	return timespecToTime(stat_t.Atimespec),
		timespecToTime(stat_t.Ctimespec),
		timespecToTime(stat_t.Mtimespec)
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

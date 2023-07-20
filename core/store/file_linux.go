//go:build linux

package store

import (
	"os"
	"syscall"
	"time"
)

func FileTime(info os.FileInfo) (access, create, modify time.Time) {
	stat_t := info.Sys().(*syscall.Stat_t)
	return timespecToTime(stat_t.Atim),
		timespecToTime(stat_t.Ctim),
		timespecToTime(stat_t.Mtim)
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

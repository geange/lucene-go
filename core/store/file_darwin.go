//go:build darwin

package store

import (
	"os"
	"syscall"
	"time"
)

func FileTime(info os.FileInfo) (access, create, modify time.Time) {
	statT := info.Sys().(*syscall.Stat_t)
	return timespecToTime(statT.Atimespec),
		timespecToTime(statT.Ctimespec),
		timespecToTime(statT.Mtimespec)
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

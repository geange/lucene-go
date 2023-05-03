//go:build windows

package store

import (
	"os"
	"syscall"
	"time"
)

func FileTime(info os.FileInfo) (access, create, modify time.Time) {
	filetime := info.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, filetime.LastAccessTime.Nanoseconds()).Local(),
		time.Unix(0, filetime.CreationTime.Nanoseconds()).Local(),
		time.Unix(0, filetime.LastWriteTime.Nanoseconds()).Local()
}

//go:build linux
// +build linux

package util

import (
	"syscall"
)

func GetDefaultAppDir() string {
	return "/apps/data"
}

type DiskStatus struct {
	All  uint64 `json:"all"`  // 总大小B
	Used uint64 `json:"used"` //使用B
	Free uint64 `json:"free"` //剩余B
}

// disk usage of path/disk
func GetDiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bavail * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

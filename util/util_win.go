//go:build windows
// +build windows

package util

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"os/exec"
	"syscall"
	"unsafe"
)

func GetDefaultAppDir() string {
	return GetAppPath() + "/data"
}

type DiskStatus struct {
	All  uint64 `json:"all"`  //总大小B
	Used uint64 `json:"used"` //使用B
	Free uint64 `json:"free"` //剩余B
}

// GetDiskUsage 获取磁盘空间,"c:"
func GetDiskUsage(path string) (disk DiskStatus) {
	h := windows.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailable := uint64(0)
	lpTotalNumberOfBytes := uint64(0)
	lpTotalNumberOfFreeBytes := uint64(0)
	c.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)))
	disk.All = lpTotalNumberOfBytes
	disk.Free = lpFreeBytesAvailable
	disk.Used = lpTotalNumberOfBytes - lpFreeBytesAvailable
	return
}

// 检测进程是否存在
func IsProcessExist(appName string) bool {
	buf := bytes.Buffer{}
	cmd := exec.Command("wmic", "process", "get", "name,executablepath")
	cmd.Stdout = &buf
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()

	cmd2 := exec.Command("findstr", appName)
	cmd2.Stdin = &buf
	data, _ := cmd2.CombinedOutput()
	if len(data) == 0 {
		return false
	}
	return true
}

func StartProcess(appPath string) bool {
	cmd := exec.Command(appPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err := cmd.Output()
	if err != nil {
		log.Info("StartProcess", appPath, " err:", err.Error())
		return false
	}
	return true
}

// 结束进程
func KillProcess(appName string) {
	exec.Command(`taskkill`, `/f`, `/t`, `/im`, appName).Run()
}

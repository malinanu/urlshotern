// +build windows

package services

import (
	"syscall"
	"unsafe"
)

// getDiskUsagePlatform retrieves disk usage statistics on Windows
func getDiskUsagePlatform() (used int64, total int64) {
	// Get current directory's drive
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return 0, 0
	}
	defer kernel32.Release()

	getDiskFreeSpaceEx, err := kernel32.FindProc("GetDiskFreeSpaceExW")
	if err != nil {
		return 0, 0
	}

	// Get current working directory to determine drive
	cwd := make([]uint16, 260)
	n := uint32(len(cwd))
	getVolumePathName, err := kernel32.FindProc("GetVolumePathNameW")
	if err != nil {
		return 0, 0
	}

	cwdPath, err := syscall.UTF16PtrFromString(".")
	if err != nil {
		return 0, 0
	}

	ret, _, _ := getVolumePathName.Call(
		uintptr(unsafe.Pointer(cwdPath)),
		uintptr(unsafe.Pointer(&cwd[0])),
		uintptr(n),
	)

	if ret == 0 {
		return 0, 0
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes int64

	ret, _, _ = getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(&cwd[0])),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return 0, 0
	}

	// Convert bytes to MB
	total = totalBytes / 1024 / 1024
	used = (totalBytes - totalFreeBytes) / 1024 / 1024

	return used, total
}

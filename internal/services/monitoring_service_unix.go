// +build !windows

package services

import (
	"os"
	"syscall"
)

// getDiskUsagePlatform retrieves disk usage statistics on Unix-like systems
func getDiskUsagePlatform() (used int64, total int64) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return 0, 0
	}

	// Get filesystem stats
	var stat syscall.Statfs_t
	err = syscall.Statfs(wd, &stat)
	if err != nil {
		return 0, 0
	}

	// Calculate total and free space in MB
	total = int64(stat.Blocks * uint64(stat.Bsize)) / 1024 / 1024
	free := int64(stat.Bfree * uint64(stat.Bsize)) / 1024 / 1024
	used = total - free

	return used, total
}

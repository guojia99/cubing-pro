//go:build unix && ((!aix && !solaris) || illumos)

package database

import "syscall"

func flockSQLite(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_EX)
}

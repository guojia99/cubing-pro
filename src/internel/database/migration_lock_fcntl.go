//go:build aix || (solaris && !illumos)

package database

import "syscall"

func flockSQLite(fd uintptr) error {
	return syscall.FcntlFlock(fd, syscall.F_SETLKW, &syscall.Flock_t{Type: syscall.F_WRLCK})
}

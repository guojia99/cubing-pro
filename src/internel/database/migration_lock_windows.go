package database

import "golang.org/x/sys/windows"

func lockSQLiteHandle(fd uintptr) error {
	return windows.LockFileEx(windows.Handle(fd), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, new(windows.Overlapped))
}

//go:build unix

package database

import (
	"os"
	"syscall"
)

func lockSQLiteFile(file *os.File) error {
	for {
		if err := flockSQLite(file.Fd()); err != syscall.EINTR {
			return err
		}
	}
}

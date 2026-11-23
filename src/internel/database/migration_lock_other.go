//go:build !unix

package database

import "os"

func lockSQLiteFile(file *os.File) error {
	return lockSQLiteHandle(file.Fd())
}

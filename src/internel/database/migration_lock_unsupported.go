//go:build !unix && !windows

package database

import "errors"

func lockSQLiteHandle(uintptr) error {
	return errors.ErrUnsupported
}

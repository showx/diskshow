//go:build windows

package scan

import (
	"os"
	"syscall"
)

func isReparsePoint(path string, de os.DirEntry) bool {
	info, err := de.Info()
	if err != nil {
		info, err = os.Lstat(path)
		if err != nil {
			return false
		}
	}
	stat, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return false
	}
	return stat.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
}

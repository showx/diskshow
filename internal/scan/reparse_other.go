//go:build !windows

package scan

import "os"

func isReparsePoint(path string, de os.DirEntry) bool {
	_ = path
	_ = de
	return false
}

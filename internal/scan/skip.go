package scan

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var defaultSkipNames = map[string]struct{}{
	"System Volume Information": {},
	"$Recycle.Bin":              {},
	"$RECYCLE.BIN":              {},
	"Recovery":                  {},
	"pagefile.sys":              {},
	"hiberfil.sys":              {},
	"swapfile.sys":              {},
	"DumpStack.log.tmp":         {},
}

func shouldSkip(path string, de os.DirEntry, skipSpecial bool) bool {
	name := de.Name()
	if skipSpecial {
		if _, ok := defaultSkipNames[name]; ok {
			return true
		}
	}
	if de.Type()&os.ModeSymlink != 0 {
		return true
	}
	if runtime.GOOS == "windows" && isReparsePoint(path, de) {
		return true
	}
	return false
}

func isHiddenName(name string) bool {
	if name == "." || name == ".." {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	return false
}

func join(parent, name string) string {
	if parent == "" {
		return name
	}
	return filepath.Join(parent, name)
}

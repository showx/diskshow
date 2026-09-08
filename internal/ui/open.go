package ui

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

func openInExplorer(path string) {
	if path == "" {
		return
	}
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("explorer", "/select,", path).Start()
	case "darwin":
		_ = exec.Command("open", "-R", path).Start()
	default:
		_ = exec.Command("xdg-open", filepath.Dir(path)).Start()
	}
}

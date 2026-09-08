package ui

import (
	"hash/fnv"
	"path/filepath"
	"strings"

	"diskshow/internal/scan"

	"github.com/charmbracelet/lipgloss"
)

var folderPalette = []string{
	"#2b4c7e", "#1e5c4a", "#6b3f1a", "#4a3d6b",
	"#1a5c5c", "#6b2f4a", "#3d5a1e", "#5c2f2f",
	"#2f4a6b", "#3d4a2f", "#5c3d1a", "#3d2f5c",
}

var (
	colorBg       = "#0e1116"
	colorHeaderBg = "#151a22"
	colorFooterBg = "#151a22"
	colorFg       = "#e6edf3"
	colorMuted    = "#8b9bb4"
	colorAccent   = "#7ee787"
	colorWarn     = "#f0883e"
	colorSelect   = "#ffd166"
	colorSelectFg = "#0e1116"
	colorOther    = "#2a2f3a"
	colorBorder   = "#9fb0c8"
)

func tileColors(n *scan.Node) (bg, fg, border string) {
	fg = "#f4f7fb"
	border = "#c5d0e0"
	if n == nil {
		return "#2a2f3a", fg, border
	}
	if n.Virtual {
		return colorOther, "#c8d0dc", "#7a8494"
	}
	if n.IsDir {
		bg = folderPalette[hashStr(n.Path)%uint32(len(folderPalette))]
		return bg, fg, lighten(bg)
	}
	return extColor(n.Ext()), fg, "#d0d8e4"
}

func extColor(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".m2ts":
		return "#8a2f3a"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a", ".wma":
		return "#5a3d8a"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".psd", ".ico", ".heic":
		return "#1e6b45"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".rtf":
		return "#2f4f8a"
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".iso":
		return "#8a5a18"
	case ".go", ".js", ".ts", ".tsx", ".jsx", ".py", ".rs", ".java", ".c", ".cpp", ".h", ".cs", ".php", ".rb", ".kt", ".swift":
		return "#186868"
	case ".exe", ".dll", ".so", ".dylib", ".msi", ".apk", ".app":
		return "#3a3f4d"
	case ".db", ".sqlite", ".mdb", ".sql":
		return "#3d5a6b"
	default:
		return "#3a4550"
	}
}

func hashStr(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func lighten(hex string) string {
	if len(hex) != 7 || hex[0] != '#' {
		return "#dde5f0"
	}
	return hex
}

func extLabel(n *scan.Node) string {
	if n == nil {
		return ""
	}
	if n.Virtual {
		return "其他"
	}
	if n.IsDir {
		return "目录"
	}
	ext := strings.ToLower(filepath.Ext(n.Name))
	if ext == "" {
		return "文件"
	}
	return strings.TrimPrefix(ext, ".")
}

func headerStyle(w int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorFg)).
		Background(lipgloss.Color(colorHeaderBg)).
		Width(w).
		Bold(true)
}

func footerStyle(w int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Background(lipgloss.Color(colorFooterBg)).
		Width(w)
}

func accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
}

func warn() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarn))
}

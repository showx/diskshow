package ui

import (
	"fmt"
	"strings"

	"diskshow/internal/scan"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *Model) innerHeight() int {
	h := m.height - 3
	if h < 1 {
		return 1
	}
	return h
}

func (m *Model) listPanelHeight() int {
	inner := m.innerHeight()
	if m.mode == modeList {
		return inner
	}
	list := 10
	if inner < 18 {
		list = max(5, inner/2)
	}
	if list > inner-5 {
		list = max(4, inner-5)
	}
	return list
}

func (m *Model) treemapHeight() int {
	if m.mode == modeList {
		return 0
	}
	h := m.innerHeight() - m.listPanelHeight()
	if h < 3 {
		return 3
	}
	return h
}

func (m *Model) listPageSize() int {
	h := m.listPanelHeight() - 1
	if h < 1 {
		return 1
	}
	return h
}

func (m *Model) maxListOffset() int {
	n := len(sortedChildren(m.current))
	maxOff := n - m.listPageSize()
	if maxOff < 0 {
		return 0
	}
	return maxOff
}

func (m *Model) scrollList(delta int) {
	kids := sortedChildren(m.current)
	if len(kids) == 0 {
		return
	}
	m.listOffset = clamp(m.listOffset+delta, 0, m.maxListOffset())
	page := m.listPageSize()
	end := m.listOffset + page
	if end > len(kids) {
		end = len(kids)
	}
	if m.listIndex < m.listOffset {
		m.listIndex = m.listOffset
		m.selected = kids[m.listIndex]
	} else if m.listIndex >= end {
		m.listIndex = end - 1
		m.selected = kids[m.listIndex]
	}
}

func (m *Model) ensureListVisible() {
	kids := sortedChildren(m.current)
	page := m.listPageSize()
	if page < 1 {
		page = 1
	}
	if m.listIndex < m.listOffset {
		m.listOffset = m.listIndex
	}
	if m.listIndex >= m.listOffset+page {
		m.listOffset = m.listIndex - page + 1
	}
	if m.listOffset < 0 {
		m.listOffset = 0
	}
	if maxOff := len(kids) - page; maxOff >= 0 && m.listOffset > maxOff {
		m.listOffset = maxOff
	}
}

func (m *Model) renderChildList(w, h int) string {
	kids := sortedChildren(m.current)
	if h < 1 || w < 1 {
		return ""
	}
	page := h - 1
	if page < 1 {
		page = 1
	}
	if m.listOffset > m.maxListOffset() {
		m.listOffset = m.maxListOffset()
	}

	total := int64(0)
	if m.current != nil {
		total = m.current.Size()
	}
	end := m.listOffset + page
	if end > len(kids) {
		end = len(kids)
	}

	shown := fmt.Sprintf("%d–%d / %d", 0, 0, len(kids))
	if len(kids) > 0 {
		from := m.listOffset + 1
		to := end
		shown = fmt.Sprintf("%d–%d / %d", from, to, len(kids))
	}
	more := ""
	if m.listOffset+page < len(kids) {
		more = "  ↓ 滚动查看更多"
	} else if m.listOffset > 0 {
		more = "  ↑ 上方还有"
	}
	title := fmt.Sprintf(" 本目录子项  %s%s", shown, more)
	head := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Background(lipgloss.Color(colorHeaderBg)).
		Width(w).
		Render(truncate(title, w))

	rows := make([]string, 0, page)
	nameW := max(10, w-42)
	barW := max(4, min(14, w-58))
	for i := m.listOffset; i < end; i++ {
		rows = append(rows, m.formatListRow(kids[i], i, w, nameW, barW, total, page, len(kids)))
	}
	for len(rows) < page {
		rows = append(rows, lipgloss.NewStyle().
			Width(w).
			Background(lipgloss.Color(colorBg)).
			Render(""))
	}
	return head + "\n" + strings.Join(rows, "\n")
}

func (m *Model) formatListRow(n *scan.Node, idx, w, nameW, barW int, total int64, page, totalN int) string {
	sel := n == m.selected || ancestorChild(m.current, m.selected) == n
	kind := "文件"
	if n.IsDir {
		kind = "目录"
	}
	if n.Scanning() {
		kind = "扫描"
	}
	sbChar := scrollbarGlyph(idx-m.listOffset, page, m.listOffset, totalN)
	line := fmt.Sprintf(" %s %s %s %s %s",
		pad(kind, 4),
		pad(truncate(n.Name, nameW), nameW),
		pad(formatBytes(n.Size()), 10),
		pad(formatPct(n.Size(), total), 7),
		bar(n.Size(), total, barW),
	)
	line = truncate(line, max(1, w-1))
	line = pad(line, max(1, w-1)) + sbChar

	st := lipgloss.NewStyle().Width(w).Background(lipgloss.Color(colorBg)).Foreground(lipgloss.Color(colorFg))
	if sel {
		st = st.Background(lipgloss.Color(colorSelect)).Foreground(lipgloss.Color(colorSelectFg)).Bold(true)
	} else if n.IsDir {
		st = st.Foreground(lipgloss.Color("#b8d4ff"))
	}
	return st.Render(line)
}

func scrollbarGlyph(rowInPage, page, offset, total int) string {
	if total <= page || page <= 0 {
		return " "
	}
	thumbH := page * page / total
	if thumbH < 1 {
		thumbH = 1
	}
	maxOff := total - page
	thumbY := 0
	if maxOff > 0 {
		thumbY = offset * (page - thumbH) / maxOff
	}
	if rowInPage >= thumbY && rowInPage < thumbY+thumbH {
		return "█"
	}
	return "│"
}

func fitHeight(s string, h, w int) string {
	if h < 1 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	blank := strings.Repeat(" ", max(0, w))
	bg := lipgloss.NewStyle().Background(lipgloss.Color(colorBg)).Width(w)
	for i := range lines {
		if runewidth.StringWidth(lines[i]) == 0 {
			lines[i] = bg.Render(blank)
		}
	}
	for len(lines) < h {
		lines = append(lines, bg.Render(blank))
	}
	return strings.Join(lines, "\n")
}

package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (m *Model) View() string {
	if !m.ready {
		return "正在启动…"
	}
	header := m.renderHeader()
	body := m.renderBody()
	status := m.renderStatus()
	help := m.renderHelpLine()
	view := strings.Join([]string{header, body, status, help}, "\n")
	if m.showHelp {
		return overlayHelp(view, m.width, m.height)
	}
	return view
}

func (m *Model) renderHeader() string {
	w := m.width
	spin := "✓"
	state := "完成"
	if !m.stats.Done {
		spin = spinFrames[m.frame%len(spinFrames)]
		state = "扫描中"
	}
	left := fmt.Sprintf(" diskshow  %s %s  %s", spin, state, m.breadcrumb(max(10, w/2)))
	right := fmt.Sprintf("%s   %s 文件  %s 目录",
		formatBytes(m.stats.Bytes),
		commaInt(m.stats.Files),
		commaInt(m.stats.Dirs),
	)
	if m.stats.Errors > 0 {
		right += fmt.Sprintf("  %s 错误", commaInt(m.stats.Errors))
	}
	if !m.stats.Done && !m.stats.StartedAt.IsZero() {
		right += "  " + formatDuration(time.Since(m.stats.StartedAt))
	} else if m.stats.Done && !m.stats.FinishedAt.IsZero() && !m.stats.StartedAt.IsZero() {
		right += "  " + formatDuration(m.stats.FinishedAt.Sub(m.stats.StartedAt))
	}
	gap := w - runewidth.StringWidth(left) - runewidth.StringWidth(right) - 1
	if gap < 1 {
		left = truncate(left, max(8, w-runewidth.StringWidth(right)-2))
		gap = w - runewidth.StringWidth(left) - runewidth.StringWidth(right) - 1
		if gap < 1 {
			gap = 1
		}
	}
	line := left + strings.Repeat(" ", gap) + right + " "
	return headerStyle(w).Render(line)
}

func (m *Model) renderBody() string {
	w, h := m.width, m.mapHeight()
	if m.current != nil && m.current.ChildCount() == 0 {
		msg := "空目录"
		if m.current.Scanning() {
			msg = "正在扫描子项…"
		}
		if m.current.Err != nil {
			msg = "无法读取: " + m.current.Err.Error()
		}
		style := lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color(colorMuted)).
			Background(lipgloss.Color(colorBg))
		return style.Render(msg)
	}
	if m.mode == modeList {
		return m.renderList(w, h)
	}
	if len(m.tiles) == 0 {
		style := lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color(colorMuted)).
			Background(lipgloss.Color(colorBg))
		msg := "正在计算布局…"
		if m.current != nil && m.current.Scanning() {
			msg = "正在扫描，矩形会随结果实时长大"
		}
		return style.Render(msg)
	}
	return renderTreemap(m.tiles, m.selected, w, h)
}

func (m *Model) renderList(w, h int) string {
	kids := sortedChildren(m.current)
	var b strings.Builder
	cols := fmt.Sprintf(" %s %s %s %s %s",
		pad("类型", 6),
		pad("名称", max(12, w-44)),
		pad("大小", 10),
		pad("占比", 7),
		"占用",
	)
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted)).
		Background(lipgloss.Color(colorBg)).
		Width(w).
		Render(" " + truncate(cols, w-1)))
	b.WriteByte('\n')

	nameW := max(12, w-44)
	barW := max(6, min(18, w-62))
	page := h - 1
	if page < 1 {
		page = 1
	}
	if m.listOffset > 0 && m.listOffset >= len(kids) {
		m.listOffset = 0
	}
	end := m.listOffset + page
	if end > len(kids) {
		end = len(kids)
	}
	total := int64(0)
	if m.current != nil {
		total = m.current.Size()
	}
	for i := m.listOffset; i < end; i++ {
		n := kids[i]
		sel := n == m.selected
		kind := "文件"
		if n.IsDir {
			kind = "目录"
		}
		if n.Scanning() {
			kind = "扫描"
		}
		line := fmt.Sprintf(" %s %s %s %s %s",
			pad(kind, 6),
			pad(truncate(n.Name, nameW), nameW),
			pad(formatBytes(n.Size()), 10),
			pad(formatPct(n.Size(), total), 7),
			bar(n.Size(), total, barW),
		)
		st := lipgloss.NewStyle().Width(w).Background(lipgloss.Color(colorBg)).Foreground(lipgloss.Color(colorFg))
		if sel {
			st = st.Background(lipgloss.Color(colorSelect)).Foreground(lipgloss.Color(colorSelectFg)).Bold(true)
		} else if n.IsDir {
			st = st.Foreground(lipgloss.Color("#b8d4ff"))
		}
		b.WriteString(st.Render(truncate(line, w)))
		if i < end-1 {
			b.WriteByte('\n')
		}
	}
	used := end - m.listOffset
	if used < 1 {
		used = 1
		empty := lipgloss.NewStyle().Width(w).Height(page).Background(lipgloss.Color(colorBg)).Foreground(lipgloss.Color(colorMuted)).Align(lipgloss.Center, lipgloss.Center)
		return empty.Render("没有可显示的子项")
	}
	if used < page {
		padH := page - used
		b.WriteByte('\n')
		b.WriteString(lipgloss.NewStyle().Width(w).Height(padH).Background(lipgloss.Color(colorBg)).Render(""))
	}
	return b.String()
}

func (m *Model) renderStatus() string {
	w := m.width
	n := m.selected
	if n == nil {
		n = m.current
	}
	left := " 未选中"
	if n != nil {
		kind := "文件"
		if n.IsDir {
			kind = "目录"
		}
		if n.Virtual {
			kind = "汇总"
		}
		parentSize := int64(0)
		if m.current != nil {
			parentSize = m.current.Size()
		}
		path := n.Path
		if n.Virtual {
			path = n.Name
		}
		left = fmt.Sprintf(" %s  %s  %s  %s",
			kind,
			truncate(path, max(10, w/2)),
			formatBytes(n.Size()),
			formatPct(n.Size(), parentSize),
		)
		if n.IsDir {
			left += fmt.Sprintf("  %s文件/%s目录", commaInt(n.Files()), commaInt(n.Dirs()))
		}
	}
	right := ""
	if m.statusNote != "" {
		right = m.statusNote + "  "
	} else if !m.stats.Done && m.stats.Current != "" {
		right = "正在扫描 " + truncate(m.stats.Current, max(12, w/3)) + "  "
	}
	gap := w - runewidth.StringWidth(left) - runewidth.StringWidth(right)
	if gap < 1 {
		left = truncate(left, w-runewidth.StringWidth(right)-1)
		gap = w - runewidth.StringWidth(left) - runewidth.StringWidth(right)
		if gap < 1 {
			gap = 1
		}
	}
	line := left + strings.Repeat(" ", gap) + right
	return footerStyle(w).Foreground(lipgloss.Color(colorFg)).Render(line)
}

func (m *Model) renderHelpLine() string {
	mode := "树图"
	if m.mode == modeList {
		mode = "列表"
	}
	s := fmt.Sprintf(" 视图:%s   点击选中  双击/Enter进入  右键/Backspace返回  Tab切换视图  o资源管理器  ?帮助  q退出", mode)
	return footerStyle(m.width).Render(truncate(s, m.width))
}

func overlayHelp(base string, w, h int) string {
	lines := []string{
		" SpaceSniffer 风格磁盘分析",
		"",
		" 鼠标左键          选中矩形（点到最内层文件/目录）",
		" 双击 / Enter      进入目录，查看其子项占用",
		" 鼠标右键 / 退格   返回上一级",
		" Tab               树图 ↔ 列表（按大小排序）",
		" ↑ ↓ ← → / hjkl    在同级项目间移动",
		" o                 在资源管理器中定位选中项",
		" ? / Esc           关闭本帮助",
		" q                 退出",
		"",
		" 矩形面积 = 占用空间。目录内部会继续嵌套子项，",
		" 无需进入也能看到里面的大文件。",
	}
	maxW := 0
	for _, ln := range lines {
		if runewidth.StringWidth(ln) > maxW {
			maxW = runewidth.StringWidth(ln)
		}
	}
	inner := maxW + 4
	if inner > w-4 {
		inner = w - 4
	}
	if inner < 20 {
		inner = max(10, w-2)
	}
	var body strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorAccent)).Render(" 快捷键 ")
	top := "┌" + "─" + title + strings.Repeat("─", max(0, inner-2-runewidth.StringWidth(" 快捷键 "))) + "┐"
	body.WriteString(top)
	body.WriteByte('\n')
	for _, ln := range lines {
		body.WriteString("│ " + pad(truncate(ln, inner-3), inner-3) + "│\n")
	}
	body.WriteString("└" + strings.Repeat("─", inner-1) + "┘")
	box := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorFg)).
		Background(lipgloss.Color("#1b2230")).
		Render(strings.TrimRight(body.String(), "\n"))
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(colorBg)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

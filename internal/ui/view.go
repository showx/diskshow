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
	w, inner := m.width, m.innerHeight()
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
			Height(inner).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color(colorMuted)).
			Background(lipgloss.Color(colorBg))
		return style.Render(msg)
	}
	if m.mode == modeList {
		return fitHeight(m.renderChildList(w, inner), inner, w)
	}

	mapH := m.treemapHeight()
	listH := inner - mapH
	var mapView string
	if len(m.tiles) == 0 {
		style := lipgloss.NewStyle().
			Width(w).
			Height(mapH).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color(colorMuted)).
			Background(lipgloss.Color(colorBg))
		msg := "正在计算布局…"
		if m.current != nil && m.current.Scanning() {
			msg = "正在扫描，下方列表可滚动查看全部子项"
		}
		mapView = style.Render(msg)
	} else {
		mapView = renderTreemap(m.tiles, m.selected, w, mapH)
	}
	listView := m.renderChildList(w, listH)
	return fitHeight(mapView, mapH, w) + "\n" + fitHeight(listView, listH, w)
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
	s := " 滚轮/↓↑滚动子项列表  PgDn/PgUp翻页  点击选中  双击/Enter进入  Backspace返回  Tab全屏列表  ?帮助  q退出"
	return footerStyle(m.width).Render(truncate(s, m.width))
}

func overlayHelp(base string, w, h int) string {
	lines := []string{
		" SpaceSniffer 风格磁盘分析",
		"",
		" 鼠标滚轮 / ↓↑     滚动下方子项列表，查看一屏装不下的目录",
		" PgDn / PgUp       列表翻页",
		" 鼠标左键          选中矩形或列表行",
		" 双击 / Enter      进入目录，查看其子项占用",
		" 鼠标右键 / 退格   返回上一级",
		" Tab               树图+列表 ↔ 全屏列表",
		" ↑ ↓ / j k         在同级项目间移动（列表会跟着滚）",
		" o                 在资源管理器中定位选中项",
		" ? / Esc           关闭本帮助",
		" q                 退出",
		"",
		" 上方矩形面积 = 占用空间；下方列表可滚动浏览全部子项。",
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

package ui

import (
	"context"
	"strings"
	"time"

	"diskshow/internal/scan"

	tea "github.com/charmbracelet/bubbletea"
)

type viewMode int

const (
	modeTreemap viewMode = iota
	modeList
)

type tickMsg time.Time

type Model struct {
	scanner *scan.Scanner
	cancel  context.CancelFunc
	ctx     context.Context

	current  *scan.Node
	selected *scan.Node
	tiles    []Tile

	mode       viewMode
	listOffset int
	listIndex  int
	showHelp   bool

	width  int
	height int
	ready  bool

	stats scan.Stats
	frame int

	lastClickAt   time.Time
	lastClickNode *scan.Node

	statusNote string
}

func New(path string, opt scan.Options) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	sc := scan.NewScanner(path, opt)
	return &Model{
		scanner:  sc,
		cancel:   cancel,
		ctx:      ctx,
		current:  sc.Root(),
		selected: sc.Root(),
	}
}

func (m *Model) Init() tea.Cmd {
	m.scanner.Start(m.ctx)
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.relayout()
		return m, nil

	case tickMsg:
		m.stats = m.scanner.Snapshot()
		m.frame++
		m.relayout()
		if m.stats.Done {
			return m, nil
		}
		return m, tickCmd()

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		switch msg.String() {
		case "q", "esc", "?", "enter", " ":
			m.showHelp = false
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		if m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
	case "?":
		m.showHelp = true
	case "tab":
		if m.mode == modeTreemap {
			m.mode = modeList
			m.syncListIndex()
		} else {
			m.mode = modeTreemap
			m.relayout()
		}
	case "enter":
		m.enterSelected()
	case "backspace", "esc":
		m.goUp()
	case "o":
		if m.selected != nil && !m.selected.Virtual {
			openInExplorer(m.selected.Path)
			m.statusNote = "已在资源管理器中定位"
		}
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(1)
	case "left", "h":
		m.goUp()
	case "right", "l":
		m.enterSelected()
	case "pgup", "ctrl+u":
		m.scrollList(-m.listPageSize())
	case "pgdown", "ctrl+d":
		m.scrollList(m.listPageSize())
	case "home":
		m.selectIndex(0)
	case "end":
		kids := sortedChildren(m.current)
		m.selectIndex(len(kids) - 1)
	}
	return m, nil
}

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if !m.ready {
		return m, nil
	}

	if msg.Button == tea.MouseButtonWheelUp {
		m.scrollList(-3)
		return m, nil
	}
	if msg.Button == tea.MouseButtonWheelDown {
		m.scrollList(3)
		return m, nil
	}

	if msg.Button == tea.MouseButtonRight && msg.Action == tea.MouseActionPress {
		m.goUp()
		return m, nil
	}

	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
		return m, nil
	}

	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	listTop := 1
	if m.mode != modeList {
		listTop = 1 + m.treemapHeight()
	}
	listBottom := listTop + m.listPanelHeight()
	if msg.Y >= listTop && msg.Y < listBottom {
		m.clickList(msg.X, msg.Y-listTop)
		return m, nil
	}

	if m.mode == modeList {
		return m, nil
	}
	relY := msg.Y - 1
	if relY < 0 || relY >= m.treemapHeight() {
		return m, nil
	}
	t := hitTile(m.tiles, msg.X, relY)
	if t == nil || t.Node == nil {
		return m, nil
	}
	m.clickNode(t.Node)
	return m, nil
}

func (m *Model) clickList(x, relY int) {
	kids := sortedChildren(m.current)
	page := m.listPageSize()
	if relY <= 0 {
		return
	}
	if x >= m.width-1 && len(kids) > page {
		ratio := float64(relY-1) / float64(page)
		m.listOffset = clamp(int(ratio*float64(m.maxListOffset()+1)), 0, m.maxListOffset())
		if m.listIndex < m.listOffset || m.listIndex >= m.listOffset+page {
			idx := clamp(m.listOffset, 0, len(kids)-1)
			m.listIndex = idx
			m.selected = kids[idx]
		}
		return
	}
	row := relY - 1 + m.listOffset
	if row >= 0 && row < len(kids) {
		m.clickNode(kids[row])
	}
}

func (m *Model) clickNode(n *scan.Node) {
	now := time.Now()
	double := m.lastClickNode == n && now.Sub(m.lastClickAt) < 400*time.Millisecond
	m.lastClickAt = now
	m.lastClickNode = n
	m.selected = n
	m.syncListIndex()
	if double {
		m.enterSelected()
	}
}

func (m *Model) enterSelected() {
	n := m.selected
	if n == nil {
		return
	}
	if n.Virtual {
		m.mode = modeList
		m.syncListIndex()
		m.statusNote = "较小项目已在列表中展开，按 Tab 可回到树图"
		return
	}
	if n.IsDir {
		m.current = n
		kids := sortedChildren(n)
		if len(kids) > 0 {
			m.selected = kids[0]
		} else {
			m.selected = n
		}
		m.listOffset = 0
		m.listIndex = 0
		m.relayout()
		m.statusNote = ""
	}
}

func (m *Model) goUp() {
	if m.current == nil || m.current.Parent == nil {
		m.statusNote = "已经在扫描根目录"
		return
	}
	prev := m.current
	m.current = m.current.Parent
	m.selected = prev
	m.syncListIndex()
	m.relayout()
	m.statusNote = ""
}

func (m *Model) moveSelection(delta int) {
	kids := sortedChildren(m.current)
	if len(kids) == 0 {
		return
	}
	idx := 0
	for i, c := range kids {
		if c == m.selected || (m.selected != nil && !m.selected.Virtual && ancestorChild(m.current, m.selected) == c) {
			idx = i
			break
		}
	}
	m.selectIndex(idx + delta)
}

func (m *Model) selectIndex(idx int) {
	kids := sortedChildren(m.current)
	if len(kids) == 0 {
		return
	}
	idx = clamp(idx, 0, len(kids)-1)
	m.selected = kids[idx]
	m.listIndex = idx
	m.ensureListVisible()
}

func (m *Model) syncListIndex() {
	kids := sortedChildren(m.current)
	target := m.selected
	if ac := ancestorChild(m.current, m.selected); ac != nil {
		target = ac
	}
	m.listIndex = 0
	for i, c := range kids {
		if c == target {
			m.listIndex = i
			break
		}
	}
	m.ensureListVisible()
}

func ancestorChild(current, target *scan.Node) *scan.Node {
	if current == nil || target == nil {
		return nil
	}
	for p := target; p != nil && p.Parent != nil; p = p.Parent {
		if p.Parent == current {
			return p
		}
	}
	return nil
}

func (m *Model) relayout() {
	if !m.ready || m.current == nil {
		return
	}
	w, h := m.width, m.treemapHeight()
	if w >= 1 && h >= 1 {
		m.tiles = buildTiles(m.current, 0, 0, w, h, 0)
	} else {
		m.tiles = nil
	}
	if m.selected != nil && m.selected.Virtual {
		for i := range m.tiles {
			if m.tiles[i].Depth == 0 && m.tiles[i].Node != nil && m.tiles[i].Node.Virtual {
				m.selected = m.tiles[i].Node
				return
			}
		}
	}
	if m.selected == nil || (m.selected != m.current && !containsNode(m.current, m.selected) && !m.selected.Virtual) {
		kids := sortedChildren(m.current)
		if len(kids) > 0 {
			m.selected = kids[0]
		} else {
			m.selected = m.current
		}
	}
}

func (m *Model) breadcrumb(maxW int) string {
	var parts []string
	for p := m.current; p != nil; p = p.Parent {
		parts = append([]string{p.Name}, parts...)
	}
	s := strings.Join(parts, " › ")
	return truncate(s, maxW)
}

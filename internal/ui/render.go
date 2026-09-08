package ui

import (
	"strings"

	"diskshow/internal/scan"

	"github.com/mattn/go-runewidth"
)

type cell struct {
	ch   rune
	fg   string
	bg   string
	bold bool
	skip bool
}

type grid struct {
	w, h  int
	cells []cell
}

func newGrid(w, h int, bg string) *grid {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	g := &grid{w: w, h: h, cells: make([]cell, w*h)}
	for i := range g.cells {
		g.cells[i] = cell{ch: ' ', bg: bg, fg: colorFg}
	}
	return g
}

func (g *grid) at(x, y int) *cell {
	if x < 0 || y < 0 || x >= g.w || y >= g.h {
		return nil
	}
	return &g.cells[y*g.w+x]
}

func (g *grid) fill(x, y, w, h int, bg, fg string) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			c := g.at(xx, yy)
			if c == nil {
				continue
			}
			c.ch = ' '
			c.bg = bg
			c.fg = fg
			c.bold = false
			c.skip = false
		}
	}
}

func (g *grid) put(x, y int, ch rune, fg, bg string, bold bool) {
	c := g.at(x, y)
	if c == nil {
		return
	}
	c.ch = ch
	c.fg = fg
	c.bg = bg
	c.bold = bold
	c.skip = false
	rw := runewidth.RuneWidth(ch)
	if rw > 1 {
		n := g.at(x+1, y)
		if n != nil {
			n.skip = true
			n.ch = 0
			n.bg = bg
			n.fg = fg
		}
	}
}

func (g *grid) print(x, y, maxW int, s, fg, bg string, bold bool) {
	if maxW <= 0 {
		return
	}
	cx := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if rw <= 0 {
			continue
		}
		if cx+rw > maxW {
			break
		}
		g.put(x+cx, y, r, fg, bg, bold)
		cx += rw
	}
}

func (g *grid) box(x, y, w, h int, fg, bg string, selected bool) {
	if w < 1 || h < 1 {
		return
	}
	if selected {
		fg = colorSelect
	}
	if w == 1 && h == 1 {
		g.put(x, y, '■', fg, bg, selected)
		return
	}
	var tl, tr, bl, br, hch, vch rune = '┌', '┐', '└', '┘', '─', '│'
	if selected {
		tl, tr, bl, br, hch, vch = '╔', '╗', '╚', '╝', '═', '║'
	}
	if h == 1 {
		g.fill(x, y, w, 1, bg, fg)
		return
	}
	if w == 1 {
		for yy := y; yy < y+h; yy++ {
			g.put(x, yy, vch, fg, bg, selected)
		}
		return
	}
	g.put(x, y, tl, fg, bg, selected)
	g.put(x+w-1, y, tr, fg, bg, selected)
	g.put(x, y+h-1, bl, fg, bg, selected)
	g.put(x+w-1, y+h-1, br, fg, bg, selected)
	for xx := x + 1; xx < x+w-1; xx++ {
		g.put(xx, y, hch, fg, bg, selected)
		g.put(xx, y+h-1, hch, fg, bg, selected)
	}
	for yy := y + 1; yy < y+h-1; yy++ {
		g.put(x, yy, vch, fg, bg, selected)
		g.put(x+w-1, yy, vch, fg, bg, selected)
	}
}

func (g *grid) render() string {
	var b strings.Builder
	b.Grow(g.w * g.h * 12)
	for y := 0; y < g.h; y++ {
		x := 0
		for x < g.w {
			c := g.at(x, y)
			if c.skip {
				x++
				continue
			}
			runEnd := x + 1
			for runEnd < g.w {
				n := g.at(runEnd, y)
				if n.skip {
					runEnd++
					continue
				}
				if n.fg != c.fg || n.bg != c.bg || n.bold != c.bold {
					break
				}
				runEnd++
			}
			var sb strings.Builder
			for i := x; i < runEnd; i++ {
				n := g.at(i, y)
				if n.skip || n.ch == 0 {
					continue
				}
				if n.ch == 0 {
					continue
				}
				sb.WriteRune(n.ch)
			}
			st := ansiStyle(c.fg, c.bg, c.bold)
			b.WriteString(st)
			b.WriteString(sb.String())
			b.WriteString("\x1b[0m")
			x = runEnd
		}
		if y < g.h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func ansiStyle(fg, bg string, bold bool) string {
	var b strings.Builder
	b.WriteString("\x1b[0m")
	if bold {
		b.WriteString("\x1b[1m")
	}
	if fr, fgok := hexRGB(fg); fgok {
		b.WriteString("\x1b[38;2;")
		b.WriteString(itoa(fr.r))
		b.WriteByte(';')
		b.WriteString(itoa(fr.g))
		b.WriteByte(';')
		b.WriteString(itoa(fr.b))
		b.WriteByte('m')
	}
	if br, bgok := hexRGB(bg); bgok {
		b.WriteString("\x1b[48;2;")
		b.WriteString(itoa(br.r))
		b.WriteByte(';')
		b.WriteString(itoa(br.g))
		b.WriteByte(';')
		b.WriteString(itoa(br.b))
		b.WriteByte('m')
	}
	return b.String()
}

type rgb struct{ r, g, b int }

func hexRGB(hex string) (rgb, bool) {
	if len(hex) != 7 || hex[0] != '#' {
		return rgb{}, false
	}
	ok := true
	n := func(i int) int {
		v, good := unhex(hex[i], hex[i+1])
		ok = ok && good
		return v
	}
	c := rgb{r: n(1), g: n(3), b: n(5)}
	return c, ok
}

func unhex(a, b byte) (int, bool) {
	hi, ok1 := hexVal(a)
	lo, ok2 := hexVal(b)
	return hi*16 + lo, ok1 && ok2
}

func hexVal(c byte) (int, bool) {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0'), true
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10, true
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10, true
	}
	return 0, false
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func renderTreemap(tiles []Tile, selected *scan.Node, w, h int) string {
	g := newGrid(w, h, colorBg)
	for i := range tiles {
		t := tiles[i]
		g.fill(t.X, t.Y, t.W, t.H, t.Bg, t.Fg)
	}
	for i := range tiles {
		t := tiles[i]
		sel := selected != nil && t.Node == selected
		g.box(t.X, t.Y, t.W, t.H, t.Border, t.Bg, sel)
		drawTileLabel(g, t, sel)
	}
	return g.render()
}

func drawTileLabel(g *grid, t Tile, selected bool) {
	if t.W < 1 || t.H < 1 || t.Node == nil {
		return
	}
	fg, bg := t.Fg, t.Bg
	bold := selected || t.Depth == 0
	name := t.Node.Name
	if t.Node.Scanning() {
		name = "…" + name
	}
	size := formatBytes(t.Node.Size())
	if t.W <= 2 || t.H == 1 {
		g.print(t.X, t.Y, t.W, firstRune(name), fg, bg, bold)
		return
	}

	labelY := t.Y
	labelX := t.X + 1
	labelW := t.W - 2
	if t.W < 3 {
		labelX = t.X
		labelW = t.W
	}
	if labelW <= 0 {
		return
	}

	if t.H == 2 {
		g.print(labelX, labelY, labelW, truncate(name+" "+size, labelW), fg, bg, bold)
		return
	}

	title := " " + truncate(name, max(1, labelW-1))
	g.print(labelX, labelY, labelW, title, fg, bg, true)
	if t.H >= 4 && labelW >= 4 {
		g.print(labelX, labelY+1, labelW, " "+size, fg, bg, false)
	} else if t.H >= 3 && labelW >= 4 && t.W < 8 {
		g.print(labelX, labelY+1, labelW, " "+size, fg, bg, false)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

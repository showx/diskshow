package ui

import (
	"fmt"
	"sort"

	"diskshow/internal/scan"
	"diskshow/internal/treemap"
)

const (
	maxTilesPerLevel = 56
	maxNestDepth     = 5
)

type Tile struct {
	Node       *scan.Node
	X, Y, W, H int
	Depth      int
	Bg, Fg     string
	Border     string
}

func pickChildren(n *scan.Node) []*scan.Node {
	if n == nil {
		return nil
	}
	all := n.Children()
	sort.Slice(all, func(i, j int) bool {
		si, sj := all[i].Size(), all[j].Size()
		if si == sj {
			return all[i].Name < all[j].Name
		}
		return si > sj
	})
	if len(all) <= maxTilesPerLevel {
		return all
	}

	head := all[:maxTilesPerLevel-1]
	rest := all[maxTilesPerLevel-1:]
	var restSize, restFiles, restDirs int64
	for _, c := range rest {
		restSize += c.Size()
		restFiles += c.Files()
		restDirs += c.Dirs()
		if !c.IsDir {
			restFiles++
		} else {
			restDirs++
		}
	}
	other := scan.NewVirtual(
		fmt.Sprintf("其他 %d 项", len(rest)),
		n.Path,
		n,
		restSize,
		restFiles,
		restDirs,
	)
	return append(head, other)
}

func sortedChildren(n *scan.Node) []*scan.Node {
	if n == nil {
		return nil
	}
	all := n.Children()
	sort.Slice(all, func(i, j int) bool {
		si, sj := all[i].Size(), all[j].Size()
		if si == sj {
			return all[i].Name < all[j].Name
		}
		return si > sj
	})
	return all
}

func sizesOf(nodes []*scan.Node) []float64 {
	s := make([]float64, len(nodes))
	nonzero := false
	for i, n := range nodes {
		s[i] = float64(n.Size())
		if s[i] > 0 {
			nonzero = true
		}
	}
	if !nonzero {
		for i := range s {
			s[i] = 1
		}
	}
	return s
}

func buildTiles(node *scan.Node, x, y, w, h, depth int) []Tile {
	if node == nil || w < 1 || h < 1 || depth > maxNestDepth {
		return nil
	}
	kids := pickChildren(node)
	if len(kids) == 0 {
		return nil
	}
	rects := treemap.Layout(sizesOf(kids), treemap.Rect{X: x, Y: y, W: w, H: h})
	var tiles []Tile
	for i, kid := range kids {
		r := clipRect(rects[i], x, y, w, h)
		if r.W < 1 || r.H < 1 {
			continue
		}
		bg, fg, border := tileColors(kid)
		t := Tile{
			Node:   kid,
			X:      r.X,
			Y:      r.Y,
			W:      r.W,
			H:      r.H,
			Depth:  depth,
			Bg:     bg,
			Fg:     fg,
			Border: border,
		}
		tiles = append(tiles, t)

		ix, iy, iw, ih := innerBounds(r.X, r.Y, r.W, r.H)
		if kid.IsDir && !kid.Virtual && depth < maxNestDepth && iw >= 8 && ih >= 3 {
			nested := buildTiles(kid, ix, iy, iw, ih, depth+1)
			tiles = append(tiles, nested...)
		}
	}
	return tiles
}

func innerBounds(x, y, w, h int) (ix, iy, iw, ih int) {
	if w >= 4 && h >= 5 {
		return x + 1, y + 2, w - 2, h - 3
	}
	if w >= 3 && h >= 3 {
		return x + 1, y + 1, w - 2, h - 2
	}
	return x, y, w, h
}

func clipRect(r treemap.Rect, x, y, w, h int) treemap.Rect {
	if r.X < x {
		r.W -= x - r.X
		r.X = x
	}
	if r.Y < y {
		r.H -= y - r.Y
		r.Y = y
	}
	if r.X+r.W > x+w {
		r.W = x + w - r.X
	}
	if r.Y+r.H > y+h {
		r.H = y + h - r.Y
	}
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return r
}

func hitTile(tiles []Tile, x, y int) *Tile {
	var found *Tile
	for i := range tiles {
		t := &tiles[i]
		if x >= t.X && x < t.X+t.W && y >= t.Y && y < t.Y+t.H {
			found = t
		}
	}
	return found
}

func containsNode(root, target *scan.Node) bool {
	if root == nil || target == nil {
		return false
	}
	for p := target; p != nil; p = p.Parent {
		if p == root {
			return true
		}
		if p.Virtual {
			return false
		}
	}
	return false
}

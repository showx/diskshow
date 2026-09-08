package treemap

import "math"

// Rect is a pixel/cell rectangle.
type Rect struct {
	X, Y, W, H int
}

type item struct {
	idx  int
	size float64
}

// Layout places items into bounds using the squarified treemap algorithm.
// sizes must be aligned with the caller’s item slice. Zero/negative sizes are ignored.
func Layout(sizes []float64, bounds Rect) []Rect {
	out := make([]Rect, len(sizes))
	if bounds.W <= 0 || bounds.H <= 0 || len(sizes) == 0 {
		return out
	}
	items := make([]item, 0, len(sizes))
	total := 0.0
	for i, s := range sizes {
		if s > 0 {
			items = append(items, item{idx: i, size: s})
			total += s
		}
	}
	if total <= 0 || len(items) == 0 {
		return out
	}
	area := float64(bounds.W * bounds.H)
	for i := range items {
		items[i].size = items[i].size / total * area
	}
	squarify(items, floatRect{
		x: float64(bounds.X),
		y: float64(bounds.Y),
		w: float64(bounds.W),
		h: float64(bounds.H),
	}, out)
	return out
}

type floatRect struct {
	x, y, w, h float64
}

func squarify(items []item, space floatRect, out []Rect) {
	if len(items) == 0 || space.w < 0.5 || space.h < 0.5 {
		return
	}
	if len(items) == 1 {
		assign(items[0], space, out)
		return
	}

	vertical := space.w >= space.h
	side := space.h
	if vertical {
		side = space.w
	}

	row := []item{items[0]}
	i := 1
	for i < len(items) {
		if worst(row, side) >= worst(append(row, items[i]), side) {
			row = append(row, items[i])
			i++
			continue
		}
		break
	}

	used := layoutRow(row, space, vertical, out)
	rest := items[i:]
	if len(rest) == 0 {
		return
	}
	if vertical {
		space.x += used
		space.w -= used
		if space.w < 0 {
			space.w = 0
		}
	} else {
		space.y += used
		space.h -= used
		if space.h < 0 {
			space.h = 0
		}
	}
	squarify(rest, space, out)
}

func layoutRow(row []item, space floatRect, vertical bool, out []Rect) float64 {
	sum := 0.0
	for _, it := range row {
		sum += it.size
	}
	if sum <= 0 {
		return 0
	}

	if vertical {
		w := sum / space.h
		if w > space.w {
			w = space.w
		}
		y := space.y
		for idx, it := range row {
			h := it.size / w
			if idx == len(row)-1 {
				h = space.y + space.h - y
			}
			assign(it, floatRect{x: space.x, y: y, w: w, h: h}, out)
			y += h
		}
		return w
	}

	h := sum / space.w
	if h > space.h {
		h = space.h
	}
	x := space.x
	for idx, it := range row {
		w := it.size / h
		if idx == len(row)-1 {
			w = space.x + space.w - x
		}
		assign(it, floatRect{x: space.x, y: space.y, w: w, h: h}, out)
		x += w
	}
	return h
}

func assign(it item, r floatRect, out []Rect) {
	x := int(math.Round(r.x))
	y := int(math.Round(r.y))
	x2 := int(math.Round(r.x + r.w))
	y2 := int(math.Round(r.y + r.h))
	w := x2 - x
	h := y2 - y
	if w < 1 && r.w > 0 {
		w = 1
	}
	if h < 1 && r.h > 0 {
		h = 1
	}
	out[it.idx] = Rect{X: x, Y: y, W: w, H: h}
}

func worst(row []item, side float64) float64 {
	if len(row) == 0 || side <= 0 {
		return math.Inf(1)
	}
	sum := 0.0
	minV := math.Inf(1)
	maxV := 0.0
	for _, it := range row {
		sum += it.size
		if it.size < minV {
			minV = it.size
		}
		if it.size > maxV {
			maxV = it.size
		}
	}
	if sum <= 0 || minV <= 0 {
		return math.Inf(1)
	}
	ss := sum * sum
	side2 := side * side
	a := side2 * maxV / ss
	b := ss / (side2 * minV)
	if a > b {
		return a
	}
	return b
}

package treemap

import "testing"

func TestLayoutFillsBounds(t *testing.T) {
	sizes := []float64{50, 30, 10, 10}
	bounds := Rect{X: 0, Y: 0, W: 40, H: 20}
	rects := Layout(sizes, bounds)

	area := 0
	for i, r := range rects {
		if r.W <= 0 || r.H <= 0 {
			t.Fatalf("item %d has empty rect: %+v", i, r)
		}
		if r.X < bounds.X || r.Y < bounds.Y {
			t.Fatalf("item %d outside bounds: %+v", i, r)
		}
		if r.X+r.W > bounds.X+bounds.W+1 || r.Y+r.H > bounds.Y+bounds.H+1 {
			t.Fatalf("item %d overflows: %+v vs %+v", i, r, bounds)
		}
		area += r.W * r.H
	}
	target := bounds.W * bounds.H
	if area < target*8/10 {
		t.Fatalf("coverage too low: %d / %d", area, target)
	}
}

func TestLayoutEmpty(t *testing.T) {
	rects := Layout(nil, Rect{W: 10, H: 10})
	if len(rects) != 0 {
		t.Fatalf("expected empty")
	}
}

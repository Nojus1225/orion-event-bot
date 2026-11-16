package board

import (
	"math"
	"orion-event-bot/bot/vars"
)

const (
	BoardWidth  = 2400
	BoardHeight = 1350
	TileRadius  = 38.0

	GondolaFromNumber = 6
	GondolaToNumber   = 14
)

type Point struct {
	X float64
	Y float64
}

type TileNode struct {
	Case vars.PlateCase
	Pos  Point
}

type BoardLayout struct {
	Tiles       []TileNode
	TrailPoints []Point
}

func BuildLayout() BoardLayout {
	waypoints := []Point{
		{X: 115, Y: 195},
		{X: 270, Y: 445},
		{X: 445, Y: 245},
		{X: 655, Y: 785},
		{X: 870, Y: 300},
		{X: 1130, Y: 1200},
		{X: 1425, Y: 345},
		{X: 1740, Y: 1230},
		{X: 2015, Y: 505},
		{X: 2265, Y: 1205},
	}

	const samplesPerSegment = 120
	segCurves := make([][]Point, len(waypoints)-1)
	segLengths := make([][]float64, len(waypoints)-1)

	var allTrail []Point
	for i := 0; i < len(waypoints)-1; i++ {
		p0 := waypoints[maxInt(0, i-1)]
		p1 := waypoints[i]
		p2 := waypoints[i+1]
		p3 := waypoints[minInt(len(waypoints)-1, i+2)]

		if i == 0 {
			p0 = Point{X: 2*p1.X - p2.X, Y: 2*p1.Y - p2.Y}
		}
		if i == len(waypoints)-2 {
			p3 = Point{X: 2*p2.X - p1.X, Y: 2*p2.Y - p1.Y}
		}

		pts := make([]Point, samplesPerSegment+1)
		cumLen := make([]float64, samplesPerSegment+1)
		for s := 0; s <= samplesPerSegment; s++ {
			t := float64(s) / float64(samplesPerSegment)
			pts[s] = catmullRom(p0, p1, p2, p3, t)
			if s > 0 {
				cumLen[s] = cumLen[s-1] + dist(pts[s-1], pts[s])
			}
		}
		segCurves[i] = pts
		segLengths[i] = cumLen

		if i == 0 {
			allTrail = append(allTrail, pts...)
		} else {
			allTrail = append(allTrail, pts[1:]...)
		}
	}

	tiles := make([]TileNode, len(vars.Plate))

	sec1 := sampleSection(segCurves[0:2], segLengths[0:2], 5)
	sec2 := sampleSection(segCurves[2:4], segLengths[2:4], 8)
	sec3 := sampleSection(segCurves[4:9], segLengths[4:9], 36)

	for i := 0; i < 6; i++ {
		tiles[i] = TileNode{Case: vars.Plate[i], Pos: sec1[i]}
	}
	for i := 1; i < len(sec2); i++ {
		tiles[5+i] = TileNode{Case: vars.Plate[5+i], Pos: sec2[i]}
	}
	for i := 1; i < len(sec3); i++ {
		tiles[13+i] = TileNode{Case: vars.Plate[13+i], Pos: sec3[i]}
	}

	return BoardLayout{
		Tiles:       tiles,
		TrailPoints: allTrail,
	}
}

func sampleSection(curves [][]Point, lengths [][]float64, intervals int) []Point {
	var pts []Point
	var cum []float64
	total := 0.0
	for segIdx := range curves {
		c := curves[segIdx]
		l := lengths[segIdx]
		startIdx := 0
		if segIdx > 0 {
			startIdx = 1
		}
		for j := startIdx; j < len(c); j++ {
			if len(pts) == 0 {
				pts = append(pts, c[j])
				cum = append(cum, 0)
			} else {
				step := 0.0
				if j > 0 {
					step = l[j] - l[j-1]
				}
				total += step
				pts = append(pts, c[j])
				cum = append(cum, total)
			}
		}
	}

	out := make([]Point, intervals+1)
	out[0] = pts[0]
	out[intervals] = pts[len(pts)-1]

	cursor := 0
	for k := 1; k < intervals; k++ {
		target := total * float64(k) / float64(intervals)
		for cursor+1 < len(cum) && cum[cursor+1] < target {
			cursor++
		}
		if cursor+1 >= len(cum) {
			out[k] = pts[len(pts)-1]
			continue
		}
		segSpan := cum[cursor+1] - cum[cursor]
		alpha := 0.0
		if segSpan > 0 {
			alpha = (target - cum[cursor]) / segSpan
		}
		out[k] = Point{
			X: pts[cursor].X + alpha*(pts[cursor+1].X-pts[cursor].X),
			Y: pts[cursor].Y + alpha*(pts[cursor+1].Y-pts[cursor].Y),
		}
	}
	return out
}

func catmullRom(p0, p1, p2, p3 Point, t float64) Point {
	t2 := t * t
	t3 := t2 * t
	x := 0.5 * ((2 * p1.X) +
		(-p0.X+p2.X)*t +
		(2*p0.X-5*p1.X+4*p2.X-p3.X)*t2 +
		(-p0.X+3*p1.X-3*p2.X+p3.X)*t3)
	y := 0.5 * ((2 * p1.Y) +
		(-p0.Y+p2.Y)*t +
		(2*p0.Y-5*p1.Y+4*p2.Y-p3.Y)*t2 +
		(-p0.Y+3*p1.Y-3*p2.Y+p3.Y)*t3)
	return Point{X: x, Y: y}
}

func dist(a, b Point) float64 {
	return math.Hypot(b.X-a.X, b.Y-a.Y)
}

func pointToSegmentDist(p, a, b Point) float64 {
	vx := b.X - a.X
	vy := b.Y - a.Y
	wx := p.X - a.X
	wy := p.Y - a.Y
	c1 := wx*vx + wy*vy
	if c1 <= 0 {
		return dist(p, a)
	}
	c2 := vx*vx + vy*vy
	if c2 <= c1 {
		return dist(p, b)
	}
	t := c1 / c2
	proj := Point{X: a.X + t*vx, Y: a.Y + t*vy}
	return dist(p, proj)
}

func segmentsIntersect(a1, a2, b1, b2 Point) bool {
	d1 := cross(b1, b2, a1)
	d2 := cross(b1, b2, a2)
	d3 := cross(a1, a2, b1)
	d4 := cross(a1, a2, b2)

	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}
	return false
}

func cross(a, b, c Point) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package board

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"sort"
	"sync"

	"orion-event-bot/bot/vars"

	"github.com/fogleman/gg"
)

type BoardLabels struct {
	LocaleKey  string
	LodgeLabel string
	StartLabel string
}

type PlayerPawn struct {
	UserID   string
	Position int
	Avatar   image.Image
}

type RenderData struct {
	LodgeTotal int
	LodgeGoal  int
	Players    []PlayerPawn
	CallerID   string
	Labels     BoardLabels
}

type tileStyle struct {
	bg      color.RGBA
	border  color.RGBA
	badgeBg color.RGBA
	isSpec  bool
}

var (
	baseCacheMu sync.RWMutex
	baseCache   = make(map[string]*image.RGBA)
)

func RenderBoardPNG(data RenderData) ([]byte, error) {
	if err := ensureAssetsLoaded(); err != nil {
		return nil, err
	}

	layout := BuildLayout()
	base, err := getOrBuildBaseImage(layout, data.Labels)
	if err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(base.Bounds())
	draw.Draw(canvas, canvas.Bounds(), base, image.Point{}, draw.Src)

	dc := gg.NewContextForRGBA(canvas)

	if err := drawRefugeGaugeDynamic(dc, data.LodgeTotal, data.LodgeGoal, data.Labels); err != nil {
		return nil, err
	}

	if err := drawPlayerPawns(dc, layout, data.Players, data.CallerID); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(&buf, canvas); err != nil {
		return nil, fmt.Errorf("encode board png: %w", err)
	}
	return buf.Bytes(), nil
}

func getOrBuildBaseImage(layout BoardLayout, labels BoardLabels) (*image.RGBA, error) {
	key := labels.LocaleKey + "|" + labels.LodgeLabel + "|" + labels.StartLabel

	baseCacheMu.RLock()
	cached, ok := baseCache[key]
	baseCacheMu.RUnlock()
	if ok {
		return cached, nil
	}

	img := image.NewRGBA(image.Rect(0, 0, BoardWidth, BoardHeight))
	dc := gg.NewContextForRGBA(img)

	drawAlpineBackdrop(dc, layout)
	drawSummitAndRefugeScenery(dc, layout, labels)
	drawSkiTrail(dc, layout)
	if err := drawGondolaCable(dc, layout); err != nil {
		return nil, err
	}
	if err := drawTiles(dc, layout); err != nil {
		return nil, err
	}

	baseCacheMu.Lock()
	baseCache[key] = img
	baseCacheMu.Unlock()

	return img, nil
}

func drawAlpineBackdrop(dc *gg.Context, layout BoardLayout) {
	grad := gg.NewLinearGradient(0, 0, BoardWidth*0.4, BoardHeight)
	grad.AddColorStop(0, color.RGBA{247, 250, 253, 255})
	grad.AddColorStop(0.45, color.RGBA{226, 240, 251, 255})
	grad.AddColorStop(1, color.RGBA{199, 224, 245, 255})
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, BoardWidth, BoardHeight)
	dc.Fill()

	slopeWash := gg.NewLinearGradient(0, 0, BoardWidth, BoardHeight)
	slopeWash.AddColorStop(0, color.RGBA{255, 252, 242, 95})
	slopeWash.AddColorStop(0.5, color.RGBA{235, 245, 255, 20})
	slopeWash.AddColorStop(1, color.RGBA{165, 202, 232, 85})
	dc.SetFillStyle(slopeWash)
	dc.DrawRectangle(0, 0, BoardWidth, BoardHeight)
	dc.Fill()

	dc.SetRGBA255(186, 214, 238, 120)
	dc.MoveTo(0, 0)
	dc.LineTo(2400, 0)
	dc.LineTo(2400, 310)
	dc.LineTo(2220, 210)
	dc.LineTo(2050, 275)
	dc.LineTo(1860, 145)
	dc.LineTo(1660, 260)
	dc.LineTo(1480, 180)
	dc.LineTo(1240, 255)
	dc.LineTo(960, 165)
	dc.LineTo(680, 225)
	dc.LineTo(420, 145)
	dc.LineTo(160, 175)
	dc.LineTo(0, 125)
	dc.ClosePath()
	dc.Fill()

	dc.SetRGBA255(154, 191, 224, 90)
	peaks := [][3]Point{
		{{420, 145}, {680, 225}, {500, 240}},
		{{960, 165}, {1240, 255}, {1040, 270}},
		{{1480, 180}, {1660, 260}, {1540, 285}},
		{{1860, 145}, {2050, 275}, {1910, 295}},
		{{2220, 210}, {2400, 310}, {2250, 320}},
	}
	for _, tri := range peaks {
		dc.MoveTo(tri[0].X, tri[0].Y)
		dc.LineTo(tri[1].X, tri[1].Y)
		dc.LineTo(tri[2].X, tri[2].Y)
		dc.ClosePath()
		dc.Fill()
	}

	dc.SetLineCapRound()
	for band := 0; band < 7; band++ {
		if band%2 == 0 {
			dc.SetRGBA255(132, 180, 222, 85)
			dc.SetLineWidth(2.4)
		} else {
			dc.SetRGBA255(148, 194, 232, 55)
			dc.SetLineWidth(1.6)
		}
		startY := 90.0 + float64(band)*185.0
		dc.MoveTo(0, startY)
		for x := 0.0; x <= BoardWidth; x += 40 {
			slopeOffset := (x / BoardWidth) * 260.0
			wave := 38.0*math.Sin(x*0.0032+float64(band)*0.85) + 18.0*math.Cos(x*0.0085-float64(band)*1.3)
			dc.LineTo(x, startY+slopeOffset+wave)
		}
		dc.Stroke()
	}

	drawPineGroves(dc, layout)
}

func drawPineGroves(dc *gg.Context, layout BoardLayout) {
	rng := rand.New(rand.NewSource(772026))
	p6 := layout.Tiles[GondolaFromNumber-1].Pos
	p14 := layout.Tiles[GondolaToNumber-1].Pos

	isClearSpot := func(p Point, minTrailDist float64) bool {
		if p.X < 45 || p.X > BoardWidth-45 || p.Y < 45 || p.Y > BoardHeight-45 {
			return false
		}
		if p.X > 1580 && p.Y < 345 {
			return false
		}
		if p.X < 310 && p.Y < 260 {
			return false
		}
		if pointToSegmentDist(p, p6, p14) < 105 {
			return false
		}
		for i := 0; i < len(layout.TrailPoints); i += 5 {
			if dist(p, layout.TrailPoints[i]) < minTrailDist {
				return false
			}
		}
		for _, t := range layout.Tiles {
			if dist(p, Point{X: t.Pos.X, Y: t.Pos.Y - 22}) < minTrailDist+12 {
				return false
			}
		}
		return true
	}

	grovesPlaced := 0
	for attempts := 0; attempts < 500 && grovesPlaced < 20; attempts++ {
		cx := 70 + rng.Float64()*(BoardWidth-140)
		cy := 90 + rng.Float64()*(BoardHeight-150)
		if !isClearSpot(Point{X: cx, Y: cy}, 125) {
			continue
		}
		grovesPlaced++

		depthScale := 0.55 + (cy/BoardHeight)*0.85
		count := 3 + rng.Intn(4)
		for k := 0; k < count; k++ {
			ox := (rng.Float64() - 0.5) * 145
			oy := (rng.Float64() - 0.5) * 95
			tp := Point{X: cx + ox, Y: cy + oy}
			if !isClearSpot(tp, 108) {
				continue
			}
			treeScale := depthScale * (0.82 + rng.Float64()*0.36)
			drawDecorPine(dc, tp.X, tp.Y, treeScale)
		}

		if rng.Float64() < 0.45 {
			rx := cx + (rng.Float64()-0.5)*110
			ry := cy + 30 + rng.Float64()*40
			if isClearSpot(Point{X: rx, Y: ry}, 105) {
				drawDecorRock(dc, rx, ry, depthScale*0.9)
			}
		}
	}
}

func drawDecorPine(dc *gg.Context, x, y, scale float64) {
	dc.SetRGBA255(30, 58, 138, 32)
	dc.DrawEllipse(x+6*scale, y+18*scale, 23*scale, 8*scale)
	dc.Fill()

	dc.SetRGB255(109, 68, 40)
	dc.DrawRectangle(x-3.5*scale, y+4*scale, 7*scale, 14*scale)
	dc.Fill()

	layers := []struct {
		dy, w, h float64
		col      color.RGBA
	}{
		{7, 23, 25, color.RGBA{20, 83, 45, 235}},
		{-6, 19, 22, color.RGBA{22, 101, 52, 240}},
		{-18, 14, 19, color.RGBA{21, 128, 61, 245}},
	}
	for _, l := range layers {
		dc.SetColor(l.col)
		dc.MoveTo(x, y+(l.dy-l.h)*scale)
		dc.LineTo(x-l.w*scale, y+l.dy*scale)
		dc.LineTo(x+l.w*scale, y+l.dy*scale)
		dc.ClosePath()
		dc.Fill()

		dc.SetRGBA255(248, 252, 255, 215)
		dc.MoveTo(x, y+(l.dy-l.h)*scale)
		dc.LineTo(x-l.w*0.46*scale, y+(l.dy-l.h*0.52)*scale)
		dc.LineTo(x+l.w*0.28*scale, y+(l.dy-l.h*0.64)*scale)
		dc.ClosePath()
		dc.Fill()
	}
}

func drawDecorRock(dc *gg.Context, x, y, scale float64) {
	dc.SetRGBA255(30, 58, 138, 26)
	dc.DrawEllipse(x+5*scale, y+11*scale, 21*scale, 7*scale)
	dc.Fill()

	dc.SetRGBA255(100, 116, 139, 215)
	dc.MoveTo(x-19*scale, y+9*scale)
	dc.LineTo(x-10*scale, y-12*scale)
	dc.LineTo(x+7*scale, y-15*scale)
	dc.LineTo(x+20*scale, y+8*scale)
	dc.ClosePath()
	dc.Fill()

	dc.SetRGBA255(241, 245, 249, 220)
	dc.MoveTo(x-10*scale, y-12*scale)
	dc.LineTo(x+7*scale, y-15*scale)
	dc.LineTo(x+11*scale, y-3*scale)
	dc.LineTo(x-4*scale, y-5*scale)
	dc.ClosePath()
	dc.Fill()
}

func drawSummitAndRefugeScenery(dc *gg.Context, layout BoardLayout, labels BoardLabels) {
	p1 := layout.Tiles[0].Pos
	drawDecorRock(dc, p1.X-62, p1.Y+28, 1.35)
	drawDecorRock(dc, p1.X+52, p1.Y-48, 1.05)

	sx, sy := p1.X-10, p1.Y-88
	dc.SetRGB255(109, 68, 40)
	dc.DrawRectangle(sx-4, sy+12, 8, 28)
	dc.Fill()
	signW := 124.0
	dc.SetRGBA255(15, 23, 42, 40)
	dc.DrawRoundedRectangle(sx-signW/2+3, sy-16+3, signW, 30, 6)
	dc.Fill()
	dc.SetRGB255(146, 84, 42)
	dc.DrawRoundedRectangle(sx-signW/2, sy-16, signW, 30, 6)
	dc.FillPreserve()
	dc.SetRGB255(92, 52, 24)
	dc.SetLineWidth(2.5)
	dc.Stroke()

	if f, err := newFontFace(weightBold, 16); err == nil {
		dc.SetFontFace(f)
		dc.SetRGB255(255, 248, 231)
		dc.DrawStringAnchored(labels.StartLabel, sx, sy-1, 0.5, 0.45)
	}

	drawWoodenRefugeIllustration(dc, 1735, 138, labels)
}

func drawWoodenRefugeIllustration(dc *gg.Context, cx, cy float64, labels BoardLabels) {
	dc.SetRGBA255(255, 255, 255, 210)
	dc.DrawEllipse(cx+190, cy+55, 310, 48)
	dc.Fill()

	dc.SetRGB255(100, 116, 139)
	dc.DrawRectangle(cx+36, cy-72, 20, 38)
	dc.Fill()
	dc.SetRGBA255(255, 255, 255, 185)
	dc.DrawCircle(cx+49, cy-88, 9)
	dc.DrawCircle(cx+61, cy-103, 12)
	dc.DrawCircle(cx+78, cy-116, 14)
	dc.Fill()

	dc.SetRGBA255(15, 23, 42, 35)
	dc.DrawRoundedRectangle(cx-62+4, cy-22+6, 124, 72, 6)
	dc.Fill()

	dc.SetRGB255(146, 84, 42)
	dc.DrawRoundedRectangle(cx-62, cy-22, 124, 72, 6)
	dc.FillPreserve()
	dc.SetRGB255(92, 52, 24)
	dc.SetLineWidth(3)
	dc.Stroke()

	dc.SetRGBA255(92, 52, 24, 110)
	dc.SetLineWidth(2)
	for y := cy - 6; y < cy+42; y += 14 {
		dc.DrawLine(cx-58, y, cx+58, y)
		dc.Stroke()
	}

	dc.SetRGB255(92, 52, 24)
	dc.DrawRoundedRectangle(cx-42, cy+8, 28, 42, 4)
	dc.Fill()
	dc.SetRGB255(251, 191, 36)
	dc.DrawRoundedRectangle(cx+6, cy+2, 34, 28, 4)
	dc.FillPreserve()
	dc.SetRGB255(92, 52, 24)
	dc.SetLineWidth(3)
	dc.Stroke()
	dc.DrawLine(cx+23, cy+2, cx+23, cy+30)
	dc.DrawLine(cx+6, cy+16, cx+40, cy+16)
	dc.Stroke()

	dc.SetRGB255(92, 52, 24)
	dc.MoveTo(cx-78, cy-18)
	dc.LineTo(cx, cy-74)
	dc.LineTo(cx+78, cy-18)
	dc.ClosePath()
	dc.Fill()

	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(11)
	dc.SetLineCapRound()
	dc.SetLineJoinRound()
	dc.MoveTo(cx-80, cy-22)
	dc.LineTo(cx, cy-78)
	dc.LineTo(cx+80, cy-22)
	dc.Stroke()

	px := cx + 98
	py := cy - 72
	dc.SetRGBA255(15, 23, 42, 35)
	dc.DrawRoundedRectangle(px+4, py+5, 420, 74, 14)
	dc.Fill()

	dc.SetRGB255(30, 58, 95)
	dc.DrawRoundedRectangle(px, py, 420, 74, 14)
	dc.FillPreserve()
	dc.SetRGB255(146, 84, 42)
	dc.SetLineWidth(4)
	dc.Stroke()

	if f, err := newFontFace(weightBold, 34); err == nil {
		dc.SetFontFace(f)
		dc.SetRGB255(255, 248, 231)
		dc.DrawStringAnchored("WINTER RUSH", px+210, py+28, 0.5, 0.45)
	}
}

func drawRefugeGaugeDynamic(dc *gg.Context, lodgeTotal, lodgeGoal int, labels BoardLabels) error {
	_ = labels
	if lodgeGoal <= 0 {
		lodgeGoal = vars.LodgeGoal
	}
	ratio := float64(lodgeTotal) / float64(lodgeGoal)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	gx := 1833.0
	gy := 155.0
	gw := 420.0
	gh := 38.0

	dc.SetRGBA255(15, 23, 42, 40)
	dc.DrawRoundedRectangle(gx+3, gy+4, gw, gh, 19)
	dc.Fill()

	dc.SetRGB255(78, 44, 21)
	dc.DrawRoundedRectangle(gx, gy, gw, gh, 19)
	dc.FillPreserve()
	dc.SetRGB255(146, 84, 42)
	dc.SetLineWidth(3.5)
	dc.Stroke()

	pad := 5.0
	ix, iy, iw, ih := gx+pad, gy+pad, gw-2*pad, gh-2*pad
	dc.SetRGB255(36, 22, 13)
	dc.DrawRoundedRectangle(ix, iy, iw, ih, ih/2)
	dc.Fill()

	fillW := iw * ratio
	if fillW > 0 {
		if fillW < ih {
			fillW = ih
		}
		grad := gg.NewLinearGradient(ix, iy, ix+iw, iy)
		grad.AddColorStop(0, color.RGBA{234, 88, 12, 255})
		grad.AddColorStop(0.55, color.RGBA{245, 158, 11, 255})
		grad.AddColorStop(1, color.RGBA{16, 185, 129, 255})
		dc.SetFillStyle(grad)
		dc.DrawRoundedRectangle(ix, iy, fillW, ih, ih/2)
		dc.Fill()
	}

	drawIconSnowflakeSmall(dc, gx+22, gy+gh/2, 9)

	pct := int(math.Round(ratio * 100))
	txt := fmt.Sprintf("%d / %d (%d%%)", lodgeTotal, lodgeGoal, pct)
	f, err := newFontFace(weightBold, 18)
	if err != nil {
		return err
	}
	dc.SetFontFace(f)
	dc.SetRGBA255(15, 23, 42, 180)
	dc.DrawStringAnchored(txt, gx+gw/2+1.5, gy+gh/2+1.5, 0.5, 0.45)
	dc.SetRGB255(255, 250, 240)
	dc.DrawStringAnchored(txt, gx+gw/2, gy+gh/2, 0.5, 0.45)

	return nil
}

func drawSkiTrail(dc *gg.Context, layout BoardLayout) {
	tracePath := func() {
		dc.MoveTo(layout.TrailPoints[0].X, layout.TrailPoints[0].Y)
		for i := 1; i < len(layout.TrailPoints); i++ {
			dc.LineTo(layout.TrailPoints[i].X, layout.TrailPoints[i].Y)
		}
	}

	dc.SetLineCapRound()
	dc.SetLineJoinRound()

	dc.SetRGBA255(30, 64, 138, 30)
	dc.SetLineWidth(76)
	tracePath()
	dc.Stroke()

	dc.SetRGB255(132, 191, 236)
	dc.SetLineWidth(64)
	tracePath()
	dc.Stroke()

	dc.SetRGB255(253, 254, 255)
	dc.SetLineWidth(54)
	tracePath()
	dc.Stroke()

	dc.SetRGBA255(147, 197, 253, 185)
	dc.SetLineWidth(3.5)
	dc.SetDash(14, 14)
	tracePath()
	dc.Stroke()
	dc.SetDash()
}

func drawGondolaCable(dc *gg.Context, layout BoardLayout) error {
	p6 := layout.Tiles[GondolaFromNumber-1].Pos
	p14 := layout.Tiles[GondolaToNumber-1].Pos

	dx := p14.X - p6.X
	dy := p14.Y - p6.Y
	length := math.Hypot(dx, dy)
	nx, ny := -dy/length, dx/length

	dc.SetRGB255(68, 64, 60)
	dc.SetLineWidth(4.5)
	dc.DrawLine(p6.X+nx*5, p6.Y+ny*5, p14.X+nx*5, p14.Y+ny*5)
	dc.Stroke()

	return nil
}

func drawTiles(dc *gg.Context, layout BoardLayout) error {
	numFontBig, err := newFontFace(weightBold, 27)
	if err != nil {
		return err
	}
	numFontSmall, err := newFontFace(weightBold, 16)
	if err != nil {
		return err
	}

	half := TileRadius
	size := half * 2

	for _, tile := range layout.Tiles {
		style := styleForTile(tile.Case)
		x := tile.Pos.X
		y := tile.Pos.Y

		dc.SetRGBA255(15, 23, 42, 52)
		dc.DrawRoundedRectangle(x-half+2.5, y-half+6, size, size, 20)
		dc.Fill()

		dc.SetColor(style.bg)
		dc.DrawRoundedRectangle(x-half, y-half, size, size, 20)
		dc.FillPreserve()
		dc.SetColor(style.border)
		dc.SetLineWidth(4)
		dc.Stroke()

		if !style.isSpec {
			dc.SetFontFace(numFontBig)
			dc.SetRGB255(30, 41, 59)
			dc.DrawStringAnchored(fmt.Sprintf("%d", tile.Case.Number), x, y, 0.5, 0.44)
			continue
		}

		dc.SetRGBA255(255, 255, 255, 45)
		dc.SetLineWidth(2)
		dc.DrawRoundedRectangle(x-half+4, y-half+4, size-8, size-8, 16)
		dc.Stroke()

		drawTileVectorIcon(dc, tile.Case, x+3, y+4, 42, style.bg)

		pillW, pillH := 34.0, 24.0
		px := x - half - 5
		py := y - half - 5
		dc.SetColor(style.badgeBg)
		dc.DrawRoundedRectangle(px, py, pillW, pillH, 9)
		dc.FillPreserve()
		dc.SetRGB255(255, 255, 255)
		dc.SetLineWidth(2.2)
		dc.Stroke()

		dc.SetFontFace(numFontSmall)
		dc.SetRGB255(255, 255, 255)
		dc.DrawStringAnchored(fmt.Sprintf("%d", tile.Case.Number), px+pillW/2, py+pillH/2, 0.5, 0.44)
	}

	return nil
}

func styleForTile(c vars.PlateCase) tileStyle {
	switch c.SpecialCase {
	case vars.SpecialCaseTree:
		return tileStyle{
			bg: color.RGBA{22, 101, 52, 255}, border: color.RGBA{134, 239, 172, 255},
			badgeBg: color.RGBA{20, 83, 45, 255}, isSpec: true,
		}
	case vars.SpecialCaseRavine:
		return tileStyle{
			bg: color.RGBA{194, 65, 12, 255}, border: color.RGBA{253, 186, 116, 255},
			badgeBg: color.RGBA{154, 52, 18, 255}, isSpec: true,
		}
	case vars.SpecialCaseYeti:
		return tileStyle{
			bg: color.RGBA{107, 33, 168, 255}, border: color.RGBA{216, 180, 254, 255},
			badgeBg: color.RGBA{88, 28, 135, 255}, isSpec: true,
		}
	case vars.SpecialCaseSlush:
		return tileStyle{
			bg: color.RGBA{71, 85, 105, 255}, border: color.RGBA{203, 213, 225, 255},
			badgeBg: color.RGBA{51, 65, 85, 255}, isSpec: true,
		}
	case vars.SpecialCaseJump:
		return tileStyle{
			bg: color.RGBA{234, 88, 12, 255}, border: color.RGBA{254, 215, 170, 255},
			badgeBg: color.RGBA{194, 65, 12, 255}, isSpec: true,
		}
	case vars.SpecialCaseBlizzard:
		return tileStyle{
			bg: color.RGBA{29, 78, 216, 255}, border: color.RGBA{147, 197, 253, 255},
			badgeBg: color.RGBA{30, 64, 175, 255}, isSpec: true,
		}
	case vars.SpecialCaseChairlift:
		return tileStyle{
			bg: color.RGBA{67, 56, 202, 255}, border: color.RGBA{165, 180, 252, 255},
			badgeBg: color.RGBA{49, 46, 129, 255}, isSpec: true,
		}
	case vars.SpecialCaseIce:
		return tileStyle{
			bg: color.RGBA{8, 145, 178, 235}, border: color.RGBA{165, 243, 252, 255},
			badgeBg: color.RGBA{14, 116, 144, 255}, isSpec: true,
		}
	case vars.SpecialCaseGondola:
		return tileStyle{
			bg: color.RGBA{146, 64, 14, 255}, border: color.RGBA{253, 230, 138, 255},
			badgeBg: color.RGBA{120, 53, 15, 255}, isSpec: true,
		}
	case vars.SpecialCaseSupply:
		return tileStyle{
			bg: color.RGBA{15, 118, 110, 255}, border: color.RGBA{153, 246, 228, 255},
			badgeBg: color.RGBA{17, 94, 89, 255}, isSpec: true,
		}
	default:
		if c.Number == 1 {
			return tileStyle{
				bg: color.RGBA{21, 128, 61, 255}, border: color.RGBA{187, 247, 208, 255},
				badgeBg: color.RGBA{20, 83, 45, 255}, isSpec: true,
			}
		}
		if c.Number == len(vars.Plate) {
			return tileStyle{
				bg: color.RGBA{217, 119, 6, 255}, border: color.RGBA{254, 240, 138, 255},
				badgeBg: color.RGBA{180, 83, 9, 255}, isSpec: true,
			}
		}
		return tileStyle{
			bg: color.RGBA{255, 253, 249, 255}, border: color.RGBA{148, 184, 215, 255},
			badgeBg: color.RGBA{30, 41, 59, 255}, isSpec: false,
		}
	}
}

func drawPlayerPawns(dc *gg.Context, layout BoardLayout, players []PlayerPawn, callerID string) error {
	if len(players) == 0 {
		return nil
	}

	grouped := make(map[int][]PlayerPawn)
	for _, p := range players {
		if p.Position < 1 {
			continue
		}
		pos := p.Position
		if pos > len(layout.Tiles) {
			pos = len(layout.Tiles)
		}
		grouped[pos] = append(grouped[pos], p)
	}

	badgeFont, err := newFontFace(weightBold, 15)
	if err != nil {
		return err
	}

	for pos, list := range grouped {
		tile := layout.Tiles[pos-1]
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].UserID == callerID {
				return true
			}
			if list[j].UserID == callerID {
				return false
			}
			return list[i].UserID < list[j].UserID
		})

		const maxVisible = 3
		visibleCount := len(list)
		if visibleCount > maxVisible {
			visibleCount = maxVisible
		}
		extra := len(list) - visibleCount

		step := 34.0
		totalSpan := float64(visibleCount-1) * step
		if extra > 0 {
			totalSpan += 30.0
		}
		startX := tile.Pos.X - totalSpan/2
		centerY := tile.Pos.Y - TileRadius - 18

		for i := visibleCount - 1; i >= 0; i-- {
			pawn := list[i]
			px := startX + float64(i)*step
			py := centerY
			isCaller := pawn.UserID == callerID

			radius := 20.0
			if isCaller {
				radius = 23.0
			}

			dc.SetRGBA255(15, 23, 42, 75)
			dc.DrawCircle(px+1.5, py+3, radius+3.5)
			dc.Fill()

			if isCaller {
				dc.SetRGBA255(250, 204, 21, 135)
				dc.DrawCircle(px, py, radius+6.5)
				dc.Fill()
				dc.SetRGB255(250, 204, 21)
				dc.DrawCircle(px, py, radius+3.5)
				dc.Fill()
			} else {
				dc.SetRGB255(255, 255, 255)
				dc.DrawCircle(px, py, radius+3)
				dc.Fill()
			}

			dc.Push()
			dc.DrawCircle(px, py, radius)
			dc.Clip()
			if pawn.Avatar != nil {
				drawImageCentered(dc, pawn.Avatar, px, py, radius*2, radius*2)
			} else {
				dc.SetColor(fallbackColor(pawn.UserID))
				dc.DrawRectangle(px-radius, py-radius, radius*2, radius*2)
				dc.Fill()
				drawIconSnowflakeSmall(dc, px, py, radius*0.55)
			}
			dc.Pop()
			dc.ResetClip()
		}

		if extra > 0 {
			bx := startX + float64(visibleCount-1)*step + 28
			by := centerY
			dc.SetRGB255(15, 23, 42)
			dc.DrawRoundedRectangle(bx-17, by-13, 35, 26, 13)
			dc.FillPreserve()
			dc.SetRGB255(255, 255, 255)
			dc.SetLineWidth(2)
			dc.Stroke()

			dc.SetFontFace(badgeFont)
			dc.SetRGB255(255, 255, 255)
			dc.DrawStringAnchored(fmt.Sprintf("+%d", extra), bx+1, by, 0.5, 0.45)
		}
	}

	return nil
}

func drawImageCentered(dc *gg.Context, img image.Image, cx, cy, w, h float64) {
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return
	}
	sx := w / float64(b.Dx())
	sy := h / float64(b.Dy())
	dc.Push()
	dc.Translate(cx, cy)
	dc.Scale(sx, sy)
	dc.DrawImageAnchored(img, 0, 0, 0.5, 0.5)
	dc.Pop()
}

func fallbackColor(userID string) color.RGBA {
	palette := []color.RGBA{
		{37, 99, 235, 255},
		{14, 116, 144, 255},
		{5, 150, 105, 255},
		{124, 58, 237, 255},
		{217, 119, 6, 255},
		{225, 29, 72, 255},
	}
	var sum uint32
	for i := 0; i < len(userID); i++ {
		sum = sum*31 + uint32(userID[i])
	}
	return palette[int(sum)%len(palette)]
}

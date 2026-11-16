package board

import (
	"image/color"
	"math"

	"orion-event-bot/bot/vars"

	"github.com/fogleman/gg"
)

func drawTileVectorIcon(dc *gg.Context, c vars.PlateCase, cx, cy, size float64, tileBg color.RGBA) {
	dc.Push()
	dc.Translate(cx, cy)
	s := size / 44.0
	dc.Scale(s, s)
	dc.SetLineCapRound()
	dc.SetLineJoinRound()

	switch c.SpecialCase {
	case vars.SpecialCaseGondola:
		drawIconGondola(dc, tileBg)
	case vars.SpecialCaseChairlift:
		drawIconChairlift(dc)
	case vars.SpecialCaseIce:
		drawIconIce(dc)
	case vars.SpecialCaseTree:
		drawIconTree(dc)
	case vars.SpecialCaseRavine:
		drawIconRavine(dc, tileBg)
	case vars.SpecialCaseYeti:
		drawIconYeti(dc)
	case vars.SpecialCaseSlush:
		drawIconSlush(dc)
	case vars.SpecialCaseJump:
		drawIconJump(dc)
	case vars.SpecialCaseBlizzard:
		drawIconBlizzard(dc)
	case vars.SpecialCaseSupply:
		drawIconSupply(dc, tileBg)
	default:
		if c.Number == 1 {
			drawIconStart(dc)
		} else if c.Number == len(vars.Plate) {
			drawIconFinish(dc, tileBg)
		}
	}

	dc.Pop()
}

func drawIconStart(dc *gg.Context) {
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(3.6)
	dc.DrawLine(-9, 15, -9, -15)
	dc.Stroke()
	dc.MoveTo(-9, -14)
	dc.LineTo(14, -7)
	dc.LineTo(-9, 2)
	dc.ClosePath()
	dc.Fill()
	dc.DrawCircle(-9, 15, 3)
	dc.Fill()
}

func drawIconFinish(dc *gg.Context, tileBg color.RGBA) {
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(3.5)
	dc.DrawLine(-11, 16, -11, -15)
	dc.Stroke()

	fx, fy := -9.0, -14.0
	cell := 5.5
	for r := 0; r < 3; r++ {
		for col := 0; col < 4; col++ {
			if (r+col)%2 == 0 {
				dc.SetRGB255(255, 255, 255)
			} else {
				dc.SetColor(tileBg)
			}
			dc.DrawRectangle(fx+float64(col)*cell, fy+float64(r)*cell, cell, cell)
			dc.Fill()
		}
	}
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(2)
	dc.DrawRectangle(fx, fy, 4*cell, 3*cell)
	dc.Stroke()
}

func drawIconGondola(dc *gg.Context, tileBg color.RGBA) {
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(3)
	dc.DrawLine(-17, -14, 17, -16)
	dc.Stroke()
	dc.DrawLine(0, -15, 0, -6)
	dc.Stroke()
	dc.DrawRoundedRectangle(-12, -6, 24, 21, 5)
	dc.Fill()
	dc.SetColor(tileBg)
	dc.DrawRoundedRectangle(-8.5, -2.5, 7, 8, 2)
	dc.Fill()
	dc.DrawRoundedRectangle(1.5, -2.5, 7, 8, 2)
	dc.Fill()
}

func drawIconChairlift(dc *gg.Context) {
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(3.2)
	dc.DrawLine(-16, -15, 16, -15)
	dc.Stroke()
	dc.DrawLine(-2, -15, -2, -3)
	dc.LineTo(-8, 7)
	dc.LineTo(9, 7)
	dc.Stroke()
	dc.SetLineWidth(4.2)
	dc.DrawLine(-9, 10, 9, 10)
	dc.Stroke()
	dc.DrawLine(-9, 10, -9, -1)
	dc.Stroke()
	dc.SetLineWidth(2.6)
	dc.DrawLine(4, 10, 7, 16)
	dc.DrawLine(2, 16, 12, 16)
	dc.Stroke()
}

func drawIconIce(dc *gg.Context) {
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(3.0)
	for i := 0; i < 6; i++ {
		ang := float64(i) * math.Pi / 3
		c, s := math.Cos(ang), math.Sin(ang)
		dc.DrawLine(0, 0, c*16, s*16)
		dc.Stroke()

		bx, by := c*10, s*10
		b1 := ang + math.Pi/4
		b2 := ang - math.Pi/4
		dc.DrawLine(bx, by, bx+math.Cos(b1)*5, by+math.Sin(b1)*5)
		dc.Stroke()
		dc.DrawLine(bx, by, bx+math.Cos(b2)*5, by+math.Sin(b2)*5)
		dc.Stroke()
	}
	dc.DrawCircle(0, 0, 3.2)
	dc.Fill()
}

func drawIconTree(dc *gg.Context) {
	dc.SetRGB255(255, 255, 255)
	dc.MoveTo(0, -17)
	dc.LineTo(-10, -3)
	dc.LineTo(10, -3)
	dc.ClosePath()
	dc.Fill()
	dc.MoveTo(0, -8)
	dc.LineTo(-14, 9)
	dc.LineTo(14, 9)
	dc.ClosePath()
	dc.Fill()
	dc.DrawRectangle(-3, 9, 6, 8)
	dc.Fill()
}

func drawIconRavine(dc *gg.Context, tileBg color.RGBA) {
	dc.SetRGB255(255, 255, 255)
	dc.MoveTo(-17, 14)
	dc.LineTo(-17, -6)
	dc.LineTo(-4, -6)
	dc.LineTo(-8, 4)
	dc.LineTo(-2, 14)
	dc.ClosePath()
	dc.Fill()
	dc.MoveTo(17, 14)
	dc.LineTo(17, -6)
	dc.LineTo(5, -6)
	dc.LineTo(8, 3)
	dc.LineTo(3, 14)
	dc.ClosePath()
	dc.Fill()
	dc.SetLineWidth(3.2)
	dc.DrawLine(0, -16, 0, -9)
	dc.Stroke()
	dc.DrawCircle(0, -4, 1.8)
	dc.Fill()
	_ = tileBg
}

func drawIconYeti(dc *gg.Context) {
	dc.SetRGB255(255, 255, 255)
	dc.DrawEllipse(0, 6, 9, 8)
	dc.Fill()
	toes := []struct{ x, y, rx, ry float64 }{
		{-10, -6, 3.2, 4.4},
		{-3.5, -10, 3.2, 4.8},
		{3.5, -10, 3.2, 4.8},
		{10, -6, 3.2, 4.4},
	}
	for _, t := range toes {
		dc.DrawEllipse(t.x, t.y, t.rx, t.ry)
		dc.Fill()
	}
}

func drawIconSlush(dc *gg.Context) {
	dc.SetRGB255(255, 255, 255)
	dc.MoveTo(-6, -15)
	dc.QuadraticTo(-13, -4, -12, 1)
	dc.QuadraticTo(-11, 7, -6, 7)
	dc.QuadraticTo(-1, 7, 0, 1)
	dc.QuadraticTo(1, -4, -6, -15)
	dc.ClosePath()
	dc.Fill()

	dc.MoveTo(9, -8)
	dc.QuadraticTo(4, -1, 5, 2)
	dc.QuadraticTo(6, 6, 9, 6)
	dc.QuadraticTo(12, 6, 13, 2)
	dc.QuadraticTo(14, -1, 9, -8)
	dc.ClosePath()
	dc.Fill()

	dc.SetLineWidth(3.2)
	dc.MoveTo(-15, 13)
	dc.QuadraticTo(-7, 9, 0, 13)
	dc.QuadraticTo(7, 17, 15, 13)
	dc.Stroke()
}

func drawIconJump(dc *gg.Context) {
	dc.SetRGB255(255, 255, 255)
	dc.MoveTo(-15, 14)
	dc.QuadraticTo(-2, 13, 8, 2)
	dc.LineTo(8, 14)
	dc.ClosePath()
	dc.Fill()

	dc.SetLineWidth(3.2)
	dc.MoveTo(-12, 2)
	dc.QuadraticTo(0, -13, 13, -10)
	dc.Stroke()

	dc.MoveTo(15, -10)
	dc.LineTo(7, -15)
	dc.LineTo(9, -6)
	dc.ClosePath()
	dc.Fill()
}

func drawIconBlizzard(dc *gg.Context) {
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(3.2)
	dc.MoveTo(-15, -6)
	dc.LineTo(6, -6)
	dc.QuadraticTo(13, -6, 13, -11)
	dc.QuadraticTo(13, -15, 8, -15)
	dc.Stroke()
	dc.MoveTo(-17, 3)
	dc.LineTo(11, 3)
	dc.QuadraticTo(16, 3, 16, 8)
	dc.QuadraticTo(16, 12, 11, 12)
	dc.Stroke()
	dc.DrawCircle(-10, 12, 2.4)
	dc.DrawCircle(-1, 13, 2.2)
	dc.DrawCircle(1, -13, 2.2)
	dc.Fill()
}

func drawIconSupply(dc *gg.Context, tileBg color.RGBA) {
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(3.2)
	dc.DrawArc(0, -10, 5.5, math.Pi, 2*math.Pi)
	dc.Stroke()
	dc.DrawRoundedRectangle(-13, -9, 26, 24, 6)
	dc.Fill()
	dc.SetColor(tileBg)
	dc.DrawRoundedRectangle(-2.5, -4, 5, 14, 1.5)
	dc.Fill()
	dc.DrawRoundedRectangle(-7, 0.5, 14, 5, 1.5)
	dc.Fill()
}

func drawIconSnowflakeSmall(dc *gg.Context, cx, cy, radius float64) {
	dc.Push()
	dc.Translate(cx, cy)
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(2.4)
	dc.SetLineCapRound()
	for i := 0; i < 6; i++ {
		ang := float64(i) * math.Pi / 3
		dc.DrawLine(0, 0, math.Cos(ang)*radius, math.Sin(ang)*radius)
		dc.Stroke()
	}
	dc.Pop()
}

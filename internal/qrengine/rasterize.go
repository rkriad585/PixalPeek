package qrengine

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	xdraw "golang.org/x/image/draw"
)

type moduleStyle struct {
	fg         color.RGBA
	bg         color.RGBA
	shape      string
	scale      int
	quiet      int
	total      int
	actualSize int
}

func layoutFor(matrix [][]bool, opts EncodeOptions, style *moduleStyle) error {
	mods := len(matrix)
	quiet := opts.QuietZone
	total := mods + quiet*2
	scale := opts.Size / total
	if scale < 1 {
		scale = 1
	}
	style.fg = mustColor(opts.FGColor)
	style.bg = mustColor(opts.BGColor)
	style.shape = opts.Shape
	style.scale = scale
	style.quiet = quiet
	style.total = total
	style.actualSize = scale * total
	return nil
}

func mustColor(hex string) color.RGBA {
	c, err := parseHexColor(hex)
	if err != nil {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	return c
}

func Rasterize(matrix [][]bool, opts EncodeOptions, logo image.Image) (*image.RGBA, error) {
	var st moduleStyle
	if err := layoutFor(matrix, opts, &st); err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(image.Rect(0, 0, st.actualSize, st.actualSize))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{st.bg}, image.Point{}, draw.Src)

	radius := float64(st.scale) * 0.42
	dotR := float64(st.scale) * 0.46

	dark := func(mx, my int) bool {
		if mx < 0 || my < 0 || my >= len(matrix) || mx >= len(matrix[my]) {
			return false
		}
		return matrix[my][mx]
	}

	n := len(matrix)
	inFinder := func(mx, my int) bool {
		return (mx < 7 && my < 7) || (mx >= n-7 && my < 7) || (mx < 7 && my >= n-7)
	}

	for my := 0; my < len(matrix); my++ {
		for mx := 0; mx < len(matrix[my]); mx++ {
			if !matrix[my][mx] {
				continue
			}
			x0 := (mx + st.quiet) * st.scale
			y0 := (my + st.quiet) * st.scale
			if opts.Shape != ShapeSquare && inFinder(mx, my) {
				fillRect(canvas, x0, y0, st.scale, st.scale, st.fg)
				continue
			}
			switch opts.Shape {
			case ShapeDot:
				fillDot(canvas, x0, y0, st.scale, dotR, st.fg)
			case ShapeRounded:
				fillRoundedConnected(canvas, x0, y0, st.scale, radius, st.fg,
					dark(mx-1, my), dark(mx, my-1), dark(mx+1, my), dark(mx, my+1))
			default:
				fillRect(canvas, x0, y0, st.scale, st.scale, st.fg)
			}
		}
	}

	if logo != nil {
		overlayLogo(canvas, logo, st)
	}
	return canvas, nil
}

func fillRect(img *image.RGBA, x0, y0, w, h int, c color.RGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func fillDot(img *image.RGBA, x0, y0, size int, r float64, c color.RGBA) {
	cx := float64(x0) + float64(size)/2
	cy := float64(y0) + float64(size)/2
	for y := y0; y < y0+size; y++ {
		for x := x0; x < x0+size; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			if dx*dx+dy*dy <= r*r {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func fillRoundedConnected(img *image.RGBA, x0, y0, size int, radius float64, c color.RGBA,
	leftDark, upDark, rightDark, downDark bool) {
	r := radius
	for y := y0; y < y0+size; y++ {
		for x := x0; x < x0+size; x++ {
			px := float64(x) + 0.5 - float64(x0)
			py := float64(y) + 0.5 - float64(y0)
			visible := true
			switch {
			case px < r && py < r:
				visible = leftDark || upDark || dist(px, py, r, r) <= r
			case px > float64(size)-r && py < r:
				visible = rightDark || upDark || dist(px, py, float64(size)-r, r) <= r
			case px < r && py > float64(size)-r:
				visible = leftDark || downDark || dist(px, py, r, float64(size)-r) <= r
			case px > float64(size)-r && py > float64(size)-r:
				visible = rightDark || downDark || dist(px, py, float64(size)-r, float64(size)-r) <= r
			}
			if visible {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func insideRounded(px, py, rx, ry, size, radius float64) bool {
	left, right := rx, rx+size
	top, bottom := ry, ry+size
	if px < left+radius && py < top+radius {
		return dist(px, py, left+radius, top+radius) <= radius
	}
	if px > right-radius && py < top+radius {
		return dist(px, py, right-radius, top+radius) <= radius
	}
	if px < left+radius && py > bottom-radius {
		return dist(px, py, left+radius, bottom-radius) <= radius
	}
	if px > right-radius && py > bottom-radius {
		return dist(px, py, right-radius, bottom-radius) <= radius
	}
	return true
}

func dist(x1, y1, x2, y2 float64) float64 {
	dx, dy := x1-x2, y1-y2
	return math.Sqrt(dx*dx + dy*dy)
}

func overlayLogo(canvas *image.RGBA, logo image.Image, st moduleStyle) {
	size := st.actualSize
	logoSide := size * 18 / 100
	if logoSide < 8 {
		return
	}
	pad := st.scale / 2
	boxSide := logoSide + pad*2
	origin := (size - boxSide) / 2

	bg := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	fillRect(canvas, origin, origin, boxSide, boxSide, bg)

	scaled := image.NewRGBA(image.Rect(0, 0, logoSide, logoSide))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), logo, logo.Bounds(), xdraw.Over, nil)

	logoX := (size - logoSide) / 2
	logoY := (size - logoSide) / 2
	for y := 0; y < logoSide; y++ {
		for x := 0; x < logoSide; x++ {
			c := scaled.RGBAAt(x, y)
			if c.A > 0 {
				canvas.SetRGBA(logoX+x, logoY+y, blendOver(canvas.RGBAAt(logoX+x, logoY+y), c))
			}
		}
	}
}

func blendUnder(dst color.RGBA, src color.RGBA) color.RGBA {
	a := float64(src.A) / 255
	return color.RGBA{
		R: uint8(float64(src.R)*a + float64(dst.R)*(1-a)),
		G: uint8(float64(src.G)*a + float64(dst.G)*(1-a)),
		B: uint8(float64(src.B)*a + float64(dst.B)*(1-a)),
		A: 255,
	}
}

func blendOver(dst color.RGBA, src color.RGBA) color.RGBA {
	return blendUnder(dst, src)
}

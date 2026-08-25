package qrengine

import (
	"fmt"
	"image"
	"strings"
)

func svgModule(x, y, s int, shape string, fg string) string {
	switch shape {
	case ShapeDot:
		r := float64(s) * 0.46
		return fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="%.2f"/>`,
			float64(x)+float64(s)/2, float64(y)+float64(s)/2, r)
	case ShapeRounded:
		r := float64(s) * 0.38
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="%.2f" ry="%.2f"/>`, x, y, s, s, r, r)
	default:
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d"/>`, x, y, s, s)
	}
}

func ToSVG(matrix [][]bool, opts EncodeOptions, logo image.Image) ([]byte, error) {
	var st moduleStyle
	if err := layoutFor(matrix, opts, &st); err != nil {
		return nil, err
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="%d" viewBox="0 0 %d %d">`,
		st.actualSize, st.actualSize, st.actualSize, st.actualSize))
	body.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="%s"/>`, st.actualSize, st.actualSize, opts.BGColor))
	body.WriteString(fmt.Sprintf(`<g fill="%s">`, opts.FGColor))

	n := len(matrix)
	inFinder := func(mx, my int) bool {
		return (mx < 7 && my < 7) || (mx >= n-7 && my < 7) || (mx < 7 && my >= n-7)
	}

	for my := 0; my < len(matrix); my++ {
		for mx := 0; mx < len(matrix[my]); mx++ {
			if !matrix[my][mx] {
				continue
			}
			x := (mx + st.quiet) * st.scale
			y := (my + st.quiet) * st.scale
			if opts.Shape != ShapeSquare && inFinder(mx, my) {
				body.WriteString(svgModule(x, y, st.scale, ShapeSquare, opts.FGColor))
				continue
			}
			body.WriteString(svgModule(x, y, st.scale, opts.Shape, opts.FGColor))
		}
	}
	body.WriteString(`</g>`)

	if logo != nil && opts.LogoB64 != "" {
		b64 := opts.LogoB64
		if idx := strings.Index(b64, ";base64,"); idx != -1 {
			mime := b64[:idx+8]
			b64 = b64[idx+8:]
			side := st.actualSize * 18 / 100
			pad := st.scale / 2
			boxSide := side + pad*2
			bo := (st.actualSize - boxSide) / 2
			o := (st.actualSize - side) / 2
			body.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#FFFFFF"/>`, bo, bo, boxSide, boxSide))
			body.WriteString(fmt.Sprintf(`<image x="%d" y="%d" width="%d" height="%d" xlink:href="%s%s"/>`, o, o, side, side, mime, b64))
		}
	} else if logo != nil {
		return nil, fmt.Errorf("svg logo embedding requires LogoB64 data")
	}

	body.WriteString(`</svg>`)
	return []byte(body.String()), nil
}

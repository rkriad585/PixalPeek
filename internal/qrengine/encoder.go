package qrengine

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

func parseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color %q", s)
	}
	var rgb int64
	if _, err := fmt.Sscanf(s, "%06x", &rgb); err != nil {
		return color.RGBA{}, fmt.Errorf("invalid hex color %q", s)
	}
	return color.RGBA{
		R: uint8(rgb >> 16),
		G: uint8((rgb >> 8) & 0xFF),
		B: uint8(rgb & 0xFF),
		A: 0xFF,
	}, nil
}

func recoveryLevel(ecc string) qrcode.RecoveryLevel {
	switch strings.ToUpper(ecc) {
	case "L":
		return qrcode.Low
	case "Q":
		return qrcode.High
	case "H":
		return qrcode.Highest
	default:
		return qrcode.Medium
	}
}

func loadLogo(opts EncodeOptions) (image.Image, error) {
	if opts.LogoPath != "" {
		data, err := os.ReadFile(opts.LogoPath)
		if err != nil {
			return nil, fmt.Errorf("logo file unreadable: %w", err)
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("logo image invalid: %w", err)
		}
		return img, nil
	}
	if opts.LogoB64 != "" {
		b64 := opts.LogoB64
		if idx := strings.Index(b64, ";base64,"); idx != -1 {
			b64 = b64[idx+8:]
		}
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("logo base64 invalid: %w", err)
		}
		img, _, err := image.Decode(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("logo image invalid: %w", err)
		}
		return img, nil
	}
	return nil, nil
}

func matrixFor(opts EncodeOptions) ([][]bool, error) {
	if strings.TrimSpace(opts.Content) == "" {
		return nil, fmt.Errorf("content must not be empty")
	}
	qr, err := qrcode.New(opts.Content, recoveryLevel(opts.ECC))
	if err != nil {
		return nil, fmt.Errorf("failed to encode content: %w", err)
	}
	qr.DisableBorder = true
	return qr.Bitmap(), nil
}

func ValidateStyle(opts EncodeOptions) []string {
	var warnings []string
	fg, err1 := parseHexColor(opts.FGColor)
	bg, err2 := parseHexColor(opts.BGColor)
	if err1 == nil && err2 == nil {
		r := contrastRatio(fg, bg)
		if r < 3.0 {
			warnings = append(warnings, fmt.Sprintf("low contrast ratio %.2f:1 between foreground and background; scanners may fail", r))
		}
	}
	hasLogo := opts.LogoPath != "" || opts.LogoB64 != ""
	if hasLogo && opts.ECC != "H" && opts.QuietZone < 2 {
		warnings = append(warnings, "center logo with low quiet zone: consider ECC level H for reliable scanning")
	}
	if hasLogo && !strings.EqualFold(opts.ECC, "H") {
		warnings = append(warnings, "a center logo obscures modules; ECC level H is recommended")
	}
	if opts.Size < 128 {
		warnings = append(warnings, "output below 128px may be hard to scan from screens")
	}
	return warnings
}

func Generate(opts EncodeOptions) ([]byte, error) {
	NormalizeEncodeOptions(&opts)

	if opts.Format != FormatPNG && opts.Format != FormatJPG && opts.Format != FormatSVG && opts.Format != FormatPDF {
		return nil, fmt.Errorf("unsupported format %q (use png, jpg, svg or pdf)", opts.Format)
	}

	matrix, err := matrixFor(opts)
	if err != nil {
		return nil, err
	}

	logo, err := loadLogo(opts)
	if err != nil {
		return nil, err
	}

	if opts.Format == FormatSVG {
		return ToSVG(matrix, opts, logo)
	}

	img, err := Rasterize(matrix, opts, logo)
	if err != nil {
		return nil, err
	}

	switch opts.Format {
	case FormatPNG:
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("png encode failed: %w", err)
		}
		return buf.Bytes(), nil
	case FormatJPG:
		flat := flatten(img)
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, flat, &jpeg.Options{Quality: 95}); err != nil {
			return nil, fmt.Errorf("jpeg encode failed: %w", err)
		}
		return buf.Bytes(), nil
	case FormatPDF:
		flat := flatten(img)
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, flat, &jpeg.Options{Quality: 92}); err != nil {
			return nil, fmt.Errorf("jpeg encode failed: %w", err)
		}
		return JPEGToPDF(buf.Bytes(), img.Bounds().Dx())
	default:
		return nil, fmt.Errorf("unsupported format %q", opts.Format)
	}
}

func flatten(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	out := image.NewRGBA(b)
	bg, _ := parseHexColor("#FFFFFF")
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.RGBAAt(x, y)
			a := float64(c.A) / 255
			n := color.RGBA{
				R: uint8(float64(c.R)*a + float64(bg.R)*(1-a)),
				G: uint8(float64(c.G)*a + float64(bg.G)*(1-a)),
				B: uint8(float64(c.B)*a + float64(bg.B)*(1-a)),
				A: 255,
			}
			out.SetRGBA(x, y, n)
		}
	}
	return out
}

func relativeLuminance(c color.RGBA) float64 {
	lin := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

func contrastRatio(a, b color.RGBA) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

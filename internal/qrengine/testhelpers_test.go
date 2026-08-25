package qrengine

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
)

func b64encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func assertFileExists(t interface{ Fatalf(string, ...interface{}) }, path string) {
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
}

func composeSideBySide(a []byte, b []byte) ([]byte, error) {
	imgA, _, err := image.Decode(bytes.NewReader(a))
	if err != nil {
		return nil, err
	}
	imgB, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	wA, hA := imgA.Bounds().Dx(), imgA.Bounds().Dy()
	wB, hB := imgB.Bounds().Dx(), imgB.Bounds().Dy()
	h := hA
	if hB > h {
		h = hB
	}
	canvas := image.NewRGBA(image.Rect(0, 0, wA+wB+40, h))
	for y := 0; y < h; y++ {
		for x := 0; x < canvas.Bounds().Dx(); x++ {
			canvas.Set(x, y, color.White)
		}
	}
	drawAt(canvas, imgA, image.Pt(10, (h-hA)/2))
	drawAt(canvas, imgB, image.Pt(wA+30, (h-hB)/2))
	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawAt(dst *image.RGBA, src image.Image, at image.Point) {
	for y := 0; y < src.Bounds().Dy(); y++ {
		for x := 0; x < src.Bounds().Dx(); x++ {
			dst.Set(at.X+x, at.Y+y, src.At(src.Bounds().Min.X+x, src.Bounds().Min.Y+y))
		}
	}
}

type fatalfable interface {
	Helper()
	Fatalf(string, ...interface{})
}

func solidTestPNG(t fatalfable) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 48, 48))
	red := color.RGBA{R: 215, G: 25, B: 33, A: 255}
	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			img.Set(x, y, red)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("solid png: %v", err)
	}
	return buf.Bytes()
}

func encodePNGHelper(img image.Image) []byte { var buf bytes.Buffer; png.Encode(&buf, img); return buf.Bytes() }

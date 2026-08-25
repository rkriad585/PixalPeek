package qrengine

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"time"

	"github.com/makiuchi-d/gozxing"
	multiqrcode "github.com/makiuchi-d/gozxing/multi/qrcode"
	qrreader "github.com/makiuchi-d/gozxing/qrcode"
)

func decodeHints() map[gozxing.DecodeHintType]interface{} {
	return map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER: true,
	}
}

func boundingBoxFromPoints(pts []gozxing.ResultPoint) BoundingBox {
	if len(pts) == 0 {
		return BoundingBox{}
	}
	sumMin, sumMax := pts[0], pts[0]
	diffMin, diffMax := pts[0], pts[0]
	for _, p := range pts[1:] {
		if p.GetX()+p.GetY() < sumMin.GetX()+sumMin.GetY() {
			sumMin = p
		}
		if p.GetX()+p.GetY() > sumMax.GetX()+sumMax.GetY() {
			sumMax = p
		}
		if p.GetX()-p.GetY() < diffMin.GetX()-diffMin.GetY() {
			diffMin = p
		}
		if p.GetX()-p.GetY() > diffMax.GetX()-diffMax.GetY() {
			diffMax = p
		}
	}
	return BoundingBox{
		TopLeft:     [2]float64{sumMin.GetX(), sumMin.GetY()},
		BottomRight: [2]float64{sumMax.GetX(), sumMax.GetY()},
		TopRight:    [2]float64{diffMax.GetX(), diffMax.GetY()},
		BottomLeft:  [2]float64{diffMin.GetX(), diffMin.GetY()},
	}
}

func resultToDecodeResult(r *gozxing.Result) DecodeResult {
	ecc := ""
	if meta := r.GetResultMetadata(); meta != nil {
		if v, ok := meta[gozxing.ResultMetadataType_ERROR_CORRECTION_LEVEL]; ok {
			ecc = fmt.Sprintf("%v", v)
		}
	}
	content := r.GetText()
	return DecodeResult{
		Content:              content,
		ContentType:          DetectContentType(content),
		Format:               "QR_CODE",
		ErrorCorrectionLevel: ecc,
		BoundingBox:          boundingBoxFromPoints(r.GetResultPoints()),
	}
}

func decodeFromImage(img image.Image, multiScan bool) []DecodeResult {
	reader := qrreader.NewQRCodeReader()
	var out []DecodeResult
	seen := map[string]bool{}

	add := func(rs []*gozxing.Result) {
		for _, r := range rs {
			dr := resultToDecodeResult(r)
			if !seen[dr.Content] {
				seen[dr.Content] = true
				out = append(out, dr)
			}
		}
	}

	if bmp, err := gozxing.NewBinaryBitmapFromImage(img); err == nil {
		if res, err := reader.Decode(bmp, decodeHints()); err == nil {
			add([]*gozxing.Result{res})
		}
	}

	if multiScan || len(out) == 0 {
		mr := multiqrcode.NewQRCodeMultiReader()
		if bmp2, err2 := gozxing.NewBinaryBitmapFromImage(img); err2 == nil {
			if rs, err := mr.DecodeMultiple(bmp2, decodeHints()); err == nil {
				add(rs)
			}
		}
	}
	return out
}

func DecodeImage(r io.Reader, sourcePath string, multiScan bool) ScanResponse {
	resp := ScanResponse{
		Success:    false,
		SourceFile: sourcePath,
		ScannedAt:  time.Now(),
		Results:    []DecodeResult{},
	}

	img, format, err := image.Decode(r)
	if err != nil {
		resp.Error = FailureBadImage + fmt.Sprintf(" (%v)", err)
		return resp
	}
	_ = format

	results := decodeFromImage(img, multiScan)
	if len(results) == 0 {
		resp.Error = FailureNoQR
		return resp
	}
	resp.Success = true
	resp.Results = AssessURLSecurityForResults(results)
	return resp
}

func DecodeFile(filePath string, multiScan bool) ScanResponse {
	f, err := os.Open(filePath)
	if err != nil {
		return ScanResponse{
			Success:    false,
			SourceFile: filePath,
			ScannedAt:  time.Now(),
			Results:    []DecodeResult{},
			Error:      FailureFileUnread + fmt.Sprintf(": %v", err),
		}
	}
	defer f.Close()
	return DecodeImage(f, filePath, multiScan)
}

func DecodeBytes(data []byte, sourceLabel string, multiScan bool) ScanResponse {
	return DecodeImage(bytes.NewReader(data), sourceLabel, multiScan)
}

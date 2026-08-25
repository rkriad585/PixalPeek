package qrengine

import "time"

type ContentType string

const (
	TypeText   ContentType = "text"
	TypeURL    ContentType = "url"
	TypeWiFi   ContentType = "wifi"
	TypeEmail  ContentType = "email"
	TypeSMS    ContentType = "sms"
	TypePhone  ContentType = "phone"
	TypeVCard  ContentType = "vcard"
	TypeGeo    ContentType = "geo"
	TypeEvent  ContentType = "event"
	TypeSocial ContentType = "social"
)

var AllTypes = []ContentType{TypeText, TypeURL, TypeWiFi, TypeEmail, TypeSMS, TypePhone, TypeVCard, TypeGeo, TypeEvent, TypeSocial}

func (t ContentType) Valid() bool {
	for _, v := range AllTypes {
		if v == t {
			return true
		}
	}
	return false
}

const (
	FormatPNG = "png"
	FormatJPG = "jpg"
	FormatSVG = "svg"
	FormatPDF = "pdf"
)

var AllFormats = []string{FormatPNG, FormatJPG, FormatSVG, FormatPDF}

var ShapeSquare = "square"
var ShapeRounded = "rounded"
var ShapeDot = "dot"

var AllShapes = []string{ShapeSquare, ShapeRounded, ShapeDot}

const (
	FailureNoQR       = "no QR code detected in image"
	FailureBadImage   = "unsupported or corrupted image"
	FailureFileUnread = "file not found or unreadable"
)

type BoundingBox struct {
	TopLeft     [2]float64 `json:"top_left"`
	TopRight    [2]float64 `json:"top_right"`
	BottomLeft  [2]float64 `json:"bottom_left"`
	BottomRight [2]float64 `json:"bottom_right,omitempty"`
}

type DecodeResult struct {
	Content              string            `json:"content"`
	ContentType          ContentType       `json:"content_type"`
	Format               string            `json:"format"`
	ErrorCorrectionLevel string            `json:"error_correction_level"`
	BoundingBox          BoundingBox       `json:"bounding_box"`
	Security             *SafetyAssessment `json:"security,omitempty"`
}

type ScanResponse struct {
	Success    bool           `json:"success"`
	SourceFile string         `json:"source_file,omitempty"`
	ScannedAt  time.Time      `json:"scanned_at"`
	Results    []DecodeResult `json:"results"`
	Error      string         `json:"error,omitempty"`
}

type EncodeOptions struct {
	Content   string      `json:"content"`
	Type      ContentType `json:"type,omitempty"`
	Size      int         `json:"size"`
	ECC       string      `json:"ecc"`
	FGColor   string      `json:"fg_color"`
	BGColor   string      `json:"bg_color"`
	Shape     string      `json:"shape"`
	LogoPath  string      `json:"-"`
	LogoB64   string      `json:"logo_b64,omitempty"`
	Format    string      `json:"format"`
	QuietZone int         `json:"quiet_zone"`
}

func NormalizeEncodeOptions(opts *EncodeOptions) {
	if opts.Size <= 0 {
		opts.Size = 512
	}
	if opts.Size > 4096 {
		opts.Size = 4096
	}
	switch opts.ECC {
	case "L", "M", "Q", "H":
	default:
		opts.ECC = "M"
	}
	if opts.FGColor == "" {
		opts.FGColor = "#000000"
	}
	if opts.BGColor == "" {
		opts.BGColor = "#FFFFFF"
	}
	if opts.Shape == "" {
		opts.Shape = ShapeSquare
	}
	if opts.Format == "jpeg" {
		opts.Format = FormatJPG
	}
	if opts.Format == "" {
		opts.Format = FormatPNG
	}
	if opts.QuietZone <= 0 {
		opts.QuietZone = 4
	}
	if opts.QuietZone > 8 {
		opts.QuietZone = 8
	}
}

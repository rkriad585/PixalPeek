package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rkriad585/PixalPeek/internal/qrengine"
)

const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitInvalidArgs      = 2
	ExitFileNotFound     = 3
	ExitNoQRCodeDetected = 4
	ExitUnsupportedImage = 5
)

var Version = "1.0.0-beta"

var WorkerRunner func(cfg *Config) int
var ClipboardCopier func(png []byte) int

type Config struct {
	DecodePath   string
	GenerateText string
	TypeHint     string
	OutputPath   string
	Size         int
	ECC          string
	FGColor      string
	BGColor      string
	LogoPath     string
	Format       string
	Margin       int
	Shape        string
	Multi        bool
	Quiet        bool
	NoColor      bool
	BatchPath    string
	Zip          bool
	Camera       bool
	Clipboard    bool
	ScanDir      string
	Verbose      bool
	Version      bool
	Help         bool
}

func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("pixalpeek", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := &Config{}
	fs.StringVar(&cfg.DecodePath, "qr", "", "decode a QR code from an image file")
	fs.StringVar(&cfg.DecodePath, "qrcode", "", "decode a QR code from an image file")
	fs.StringVar(&cfg.GenerateText, "g", "", "generate a QR code encoding the given content")
	fs.StringVar(&cfg.GenerateText, "generate", "", "generate a QR code encoding the given content")
	fs.StringVar(&cfg.OutputPath, "o", "", "write result to file instead of stdout ('-' for stdout JSON)")
	fs.StringVar(&cfg.OutputPath, "output", "", "write result to file instead of stdout")
	fs.StringVar(&cfg.TypeHint, "t", "", "content-type hint: text,url,wifi,email,sms,phone,vcard,geo,event,social")
	fs.StringVar(&cfg.TypeHint, "type", "", "content-type hint")
	fs.IntVar(&cfg.Size, "size", 512, "output image size in pixels")
	fs.StringVar(&cfg.ECC, "ecc", "M", "error-correction level: L, M, Q, H")
	fs.StringVar(&cfg.FGColor, "fg", "#000000", "foreground hex color")
	fs.StringVar(&cfg.BGColor, "bg", "#FFFFFF", "background hex color")
	fs.StringVar(&cfg.LogoPath, "logo", "", "embed a logo image in the center")
	fs.StringVar(&cfg.Format, "format", "png", "output format: png, jpg, svg, pdf")
	fs.IntVar(&cfg.Margin, "margin", 4, "quiet-zone margin in modules (0-8)")
	fs.StringVar(&cfg.Shape, "shape", "square", "module style: square, rounded, dot")
	fs.BoolVar(&cfg.Multi, "multi", false, "detect and return all QR codes in the image")
	fs.StringVar(&cfg.BatchPath, "batch", "", "batch-generate from a CSV or JSON list")
	fs.BoolVar(&cfg.Zip, "zip", false, "also produce a zip archive when batching")
	fs.BoolVar(&cfg.Camera, "camera", false, "scan live from the default webcam")
	fs.BoolVar(&cfg.Clipboard, "clipboard", false, "use the clipboard: decode an image, or copy generated code")
	fs.StringVar(&cfg.ScanDir, "scan-dir", "", "scan all images in a directory recursively")
	fs.BoolVar(&cfg.Quiet, "quiet", false, "suppress non-essential output")
	fs.BoolVar(&cfg.Quiet, "s", false, "suppress non-essential output")
	fs.BoolVar(&cfg.Quiet, "silent", false, "suppress non-essential output")
	fs.BoolVar(&cfg.NoColor, "no-color", false, "disable ANSI colors")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "verbose diagnostics on stderr")
	fs.BoolVar(&cfg.Version, "v", false, "print version and exit")
	fs.BoolVar(&cfg.Version, "version", false, "print version and exit")
	fs.BoolVar(&cfg.Help, "h", false, "print usage and exit")
	fs.BoolVar(&cfg.Help, "help", false, "print usage and exit")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return cfg, nil
}

var knownFlags = []string{
	"-qr", "--qrcode", "-g", "--generate", "-o", "--output",
	"-t", "--type", "--size", "--ecc", "--fg", "--bg", "--logo",
	"--format", "--margin", "--shape", "--multi", "--batch", "--zip",
	"--camera", "--clipboard", "--scan-dir", "--quiet", "-s", "--silent",
	"--no-color", "--verbose", "-v", "--version", "-h", "--help",
}

func IsFlagged(args []string) bool {
	if len(args) <= 1 {
		return false
	}
	for _, arg := range args[1:] {
		base := arg
		if idx := strings.Index(arg, "="); idx > 1 {
			base = arg[:idx]
		}
		for _, kf := range knownFlags {
			if base == kf {
				return true
			}
		}
	}
	return false
}

func validate(cfg *Config) error {
	if strings.TrimSpace(cfg.GenerateText) == "" && cfg.DecodePath == "" &&
		cfg.BatchPath == "" && cfg.ScanDir == "" && !cfg.Camera && !cfg.Clipboard && !cfg.Help && !cfg.Version {
		return fmt.Errorf("nothing to do: pass -qr <file>, -g <text>, --batch <file> or --scan-dir <dir>")
	}
	if cfg.DecodePath != "" && cfg.GenerateText != "" {
		return fmt.Errorf("-qr and -g are mutually exclusive")
	}
	switch strings.ToUpper(cfg.ECC) {
	case "L", "M", "Q", "H":
	default:
		return fmt.Errorf("invalid --ecc %q (use L, M, Q or H)", cfg.ECC)
	}
	if cfg.Size < 64 || cfg.Size > 4096 {
		return fmt.Errorf("invalid --size %d (allowed 64-4096)", cfg.Size)
	}
	if cfg.Margin < 0 || cfg.Margin > 8 {
		return fmt.Errorf("invalid --margin %d (allowed 0-8)", cfg.Margin)
	}
	validShape := false
	for _, s := range qrengine.AllShapes {
		if s == cfg.Shape {
			validShape = true
		}
	}
	if !validShape {
		return fmt.Errorf("invalid --shape %q (use square, rounded or dot)", cfg.Shape)
	}
	fmt_ := strings.ToLower(cfg.Format)
	if fmt_ == "jpeg" {
		fmt_ = "jpg"
	}
	validFormat := false
	for _, f := range qrengine.AllFormats {
		if f == fmt_ {
			validFormat = true
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid --format %q (use png, jpg, svg or pdf)", cfg.Format)
	}
	cfg.Format = fmt_
	if cfg.TypeHint != "" {
		hint := qrengine.ContentType(strings.ToLower(cfg.TypeHint))
		if !hint.Valid() {
			return fmt.Errorf("invalid --type %q (use text, url, wifi, email, sms, phone, vcard, geo, event, social)", cfg.TypeHint)
		}
		cfg.TypeHint = string(hint)
	}
	return nil
}

func Execute(cfg *Config) int {
	if cfg.Help {
		PrintUsage(os.Stdout)
		return ExitSuccess
	}
	if cfg.Version {
		fmt.Printf("PIXALPEEK v%s\nengine: gozxing + skip2/go-qrcode\n", Version)
		return ExitSuccess
	}

	if err := validate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'pixalpeek -h' for usage.\n")
		return ExitInvalidArgs
	}

	if cfg.BatchPath != "" {
		return runBatch(cfg)
	}

	if cfg.ScanDir != "" {
		return runScanDir(cfg)
	}

	if cfg.Camera || (cfg.Clipboard && cfg.GenerateText == "") {
		if WorkerRunner == nil {
			fmt.Fprintln(os.Stderr, "Error: --camera/--clipboard require the desktop build of PixalPeek")
			return ExitGeneralError
		}
		return WorkerRunner(cfg)
	}

	if cfg.DecodePath != "" {
		return runDecode(cfg)
	}

	if cfg.GenerateText != "" {
		return runGenerate(cfg)
	}

	return ExitInvalidArgs
}

func runDecode(cfg *Config) int {
	var res qrengine.ScanResponse
	if cfg.DecodePath == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read from stdin: %v\n", err)
			return ExitGeneralError
		}
		res = qrengine.DecodeBytes(data, "stdin://", cfg.Multi)
	} else {
		if _, err := os.Stat(cfg.DecodePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s: %v\n", qrengine.FailureFileUnread, err)
			return ExitFileNotFound
		}
		res = qrengine.DecodeFile(cfg.DecodePath, cfg.Multi)
	}

	if !res.Success {
		if cfg.OutputPath != "" && cfg.OutputPath != "-" {
			if werr := WriteJSONFile(cfg.OutputPath, res); werr != nil {
				fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", werr)
				return ExitGeneralError
			}
		}
		fmt.Fprintf(os.Stderr, "Error: %s (%s)\n", res.Error, cfg.DecodePath)
		if strings.HasPrefix(res.Error, qrengine.FailureBadImage) {
			return ExitUnsupportedImage
		}
		return ExitNoQRCodeDetected
	}

	if cfg.Clipboard {
		res.SourceFile = "clipboard://" + cfg.DecodePath
	}

	if cfg.OutputPath == "-" {
		if err := WriteJSON(os.Stdout, res); err != nil {
			return ExitGeneralError
		}
		return ExitSuccess
	}
	if cfg.OutputPath != "" {
		if err := WriteJSONFile(cfg.OutputPath, res); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			return ExitGeneralError
		}
		if !cfg.Quiet {
			fmt.Fprintf(os.Stdout, "Scan result saved to %s\n", cfg.OutputPath)
		}
		return ExitSuccess
	}

	PrintScanResult(os.Stdout, res, cfg)
	return ExitSuccess
}

func runGenerate(cfg *Config) int {
	content := applyTypeHint(cfg.GenerateText, cfg.TypeHint)

	opts := qrengine.EncodeOptions{
		Content:   content,
		Type:      qrengine.ContentType(cfg.TypeHint),
		Size:      cfg.Size,
		ECC:       strings.ToUpper(cfg.ECC),
		FGColor:   cfg.FGColor,
		BGColor:   cfg.BGColor,
		Shape:     cfg.Shape,
		LogoPath:  cfg.LogoPath,
		Format:    cfg.Format,
		QuietZone: cfg.Margin,
	}

	data, err := qrengine.Generate(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: generation failed: %v\n", err)
		return ExitInvalidArgs
	}
	for _, w := range qrengine.ValidateStyle(opts) {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}

	if cfg.Clipboard {
		if ClipboardCopier == nil {
			fmt.Fprintln(os.Stderr, "Error: --clipboard requires the desktop build of PixalPeek")
			return ExitGeneralError
		}
		code := ClipboardCopier(data)
		if code != ExitSuccess {
			return code
		}
		if !cfg.Quiet {
			fmt.Fprintln(os.Stdout, "QR code copied to clipboard")
		}
		if cfg.OutputPath == "" {
			return ExitSuccess
		}
	}

	target := cfg.OutputPath
	if target == "" {
		target = "pixalpeek_qr." + opts.Format
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing image: %v\n", err)
		return ExitGeneralError
	}
	if !cfg.Quiet {
		fmt.Fprintf(os.Stdout, "Generated QR -> %s (%dx%d, ecc=%s, format=%s)\n",
			target, opts.Size, opts.Size, opts.ECC, opts.Format)
	}
	return ExitSuccess
}

func runBatch(cfg *Config) int {
	entries, err := qrengine.ParseBatchFile(cfg.BatchPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if strings.HasPrefix(err.Error(), "batch file unreadable") {
			return ExitFileNotFound
		}
		return ExitInvalidArgs
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "Error: batch file contains no entries")
		return ExitInvalidArgs
	}

	opts := qrengine.EncodeOptions{
		Size:      cfg.Size,
		ECC:       strings.ToUpper(cfg.ECC),
		FGColor:   cfg.FGColor,
		BGColor:   cfg.BGColor,
		Shape:     cfg.Shape,
		LogoPath:  cfg.LogoPath,
		Format:    cfg.Format,
		QuietZone: cfg.Margin,
	}

	outDir := cfg.OutputPath
	logf := func(format string, args ...interface{}) {
		if !cfg.Quiet {
			fmt.Fprintf(os.Stderr, format, args...)
		}
	}
	written, failures, err := qrengine.BatchGenerate(entries, opts, outDir, cfg.Zip, logf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitGeneralError
	}
	if len(written) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no QR codes could be generated")
		return ExitGeneralError
	}
	if failures > 0 {
		fmt.Fprintf(os.Stderr, "Completed with %d/%d failures\n", failures, len(entries))
		return ExitGeneralError
	}
	if !cfg.Quiet {
		fmt.Fprintf(os.Stdout, "Batch complete: %d codes in %s\n", len(written), outDirOr(cfg.OutputPath))
	}
	return ExitSuccess
}

func runScanDir(cfg *Config) int {
	if _, err := os.Stat(cfg.ScanDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: directory not found: %s\n", cfg.ScanDir)
		return ExitFileNotFound
	}

	if !cfg.Quiet {
		fmt.Fprintf(os.Stderr, "Scanning directory: %s (multi=%v)...\n", cfg.ScanDir, cfg.Multi)
	}

	res, err := qrengine.DecodeDirectory(cfg.ScanDir, cfg.Multi)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitGeneralError
	}

	if cfg.OutputPath == "-" {
		if err := WriteJSON(os.Stdout, res); err != nil {
			return ExitGeneralError
		}
		return ExitSuccess
	}
	if cfg.OutputPath != "" {
		if err := WriteJSONFile(cfg.OutputPath, res); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			return ExitGeneralError
		}
		if !cfg.Quiet {
			fmt.Fprintf(os.Stdout, "Scan result saved to %s\n", cfg.OutputPath)
		}
		return ExitSuccess
	}

	if !cfg.Quiet {
		fmt.Fprintf(os.Stdout, "\nDirectory Scan Complete\n")
		fmt.Fprintf(os.Stdout, "════════════════════════════════════\n")
		fmt.Fprintf(os.Stdout, "  Files scanned : %d\n", res.TotalFilesScanned)
		fmt.Fprintf(os.Stdout, "  QR codes found: %d\n", res.SuccessfulDecodes)
		fmt.Fprintf(os.Stdout, "  Duration      : %dms\n", res.DurationMs)
		fmt.Fprintf(os.Stdout, "════════════════════════════════════\n\n")
	}

	for _, entry := range res.Entries {
		if !cfg.Quiet {
			PrintScanResult(os.Stdout, entry, cfg)
			fmt.Fprintln(os.Stdout)
		}
	}

	if res.SuccessfulDecodes == 0 {
		fmt.Fprintln(os.Stderr, "No QR codes found in any images in the directory")
		return ExitNoQRCodeDetected
	}
	return ExitSuccess
}

func outDirOr(p string) string {
	if p == "" {
		return "out_qr"
	}
	return p
}

func applyTypeHint(content, hint string) string {
	c := strings.TrimSpace(content)
	upper := strings.ToUpper(c)
	switch strings.ToLower(hint) {
	case "wifi":
		if !strings.HasPrefix(upper, "WIFI:") {
			return mustBuild(hint, map[string]string{"ssid": c})
		}
	case "email":
		if !strings.HasPrefix(upper, "MAILTO:") {
			return mustBuild(hint, map[string]string{"to": c})
		}
	case "sms":
		if !strings.HasPrefix(upper, "SMSTO:") {
			return mustBuild(hint, map[string]string{"number": c})
		}
	case "phone":
		if !strings.HasPrefix(upper, "TEL:") {
			return mustBuild(hint, map[string]string{"number": c})
		}
	case "geo":
		if !strings.HasPrefix(upper, "GEO:") {
			parts := strings.SplitN(c, ",", 2)
			f := map[string]string{"lat": strings.TrimSpace(parts[0])}
			if len(parts) == 2 {
				f["lng"] = strings.TrimSpace(parts[1])
			}
			return mustBuild(hint, f)
		}
	case "url":
		if !strings.Contains(strings.ToLower(c), "://") && !strings.HasPrefix(strings.ToLower(c), "www.") {
			return "https://" + c
		}
	}
	return content
}

func mustBuild(hint string, fields map[string]string) string {
	s, err := qrengine.BuildContent(qrengine.BuildContentRequest{Type: qrengine.ContentType(hint), Fields: fields})
	if err != nil {
		return ""
	}
	return s
}

func PrintUsage(w *os.File) {
	fmt.Fprintf(w, `PIXALPEEK v%s - QR Code Scanner & Generator

Usage:
  pixalpeek -qr <image> [options]        decode QR code(s) from an image
  pixalpeek -g <content> [options]       generate a QR code image
  pixalpeek --batch <csv|json> [options] batch-generate from a list
  pixalpeek --scan-dir <dir> [options]   scan all images in a directory

Decode options:
  -qr, --qrcode <path>    image file to decode
  -o,  --output <path>    write JSON to file ('-' for stdout); pretty print otherwise
  --multi                 detect all QR codes in one image
  --camera                scan live from the default webcam
  --clipboard             decode an image from the clipboard
  --scan-dir <dir>        scan all images in a directory recursively

Generate options:
  -g,  --generate <text>  payload to encode
  -t,  --type <type>      hint: text url wifi email sms phone vcard geo event social
  -o,  --output <path>    output image path (default pixalpeek_qr.<format>)
  --size <px>             image size in pixels (default 512)
  --ecc <L|M|Q|H>         error-correction level (default M)
  --fg / --bg <hex>       foreground/background colors
  --logo <path>           center logo image
  --format <fmt>          png, jpg, svg, pdf (default png)
  --shape <style>         square, rounded, dot (default square)
  --margin <modules>      quiet-zone margin 0-8 (default 4)
  --clipboard             also copy the generated image to the clipboard

Batch options:
  --batch <path>          CSV or JSON entries ({name?, content})
  --zip                   additionally write <outdir>.zip
  -o,  --output <dir>     output directory (default ./out_qr)

General:
  --quiet, -s             suppress non-essential output
  --no-color              disable ANSI colors
  --verbose               debug output on stderr
  -v, --version           print version
  -h, --help              show this help

Examples:
  pixalpeek -qr invoice.png
  pixalpeek -qr photo.png --multi -o results.json
  pixalpeek -g "https://mysite.com" -o site.png --ecc H
  pixalpeek -g "MyWifi" -t wifi --ecc H -o wifi.png
  pixalpeek --batch contacts.csv -t vcard -o ./out/ --zip
  pixalpeek --scan-dir ./images --multi -o scan_results.json
`, Version)
}

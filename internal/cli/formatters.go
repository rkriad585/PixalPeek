package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rkriad585/PixalPeek/internal/qrengine"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorBold  = "\033[1m"
	colorDim   = "\033[90m"
	colorWhite = "\033[97m"
)

func useColor(noColor bool, w io.Writer) bool {
	if noColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	if f, ok := w.(*os.File); ok {
		if fi, err := f.Stat(); err == nil {
			return fi.Mode()&os.ModeCharDevice != 0
		}
	}
	return false
}

type painter struct{ on bool }

func (p painter) dim(s string) string {
	if !p.on {
		return s
	}
	return colorDim + s + colorReset
}

func (p painter) red(s string) string {
	if !p.on {
		return s
	}
	return colorRed + s + colorReset
}

func (p painter) white(s string) string {
	if !p.on {
		return s
	}
	return colorWhite + s + colorReset
}

func (p painter) bold(s string) string {
	if !p.on {
		return s
	}
	return colorBold + s + colorReset
}

func PrintScanResult(w io.Writer, res qrengine.ScanResponse, cfg *Config) {
	if cfg.Quiet {
		for _, r := range res.Results {
			fmt.Fprintln(w, r.Content)
		}
		return
	}
	p := painter{on: useColor(cfg.NoColor, w)}
	line := p.dim(strings.Repeat("─", 62))
	fmt.Fprintf(w, "%s\n", line)
	fmt.Fprintf(w, "%s %s  %s  %s\n",
		p.dim("PIXALPEEK // SCAN"),
		p.dim("SRC:"+shortPath(res.SourceFile)),
		p.dim(res.ScannedAt.Format("2006-01-02T15:04:05Z07:00")),
		p.red(fmt.Sprintf("RESULTS:%d", len(res.Results))))
	fmt.Fprintf(w, "%s\n", line)

	for i, r := range res.Results {
		idx := ""
		if len(res.Results) > 1 {
			idx = fmt.Sprintf("[%02d] ", i+1)
		}
		fmt.Fprintf(w, "%s %s %-8s %s %-4s\n",
			p.dim("│"),
			p.bold(idx+"TYPE:"),
			p.white(string(r.ContentType)),
			p.dim("ECC:"),
			p.white(r.ErrorCorrectionLevel))
		fmt.Fprintf(w, "%s %s %s\n", p.dim("│"), p.dim("FMT:"), p.white(r.Format))
		b := r.BoundingBox
		fmt.Fprintf(w, "%s %s (%.0f,%.0f)-(%.0f,%.0f)\n",
			p.dim("│"), p.dim("BOUNDS:"),
			b.TopLeft[0], b.TopLeft[1], b.BottomRight[0], b.BottomRight[1])
		fmt.Fprintf(w, "%s %s\n", p.dim("│"), p.bold("PAYLOAD:"))
		for _, ln := range strings.Split(r.Content, "\n") {
			fmt.Fprintf(w, "%s %s\n", p.dim("│"), p.white(ln))
		}
		if i < len(res.Results)-1 {
			fmt.Fprintf(w, "%s\n", p.dim("│ ······"))
		}
	}
	fmt.Fprintf(w, "%s\n", line)
}

func shortPath(p string) string {
	if len(p) <= 40 {
		return p
	}
	return "…" + p[len(p)-39:]
}

func WriteJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func WriteJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

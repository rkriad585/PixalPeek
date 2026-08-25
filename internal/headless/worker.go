package headless

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/rkriad585/PixalPeek/internal/cli"
	"github.com/rkriad585/PixalPeek/internal/qrengine"
	"github.com/rkriad585/PixalPeek/internal/service"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	taskCamera     = "camera"
	taskClipDecode = "clipdecode"
	taskClipCopy   = "clipcopy"
)

var embeddedAssets fs.FS

func SetAssets(a fs.FS) {
	embeddedAssets = a
}

func Run(cfg *cli.Config) int {
	mode := taskClipDecode
	timeout := 25 * time.Second
	if cfg.Camera {
		mode = taskCamera
		timeout = 120 * time.Second
	}

	payload := runWorkerApp(mode, timeout)
	switch {
	case payload == "":
		fmt.Fprintf(os.Stderr, "Error: %s timed out without detecting a QR code\n", modeLabel(mode))
		return cli.ExitNoQRCodeDetected
	case strings.HasPrefix(payload, "ERR:"):
		fmt.Fprintf(os.Stderr, "Error: %s\n", strings.TrimPrefix(payload, "ERR:"))
		return cli.ExitGeneralError
	}

	res := interpretPayload(payload, mode)

	if cfg.OutputPath == "-" {
		if err := cli.WriteJSON(os.Stdout, res); err != nil {
			return cli.ExitGeneralError
		}
		return cli.ExitSuccess
	}
	if cfg.OutputPath != "" {
		if err := cli.WriteJSONFile(cfg.OutputPath, res); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			return cli.ExitGeneralError
		}
		if !cfg.Quiet {
			fmt.Fprintf(os.Stdout, "Scan result saved to %s\n", cfg.OutputPath)
		}
		return cli.ExitSuccess
	}
	if !res.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", res.Error)
		return cli.ExitNoQRCodeDetected
	}
	cli.PrintScanResult(os.Stdout, res, cfg)
	return cli.ExitSuccess
}

func CopyImageToClipboard(png []byte) int {
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	payload := runWorkerAppWithPayload(taskClipCopy, 25*time.Second, dataURL)
	if payload == "OK" {
		return cli.ExitSuccess
	}
	if strings.HasPrefix(payload, "ERR:") {
		fmt.Fprintf(os.Stderr, "Error: %s\n", strings.TrimPrefix(payload, "ERR:"))
		return cli.ExitGeneralError
	}
	fmt.Fprintln(os.Stderr, "Error: could not access the system clipboard")
	return cli.ExitGeneralError
}

func runWorkerApp(mode string, timeout time.Duration) string {
	return runWorkerAppWithPayload(mode, timeout, "")
}

func runWorkerAppWithPayload(mode string, timeout time.Duration, payload string) string {
	if embeddedAssets == nil {
		return "ERR:assets not initialized"
	}
	svc := service.NewQRService()
	if payload != "" {
		svc.SetWorkerPayload(payload)
	}
	resultCh := make(chan string, 1)
	svc.SetCompletionHandler(func(p string) {
		select {
		case resultCh <- p:
		default:
		}
		go func() {
			time.Sleep(150 * time.Millisecond)
			if app := application.Get(); app != nil {
				app.Quit()
			}
		}()
	})

	app := application.New(application.Options{
		Name:        "PixalPeek",
		Description: "PixalPeek background worker",
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(embeddedAssets),
		},
	})

	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "worker",
		Title:            "PIXALPEEK",
		Width:            480,
		Height:           560,
		MinWidth:         360,
		MinHeight:        400,
		URL:              "/?task=" + mode,
		BackgroundColour: application.NewRGBA(10, 10, 11, 255),
	})
	win.Center()

	time.AfterFunc(timeout, func() {
		select {
		case resultCh <- "":
		default:
		}
		go func() {
			time.Sleep(50 * time.Millisecond)
			if a := application.Get(); a != nil {
				a.Quit()
			}
		}()
	})

	_ = app.Run()

	select {
	case p := <-resultCh:
		return p
	default:
		return ""
	}
}

func modeLabel(mode string) string {
	switch mode {
	case taskCamera:
		return "camera scan"
	case taskClipCopy:
		return "clipboard copy"
	default:
		return "clipboard decode"
	}
}

func interpretPayload(payload string, mode string) qrengine.ScanResponse {
	base := qrengine.ScanResponse{
		ScannedAt: time.Now(),
		Results:   []qrengine.DecodeResult{},
	}
	if strings.HasPrefix(payload, "{") {
		var parsed qrengine.ScanResponse
		if err := json.Unmarshal([]byte(payload), &parsed); err == nil {
			return parsed
		}
	}
	if payload == "" {
		base.Error = qrengine.FailureNoQR
		return base
	}
	base.Success = true
	base.SourceFile = mode
	base.Results = append(base.Results, qrengine.DecodeResult{
		Content:     payload,
		ContentType: qrengine.DetectContentType(payload),
		Format:      "QR_CODE",
	})
	return base
}

func ExtractText(payload string) string {
	res := interpretPayload(payload, "tray")
	if res.Success && len(res.Results) > 0 {
		return res.Results[0].Content
	}
	return ""
}

package main

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rkriad585/PixalPeek/internal/cache"
	"github.com/rkriad585/PixalPeek/internal/cli"
	"github.com/rkriad585/PixalPeek/internal/headless"
	"github.com/rkriad585/PixalPeek/internal/service"
	"github.com/rkriad585/PixalPeek/internal/storage"
	"github.com/rkriad585/PixalPeek/internal/watcher"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

var (
	appInstance *application.App
	splashWin   application.Window
	mainWin     application.Window
	svc         *service.QRService
	appCache    *cache.TTLCache
	dirWatcher  *watcher.DirWatcher
)

func main() {
	if len(os.Args) > 1 {
		attachParentConsole()
		cfg, err := cli.ParseFlags(os.Args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintln(os.Stderr, "Run 'pixalpeek -h' for usage.")
			os.Exit(cli.ExitInvalidArgs)
		}
		headless.SetAssets(assets)
		cli.WorkerRunner = headless.Run
		cli.ClipboardCopier = headless.CopyImageToClipboard
		os.Exit(cli.Execute(cfg))
	}
	runGUI()
}

func runGUI() {
	_ = service.InitStorage()
	defer service.ShutdownStorage()

	appCache = cache.New(5 * time.Minute)

	svc = service.NewQRService()

	var watchErr error
	dirWatcher, watchErr = watcher.New(func(ev watcher.Event) {
		appCache.InvalidatePrefix("file:")
	})
	if watchErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: file watcher init failed: %v\n", watchErr)
	}

	appInstance = application.New(application.Options{
		Name:        "PixalPeek",
		Description: "Dot-matrix QR code scanner and generator",
		Icon:        appIcon,
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		OnShutdown: func() {
			if dirWatcher != nil {
				dirWatcher.Stop()
			}
			if appCache != nil {
				appCache.Stop()
			}
			persistWindow()
			persistConfig()
		},
	})

	svc.SetSaveDialog(func(suggestedName string) (string, error) {
		dlg := appInstance.Dialog.SaveFile().
			SetFilename(suggestedName).
			AddFilter("PNG Image", "*.png").
			AddFilter("JPEG Image", "*.jpg;*.jpeg").
			AddFilter("SVG Image", "*.svg").
			AddFilter("PDF Document", "*.pdf").
			AddFilter("All Files", "*.*").
			CanCreateDirectories(true)
		path, err := dlg.PromptForSingleSelection()
		if err != nil {
			return "", err
		}
		if path == "" {
			return "", fmt.Errorf("save cancelled")
		}
		return path, nil
	})

	cfg, _ := service.LoadConfig()
	winW := cfg.WindowW
	winH := cfg.WindowH
	if winW < 520 {
		winW = 1120
	}
	if winH < 600 {
		winH = 760
	}

	splashWin = appInstance.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "splash",
		Title:            "",
		Width:            360,
		Height:           360,
		Frameless:        true,
		AlwaysOnTop:      true,
		BackgroundColour: application.NewRGBA(10, 10, 11, 255),
		URL:              "/?splash=1",
	})
	splashWin.Center()

	go func() {
		time.Sleep(1800 * time.Millisecond)
		createMainWindow(winW, winH, cfg)
		if splashWin != nil {
			splashWin.Close()
			splashWin = nil
		}
	}()

	if err := appInstance.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to start application:", err.Error())
		os.Exit(cli.ExitGeneralError)
	}
}

func createMainWindow(w, h int, cfg storage.AppConfig) {
	restorePos := cfg.WindowX != 0 || cfg.WindowY != 0

	winOpts := application.WebviewWindowOptions{
		Name:             "main",
		Title:            "PIXALPEEK",
		Width:            w,
		Height:           h,
		MinWidth:         520,
		MinHeight:        600,
		Frameless:        true,
		BackgroundColour: application.NewRGBA(10, 10, 11, 255),
		URL:              "/",
	}
	if restorePos {
		winOpts.X = cfg.WindowX
		winOpts.Y = cfg.WindowY
		winOpts.InitialPosition = application.WindowXY
	}

	mainWin = appInstance.Window.NewWithOptions(winOpts)
	if !restorePos {
		mainWin.Center()
	}

	stopPersist := make(chan struct{})
	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				persistWindow()
			case <-stopPersist:
				persistWindow()
				return
			}
		}
	}()
	appInstance.OnShutdown(func() {
		close(stopPersist)
	})

	setupTray(appInstance, svc)

	if ms := os.Getenv("PIXALPEEK_AUTOTEST_MS"); ms != "" {
		if n, err := strconv.Atoi(ms); err == nil && n > 0 {
			time.AfterFunc(time.Duration(n)*time.Millisecond, func() {
				persistWindow()
				persistConfig()
				appInstance.Quit()
			})
		}
	}
}

func persistWindow() {
	if mainWin == nil {
		return
	}
	x, y := mainWin.Position()
	mw, mh := mainWin.Size()
	if mw < 200 || mh < 200 {
		return
	}
	cfg, err := service.LoadConfig()
	if err != nil {
		return
	}
	cfg.WindowX = x
	cfg.WindowY = y
	cfg.WindowW = mw
	cfg.WindowH = mh
	_ = service.SaveConfig(cfg)
}

func persistConfig() {
	s, _ := service.GetSettings()
	cfg, err := service.LoadConfig()
	if err != nil {
		return
	}
	cfg.Theme = s.Theme
	cfg.DefaultFormat = s.DefaultFormat
	cfg.DefaultECC = s.DefaultECC
	cfg.Language = s.Language
	cfg.Size = s.Size
	cfg.Margin = s.Margin
	cfg.Shape = s.Shape
	cfg.CheckURLSafety = s.CheckURLSafety
	_ = service.SaveConfig(cfg)
}

func setupTray(app *application.App, svc *service.QRService) {
	tray := app.SystemTray.New()
	tray.SetIcon(appIcon)
	tray.SetTooltip("PixalPeek")

	menu := application.NewMenu()
	menu.Add("Open PixalPeek").OnClick(func(*application.Context) {
		if mainWin != nil {
			mainWin.Show()
			mainWin.Focus()
		} else {
			app.Show()
		}
	})
	menu.Add("Scan from clipboard").OnClick(func(*application.Context) {
		scanFromClipboardTray(app, svc)
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(*application.Context) {
		persistWindow()
		persistConfig()
		app.Quit()
	})
	tray.SetMenu(menu)
	tray.OnClick(func() {
		if mainWin != nil {
			mainWin.Show()
			mainWin.Focus()
		}
	})
	tray.Run()
}

func scanFromClipboardTray(app *application.App, svc *service.QRService) {
	resultCh := make(chan string, 1)
	svc.SetCompletionHandler(func(p string) {
		select {
		case resultCh <- p:
		default:
		}
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "tray-worker",
		Title:            "PIXALPEEK",
		Width:            420,
		Height:           480,
		URL:              "/?task=clipdecode",
		BackgroundColour: application.NewRGBA(10, 10, 11, 255),
	})

	go func() {
		content := ""
		select {
		case p := <-resultCh:
			content = headless.ExtractText(p)
		case <-time.After(30 * time.Second):
		}
		if w, ok := app.Window.GetByName("tray-worker"); ok {
			w.Close()
		}
		if content != "" {
			app.Event.Emit("pixalpeek:scan-result", map[string]string{"content": content})
		} else {
			app.Event.Emit("pixalpeek:toast", map[string]string{"message": "No QR code found in clipboard image"})
		}
	}()
}

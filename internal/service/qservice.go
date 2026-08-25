package service

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rkriad585/PixalPeek/internal/qrengine"
	"github.com/rkriad585/PixalPeek/internal/storage"
)

type QRService struct {
	onComplete   func(payload string)
	saveDialogFn func(suggestedName string) (string, error)
}

func NewQRService() *QRService {
	return &QRService{}
}

func (s *QRService) SetCompletionHandler(fn func(payload string)) {
	s.onComplete = fn
}

func (s *QRService) SetSaveDialog(fn func(suggestedName string) (string, error)) {
	s.saveDialogFn = fn
}

func decodeB64Image(imageData string) ([]byte, error) {
	b64 := imageData
	if idx := strings.Index(b64, ";base64,"); idx != -1 {
		b64 = b64[idx+8:]
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 image payload: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty image payload")
	}
	return raw, nil
}

func (s *QRService) DecodeBase64(imageData string, multiScan bool) qrengine.ScanResponse {
	raw, err := decodeB64Image(imageData)
	if err != nil {
		return qrengine.ScanResponse{
			Success: false,
			Error:   err.Error(),
			Results: []qrengine.DecodeResult{},
		}
	}
	return qrengine.DecodeBytes(raw, "memory://image", multiScan)
}

func (s *QRService) DecodeFile(path string, multiScan bool) qrengine.ScanResponse {
	return qrengine.DecodeFile(path, multiScan)
}

func (s *QRService) DetectContentType(content string) string {
	return string(qrengine.DetectContentType(content))
}

func (s *QRService) BuildContent(req qrengine.BuildContentRequest) (string, error) {
	return qrengine.BuildContent(req)
}

func (s *QRService) GenerateBase64(opts qrengine.EncodeOptions) (string, error) {
	data, err := qrengine.Generate(opts)
	if err != nil {
		return "", err
	}
	for _, w := range qrengine.ValidateStyle(opts) {
		_ = w
	}
	return "data:image/" + mimeFor(opts.Format) + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func mimeFor(format string) string {
	switch format {
	case qrengine.FormatJPG:
		return "jpeg"
	case qrengine.FormatSVG:
		return "svg+xml"
	default:
		return "png"
	}
}

func (s *QRService) GenerateWithWarnings(opts qrengine.EncodeOptions) (string, []string, error) {
	data, err := qrengine.Generate(opts)
	if err != nil {
		return "", nil, err
	}
	warnings := qrengine.ValidateStyle(opts)
	url := "data:image/" + mimeFor(opts.Format) + ";base64," + base64.StdEncoding.EncodeToString(data)
	return url, warnings, nil
}

func (s *QRService) ValidateStyle(opts qrengine.EncodeOptions) []string {
	return qrengine.ValidateStyle(opts)
}

func (s *QRService) SaveDataURL(dataURL string, suggestedName string) (string, error) {
	if suggestedName == "" {
		suggestedName = "pixalpeek_qr.png"
	}
	target := suggestedName
	if s.saveDialogFn != nil {
		chosen, err := s.saveDialogFn(suggestedName)
		if err != nil {
			return "", fmt.Errorf("save cancelled or failed: %w", err)
		}
		if chosen == "" {
			return "", fmt.Errorf("save cancelled")
		}
		target = chosen
	} else {
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, "Downloads", "PixalPeek")
		if err := os.MkdirAll(dir, 0o755); err == nil {
			target = filepath.Join(dir, suggestedName)
		}
	}
	raw, err := decodeB64Image(dataURL)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(target, raw, 0o644); err != nil {
		return "", fmt.Errorf("write failed: %w", err)
	}
	return target, nil
}

func (s *QRService) ListHistory() ([]HistoryEntry, error) {
	return ListHistory()
}

func (s *QRService) AddHistory(entry HistoryEntry) (HistoryEntry, error) {
	if entry.Kind == KindGenerate && entry.Style != nil {
		entry.Content = strings.TrimSpace(entry.Content)
	}
	return AddHistoryEntry(entry)
}

func (s *QRService) DeleteHistory(id string) error {
	return DeleteHistoryEntry(id)
}

func (s *QRService) ClearHistory(kind string) error {
	return ClearHistory(Kind(kind))
}

func (s *QRService) SetPinned(id string, pinned bool) error {
	return SetPinned(id, pinned)
}

func (s *QRService) GetSettings() (Settings, error) {
	return GetSettings()
}

func (s *QRService) SaveSettings(settings Settings) error {
	return SaveSettings(settings)
}

func (s *QRService) UpsertPreset(preset Preset) ([]Preset, error) {
	return UpsertPreset(preset)
}

func (s *QRService) DeletePreset(name string) ([]Preset, error) {
	return DeletePreset(name)
}

func (s *QRService) ListPresets() ([]Preset, error) {
	settings, err := GetSettings()
	if err != nil {
		return nil, err
	}
	return settings.Presets, nil
}

func (s *QRService) RegenerateFromHistory(id string) (string, error) {
	entry, ok, err := FindEntry(id)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("entry not found: %s", id)
	}
	if entry.Kind != KindGenerate || entry.Style == nil {
		return "", fmt.Errorf("entry is not a generated code")
	}
	opts := qrengine.EncodeOptions{
		Content:   entry.Content,
		Size:      entry.Style.Size,
		ECC:       entry.Style.ECC,
		FGColor:   entry.Style.FGColor,
		BGColor:   entry.Style.BGColor,
		Shape:     entry.Style.Shape,
		LogoB64:   entry.Style.LogoB64,
		Format:    entry.Style.Format,
		QuietZone: entry.Style.Margin,
	}
	return s.GenerateBase64(opts)
}

func (s *QRService) SetWorkerPayload(payload string) {
	workerPayload = payload
}

func (s *QRService) GetWorkerPayload() string {
	return workerPayload
}

func (s *QRService) CompleteTask(result string) {
	if s.onComplete != nil {
		s.onComplete(result)
	}
}

var workerPayload string

func (s *QRService) AppVersion() string {
	return "0.1.5-beta"
}

func (s *QRService) AssessURLSecurity(rawURL string) qrengine.SafetyAssessment {
	return qrengine.AssessURLSecurity(rawURL)
}

func (s *QRService) ScanDirectory(dirPath string, multi bool) (qrengine.BatchScanResult, error) {
	return qrengine.DecodeDirectory(dirPath, multi)
}

func (s *QRService) OpenURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func LoadConfig() (storage.AppConfig, error) {
	return storage.LoadConfig()
}

func SaveConfig(cfg storage.AppConfig) error {
	return storage.SaveConfig(cfg)
}

func (s *QRService) ProcessBatchGUI(batchPath string, opts qrengine.EncodeOptions) (string, int, error) {
	entries, err := qrengine.ParseBatchFile(batchPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse batch file: %w", err)
	}
	if len(entries) == 0 {
		return "", 0, fmt.Errorf("batch file contains no entries")
	}

	tmpDir, err := os.MkdirTemp("", "pixalpeek_batch_*")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	_, count, err := qrengine.BatchGenerate(entries, opts, tmpDir, true, func(format string, args ...interface{}) {})
	if err != nil {
		return "", 0, fmt.Errorf("batch generation failed: %w", err)
	}

	zipPath := filepath.Join(tmpDir, "batch.zip")
	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read batch zip: %w", err)
	}

	return base64.StdEncoding.EncodeToString(zipData), count, nil
}

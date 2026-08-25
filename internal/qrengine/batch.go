package qrengine

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BatchEntry struct {
	Name    string      `json:"name"`
	Content string      `json:"content"`
	Type    ContentType `json:"type,omitempty"`
}

func ParseBatchFile(path string) ([]BatchEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("batch file unreadable: %w", err)
	}
	trimmed := strings.TrimSpace(strings.TrimPrefix(string(data), "\xEF\xBB\xBF"))

	if strings.HasSuffix(strings.ToLower(path), ".json") || strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		return parseBatchJSON([]byte(trimmed))
	}
	return parseBatchCSV([]byte(trimmed))
}

func parseBatchJSON(data []byte) ([]BatchEntry, error) {
	var entries []BatchEntry
	if err := json.Unmarshal(data, &entries); err == nil {
		return entries, nil
	}
	var wrapped struct {
		Entries []BatchEntry `json:"entries"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, fmt.Errorf("invalid batch JSON: %w", err)
	}
	return wrapped.Entries, nil
}

func parseBatchCSV(data []byte) ([]BatchEntry, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid batch CSV: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("batch CSV is empty")
	}

	start := 0
	headerCols := map[string]int{}
	for j, cell := range records[0] {
		headerCols[strings.ToLower(strings.TrimSpace(cell))] = j
	}
	isHeader := false
	for _, known := range []string{"content", "url", "name", "type"} {
		if j, ok := headerCols[known]; ok {
			isHeader = true
			if known == "url" {
				headerCols["content"] = j
			}
		}
	}
	if !isHeader {
		headerCols["content"] = 0
	} else {
		start = 1
	}

	var entries []BatchEntry
	for i := start; i < len(records); i++ {
		row := records[i]
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}
		pick := func(col string) string {
			j, ok := headerCols[col]
			if !ok || j >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[j])
		}
		entry := BatchEntry{Content: pick("content"), Name: pick("name")}
		if t := pick("type"); t != "" {
			entry.Type = ContentType(t)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func sanitizeFilename(s string) string {
	replacer := strings.NewReplacer(
		"\\", "_", "/", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_", "\n", "_", "\r", "_",
	)
	s = strings.TrimSpace(replacer.Replace(s))
	if s == "" {
		return ""
	}
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

func BatchGenerate(entries []BatchEntry, base EncodeOptions, outDir string, makeZip bool, logf func(format string, args ...interface{})) ([]string, int, error) {
	if outDir == "" {
		outDir = "out_qr"
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, 0, fmt.Errorf("cannot create output directory: %w", err)
	}

	var written []string
	failures := 0
	for i, e := range entries {
		content := e.Content
		if content == "" {
			logf("SKIP [%d]: empty content\n", i+1)
			failures++
			continue
		}
		opts := base
		opts.Content = content
		if e.Type.Valid() {
			opts.Type = e.Type
		}

		name := sanitizeFilename(e.Name)
		if name == "" {
			name = fmt.Sprintf("qr_%03d", i+1)
		}
		target := filepath.Join(outDir, name+"."+opts.Format)

		data, err := Generate(opts)
		if err != nil {
			logf("FAIL [%d] %q: %v\n", i+1, truncate(content, 40), err)
			failures++
			continue
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			logf("FAIL [%d]: write error: %v\n", i+1, err)
			failures++
			continue
		}
		written = append(written, target)
		logf("OK   [%d/%d] -> %s\n", i+1, len(entries), target)
	}

	if makeZip && len(written) > 0 {
		zipPath := strings.TrimSuffix(outDir, string(os.PathSeparator)) + ".zip"
		if err := zipFiles(zipPath, written); err != nil {
			logf("WARN: zip creation failed: %v\n", err)
		} else {
			logf("ZIP  -> %s\n", zipPath)
		}
	}
	return written, failures, nil
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func zipFiles(zipPath string, files []string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		entry, err := w.Create(filepath.Base(path))
		if err != nil {
			continue
		}
		if _, err := entry.Write(data); err != nil {
			continue
		}
	}
	return nil
}

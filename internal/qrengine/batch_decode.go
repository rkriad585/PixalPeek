package qrengine

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type BatchScanResult struct {
	TotalFilesScanned int            `json:"total_files_scanned"`
	SuccessfulDecodes int            `json:"successful_decodes"`
	DurationMs        int64          `json:"duration_ms"`
	Entries           []ScanResponse `json:"entries"`
}

var supportedImageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".bmp": true, ".gif": true,
}

func DecodeDirectory(dirPath string, multi bool) (BatchScanResult, error) {
	start := time.Now()
	res := BatchScanResult{
		Entries: make([]ScanResponse, 0),
	}

	var filesToProcess []string
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if supportedImageExts[ext] {
			filesToProcess = append(filesToProcess, path)
		}
		return nil
	})
	if err != nil {
		return res, err
	}

	res.TotalFilesScanned = len(filesToProcess)
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8)

	for _, file := range filesToProcess {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(filePath string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			scanRes := DecodeFile(filePath, multi)
			if scanRes.Success {
				mu.Lock()
				res.Entries = append(res.Entries, scanRes)
				res.SuccessfulDecodes++
				mu.Unlock()
			}
		}(file)
	}

	wg.Wait()
	res.DurationMs = time.Since(start).Milliseconds()
	return res, nil
}

package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rkriad585/PixalPeek/internal/cache"
	"github.com/rkriad585/PixalPeek/internal/qrengine"
	"github.com/rkriad585/PixalPeek/internal/storage"
)

func newStateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("id-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

var (
	mu        sync.Mutex
	dataCache *cache.TTLCache
)

func init() {
	dataCache = cache.New(5 * time.Minute)
}

func dataDir() string {
	if d := storage.ConfigDir(); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "neostore", "pixalpeek")
}

func InitStorage() error {
	dir := dataDir()
	storage.SetConfigDir(dir)
	return storage.Init(dir)
}

func ShutdownStorage() {
	storage.Close()
}

func Load() (*State, error) {
	mu.Lock()
	defer mu.Unlock()

	if cached, ok := dataCache.Get("state"); ok {
		st := cached.(*State)
		return st, nil
	}

	cfg, err := storage.LoadConfig()
	if err != nil {
		cfg = storage.DefaultAppConfig()
	}

	history, err := storage.ListHistoryEntries()
	if err != nil {
		history = []storage.DBHistoryEntry{}
	}

	entries := make([]HistoryEntry, 0, len(history))
	for _, h := range history {
		entry := HistoryEntry{
			ID:          h.ID,
			Kind:        Kind(h.Kind),
			Content:     h.Content,
			ContentType: qrengine.ContentType(h.ContentType),
			Pinned:      h.Pinned,
			Source:      h.Source,
			ErrorLevel:  h.ErrorLevel,
		}
		if h.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, h.Timestamp); err == nil {
				entry.Timestamp = t
			}
		}
		if h.Style != "" {
			var s StyleOptions
			if err := json.Unmarshal([]byte(h.Style), &s); err == nil {
				entry.Style = &s
			}
		}
		if h.Meta != "" {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(h.Meta), &m); err == nil {
				entry.Meta = m
			}
		}
		entries = append(entries, entry)
	}

	presets := make([]Preset, 0, len(cfg.Presets))
	for _, p := range cfg.Presets {
		presets = append(presets, Preset{
			Name: p.Name,
			Style: StyleOptions{
				Format:  p.Style.Format,
				ECC:     p.Style.ECC,
				FGColor: p.Style.FGColor,
				BGColor: p.Style.BGColor,
				Shape:   p.Style.Shape,
				Size:    p.Style.Size,
				Margin:  p.Style.Margin,
				LogoB64: p.Style.LogoB64,
			},
		})
	}

	st := &State{
		Version: 1,
		History: entries,
		Settings: Settings{
			Theme:          cfg.Theme,
			DefaultFormat:  cfg.DefaultFormat,
			DefaultECC:     cfg.DefaultECC,
			Language:       cfg.Language,
			Size:           cfg.Size,
			Margin:         cfg.Margin,
			Shape:          cfg.Shape,
			CheckURLSafety: cfg.CheckURLSafety,
			LastWindow: &WindowPos{
				X: cfg.WindowX, Y: cfg.WindowY,
				W: cfg.WindowW, H: cfg.WindowH,
			},
			Presets: presets,
		},
	}
	dataCache.Set("state", st)
	return st, nil
}

func Save(st *State) error {
	mu.Lock()
	defer mu.Unlock()
	dataCache.Set("state", st)
	return saveToConfig(st)
}

func saveToConfig(st *State) error {
	cfg := storage.AppConfig{
		Theme:          st.Settings.Theme,
		DefaultFormat:  st.Settings.DefaultFormat,
		DefaultECC:     st.Settings.DefaultECC,
		Language:       st.Settings.Language,
		Size:           st.Settings.Size,
		Margin:         st.Settings.Margin,
		Shape:          st.Settings.Shape,
		CheckURLSafety: st.Settings.CheckURLSafety,
		LastTab:        "scan",
		Presets:        make([]storage.Preset, 0, len(st.Settings.Presets)),
	}
	if st.Settings.LastWindow != nil {
		cfg.WindowX = st.Settings.LastWindow.X
		cfg.WindowY = st.Settings.LastWindow.Y
		cfg.WindowW = st.Settings.LastWindow.W
		cfg.WindowH = st.Settings.LastWindow.H
	}
	for _, p := range st.Settings.Presets {
		cfg.Presets = append(cfg.Presets, storage.Preset{
			Name: p.Name,
			Style: storage.PresetStyle{
				Format:  p.Style.Format,
				ECC:     p.Style.ECC,
				FGColor: p.Style.FGColor,
				BGColor: p.Style.BGColor,
				Shape:   p.Style.Shape,
				Size:    p.Style.Size,
				Margin:  p.Style.Margin,
				LogoB64: p.Style.LogoB64,
			},
		})
	}
	return storage.SaveConfig(cfg)
}

func AddHistoryEntry(entry HistoryEntry) (HistoryEntry, error) {
	mu.Lock()
	defer mu.Unlock()
	entry.ID = newStateID()
	entry.Timestamp = time.Now().UTC()

	styleJSON := ""
	if entry.Style != nil {
		data, _ := json.Marshal(entry.Style)
		styleJSON = string(data)
	}
	metaJSON := ""
	if entry.Meta != nil {
		data, _ := json.Marshal(entry.Meta)
		metaJSON = string(data)
	}

	dbEntry := storage.DBHistoryEntry{
		ID:          entry.ID,
		Kind:        string(entry.Kind),
		Content:     entry.Content,
		ContentType: string(entry.ContentType),
		Timestamp:   entry.Timestamp.Format(time.RFC3339),
		Pinned:      entry.Pinned,
		Source:      entry.Source,
		ErrorLevel:  entry.ErrorLevel,
		Style:       styleJSON,
		Meta:        metaJSON,
	}
	if err := storage.InsertHistoryEntry(dbEntry); err != nil {
		return entry, fmt.Errorf("insert history: %w", err)
	}
	dataCache.Delete("state")
	return entry, nil
}

func ListHistory() ([]HistoryEntry, error) {
	mu.Lock()
	defer mu.Unlock()
	dbEntries, err := storage.ListHistoryEntries()
	if err != nil {
		return nil, err
	}
	out := make([]HistoryEntry, 0, len(dbEntries))
	for _, h := range dbEntries {
		entry := HistoryEntry{
			ID:          h.ID,
			Kind:        Kind(h.Kind),
			Content:     h.Content,
			ContentType: qrengine.ContentType(h.ContentType),
			Pinned:      h.Pinned,
			Source:      h.Source,
			ErrorLevel:  h.ErrorLevel,
		}
		if h.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, h.Timestamp); err == nil {
				entry.Timestamp = t
			}
		}
		if h.Style != "" {
			var s StyleOptions
			if err := json.Unmarshal([]byte(h.Style), &s); err == nil {
				entry.Style = &s
			}
		}
		if h.Meta != "" {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(h.Meta), &m); err == nil {
				entry.Meta = m
			}
		}
		out = append(out, entry)
	}
	return out, nil
}

func DeleteHistoryEntry(id string) error {
	mu.Lock()
	defer mu.Unlock()
	err := storage.DeleteHistoryEntryByID(id)
	dataCache.Delete("state")
	return err
}

func ClearHistory(kind Kind) error {
	mu.Lock()
	defer mu.Unlock()
	err := storage.ClearHistoryEntries(string(kind))
	dataCache.Delete("state")
	return err
}

func SetPinned(id string, pinned bool) error {
	mu.Lock()
	defer mu.Unlock()
	err := storage.SetHistoryEntryPinned(id, pinned)
	dataCache.Delete("state")
	return err
}

func FindEntry(id string) (HistoryEntry, bool, error) {
	mu.Lock()
	defer mu.Unlock()
	dbEntry, err := storage.FindHistoryEntry(id)
	if err != nil {
		return HistoryEntry{}, false, err
	}
	if dbEntry == nil {
		return HistoryEntry{}, false, nil
	}
	entry := HistoryEntry{
		ID:          dbEntry.ID,
		Kind:        Kind(dbEntry.Kind),
		Content:     dbEntry.Content,
		ContentType: qrengine.ContentType(dbEntry.ContentType),
		Pinned:      dbEntry.Pinned,
		Source:      dbEntry.Source,
		ErrorLevel:  dbEntry.ErrorLevel,
	}
	if dbEntry.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, dbEntry.Timestamp); err == nil {
			entry.Timestamp = t
		}
	}
	if dbEntry.Style != "" {
		var s StyleOptions
		if err := json.Unmarshal([]byte(dbEntry.Style), &s); err == nil {
			entry.Style = &s
		}
	}
	if dbEntry.Meta != "" {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(dbEntry.Meta), &m); err == nil {
			entry.Meta = m
		}
	}
	return entry, true, nil
}

func GetSettings() (Settings, error) {
	st, err := Load()
	if err != nil {
		return DefaultSettings(), err
	}
	return st.Settings, nil
}

func SaveSettings(s Settings) error {
	mu.Lock()
	defer mu.Unlock()
	st, err := Load()
	if err != nil {
		return err
	}
	st.Settings = s
	if st.Settings.Presets == nil {
		st.Settings.Presets = []Preset{}
	}
	dataCache.Set("state", st)
	return saveToConfig(st)
}

func UpsertPreset(p Preset) ([]Preset, error) {
	mu.Lock()
	defer mu.Unlock()
	st, err := Load()
	if err != nil {
		return nil, err
	}
	replaced := false
	for i := range st.Settings.Presets {
		if st.Settings.Presets[i].Name == p.Name {
			st.Settings.Presets[i] = p
			replaced = true
		}
	}
	if !replaced {
		st.Settings.Presets = append(st.Settings.Presets, p)
	}
	dataCache.Set("state", st)
	return st.Settings.Presets, saveToConfig(st)
}

func DeletePreset(name string) ([]Preset, error) {
	mu.Lock()
	defer mu.Unlock()
	st, err := Load()
	if err != nil {
		return nil, err
	}
	out := st.Settings.Presets[:0]
	for _, existing := range st.Settings.Presets {
		if existing.Name != name {
			out = append(out, existing)
		}
	}
	st.Settings.Presets = out
	dataCache.Set("state", st)
	return st.Settings.Presets, saveToConfig(st)
}

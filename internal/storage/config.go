package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

type AppConfig struct {
	Theme          string   `toml:"theme"`
	Language       string   `toml:"language"`
	DefaultFormat  string   `toml:"default_format"`
	DefaultECC     string   `toml:"default_ecc"`
	Size           int      `toml:"size"`
	Margin         int      `toml:"margin"`
	Shape          string   `toml:"shape"`
	CheckURLSafety bool     `toml:"check_url_safety"`
	WindowX        int      `toml:"window_x"`
	WindowY        int      `toml:"window_y"`
	WindowW        int      `toml:"window_w"`
	WindowH        int      `toml:"window_h"`
	Maximized      bool     `toml:"maximized"`
	Fullscreen     bool     `toml:"fullscreen"`
	LastTab        string   `toml:"last_tab"`
	LastFolder     string   `toml:"last_folder"`
	LastSort       string   `toml:"last_sort"`
	LastFilter     string   `toml:"last_filter"`
	Presets        []Preset `toml:"presets"`
}

type Preset struct {
	Name  string      `toml:"name"`
	Style PresetStyle `toml:"style"`
}

type PresetStyle struct {
	Format  string `toml:"format"`
	ECC     string `toml:"ecc"`
	FGColor string `toml:"fg_color"`
	BGColor string `toml:"bg_color"`
	Shape   string `toml:"shape"`
	Size    int    `toml:"size"`
	Margin  int    `toml:"margin"`
	LogoB64 string `toml:"logo_b64,omitempty"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Theme:          "system",
		Language:       "en",
		DefaultFormat:  "png",
		DefaultECC:     "M",
		Size:           512,
		Margin:         4,
		Shape:          "square",
		CheckURLSafety: true,
		WindowW:        1120,
		WindowH:        760,
		LastTab:        "scan",
		Presets:        []Preset{},
	}
}

var (
	cfgMu  sync.Mutex
	cfgDir string
)

func ConfigDir() string {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	return cfgDir
}

func SetConfigDir(d string) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	cfgDir = d
}

func configPath() string {
	return filepath.Join(cfgDir, "config.toml")
}

func LoadConfig() (AppConfig, error) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	cfg := DefaultAppConfig()
	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

func SaveConfig(cfg AppConfig) error {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return err
	}
	path := configPath()
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config: %w", err)
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}

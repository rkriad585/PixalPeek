package service

import (
	"time"

	"github.com/rkriad585/PixalPeek/internal/qrengine"
)

type Kind string

const (
	KindScan     Kind = "scan"
	KindGenerate Kind = "generate"
)

type StyleOptions struct {
	Format  string `json:"format"`
	ECC     string `json:"ecc"`
	FGColor string `json:"fg_color"`
	BGColor string `json:"bg_color"`
	Shape   string `json:"shape"`
	Size    int    `json:"size"`
	Margin  int    `json:"margin"`
	LogoB64 string `json:"logo_b64,omitempty"`
}

type HistoryEntry struct {
	ID          string                 `json:"id"`
	Kind        Kind                   `json:"kind"`
	Content     string                 `json:"content"`
	ContentType qrengine.ContentType   `json:"content_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Pinned      bool                   `json:"pinned"`
	Source      string                 `json:"source,omitempty"`
	ErrorLevel  string                 `json:"error_level,omitempty"`
	Style       *StyleOptions          `json:"style,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

type WindowPos struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type Preset struct {
	Name  string       `json:"name"`
	Style StyleOptions `json:"style"`
}

type Settings struct {
	Theme          string     `json:"theme"`
	DefaultFormat  string     `json:"default_format"`
	DefaultECC     string     `json:"default_ecc"`
	Language       string     `json:"language"`
	Size           int        `json:"size"`
	Margin         int        `json:"margin"`
	Shape          string     `json:"shape"`
	CheckURLSafety bool       `json:"check_url_safety"`
	LastWindow     *WindowPos `json:"last_window,omitempty"`
	Presets        []Preset   `json:"presets"`
}

func DefaultSettings() Settings {
	return Settings{
		Theme:          "system",
		DefaultFormat:  "png",
		DefaultECC:     "M",
		Language:       "en",
		Size:           512,
		Margin:         4,
		Shape:          "square",
		CheckURLSafety: true,
		Presets:        []Preset{},
	}
}

type State struct {
	Version  int            `json:"version"`
	History  []HistoryEntry `json:"history"`
	Settings Settings       `json:"settings"`
}

func defaultState() *State {
	return &State{
		Version:  1,
		History:  []HistoryEntry{},
		Settings: DefaultSettings(),
	}
}

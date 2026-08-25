package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db  *sql.DB
	mu  sync.Mutex
	dir string
)

func Init(dataDir string) error {
	mu.Lock()
	defer mu.Unlock()
	dir = dataDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	dbPath := filepath.Join(dir, "pixalpeek.db")
	var err error
	db, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	return migrate()
}

func DB() *sql.DB {
	mu.Lock()
	defer mu.Unlock()
	return db
}

func Dir() string {
	mu.Lock()
	defer mu.Unlock()
	return dir
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if db != nil {
		db.Close()
		db = nil
	}
}

func migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			data TEXT NOT NULL DEFAULT '{}',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS history (
			id TEXT PRIMARY KEY,
			kind TEXT NOT NULL CHECK (kind IN ('scan','generate')),
			content TEXT NOT NULL DEFAULT '',
			content_type TEXT NOT NULL DEFAULT 'text',
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			pinned INTEGER NOT NULL DEFAULT 0,
			source TEXT DEFAULT '',
			error_level TEXT DEFAULT '',
			style TEXT DEFAULT '',
			meta TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS history_fts (
			id TEXT NOT NULL,
			content TEXT NOT NULL,
			FOREIGN KEY(id) REFERENCES history(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_history_kind ON history(kind)`,
		`CREATE INDEX IF NOT EXISTS idx_history_pinned ON history(pinned DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history(timestamp DESC)`,
		`CREATE TABLE IF NOT EXISTS cache_entries (
			key TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS file_metadata (
			path TEXT PRIMARY KEY,
			mod_time DATETIME NOT NULL,
			size INTEGER NOT NULL,
			thumb_data TEXT DEFAULT '',
			preview_data TEXT DEFAULT '',
			cached_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS last_session (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			state TEXT NOT NULL DEFAULT '{}',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w\nstatement: %s", err, stmt)
		}
	}
	return nil
}

type DBSettings struct {
	Data      string `json:"data"`
	UpdatedAt string `json:"updated_at"`
}

func LoadSettingsJSON() (string, error) {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return "{}", nil
	}
	var data string
	err := db.QueryRow("SELECT data FROM settings WHERE id = 1").Scan(&data)
	if err == sql.ErrNoRows {
		return "{}", nil
	}
	return data, err
}

func SaveSettingsJSON(data string) error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	_, err := db.Exec(
		`INSERT INTO settings (id, data, updated_at) VALUES (1, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at`,
		data, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

type DBHistoryEntry struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	Timestamp   string `json:"timestamp"`
	Pinned      bool   `json:"pinned"`
	Source      string `json:"source"`
	ErrorLevel  string `json:"error_level"`
	Style       string `json:"style"`
	Meta        string `json:"meta"`
}

func ListHistoryEntries() ([]DBHistoryEntry, error) {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil, nil
	}
	rows, err := db.Query(
		`SELECT id, kind, content, content_type, timestamp, pinned, source, error_level, style, meta
		 FROM history ORDER BY pinned DESC, timestamp DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DBHistoryEntry
	for rows.Next() {
		var e DBHistoryEntry
		var pinned int
		if err := rows.Scan(&e.ID, &e.Kind, &e.Content, &e.ContentType, &e.Timestamp, &pinned, &e.Source, &e.ErrorLevel, &e.Style, &e.Meta); err != nil {
			return nil, err
		}
		e.Pinned = pinned == 1
		out = append(out, e)
	}
	return out, rows.Err()
}

func InsertHistoryEntry(e DBHistoryEntry) error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	pinned := 0
	if e.Pinned {
		pinned = 1
	}
	_, err := db.Exec(
		`INSERT INTO history (id, kind, content, content_type, timestamp, pinned, source, error_level, style, meta)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Kind, e.Content, e.ContentType, e.Timestamp, pinned, e.Source, e.ErrorLevel, e.Style, e.Meta,
	)
	return err
}

func DeleteHistoryEntryByID(id string) error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	_, err := db.Exec("DELETE FROM history WHERE id = ?", id)
	return err
}

func ClearHistoryEntries(kind string) error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	if kind == "" || kind == "all" {
		_, err := db.Exec("DELETE FROM history")
		return err
	}
	_, err := db.Exec("DELETE FROM history WHERE kind = ?", kind)
	return err
}

func SetHistoryEntryPinned(id string, pinned bool) error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	v := 0
	if pinned {
		v = 1
	}
	_, err := db.Exec("UPDATE history SET pinned = ? WHERE id = ?", v, id)
	return err
}

func FindHistoryEntry(id string) (*DBHistoryEntry, error) {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil, nil
	}
	var e DBHistoryEntry
	var pinned int
	err := db.QueryRow(
		`SELECT id, kind, content, content_type, timestamp, pinned, source, error_level, style, meta
		 FROM history WHERE id = ?`, id,
	).Scan(&e.ID, &e.Kind, &e.Content, &e.ContentType, &e.Timestamp, &pinned, &e.Source, &e.ErrorLevel, &e.Style, &e.Meta)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.Pinned = pinned == 1
	return &e, nil
}

type SessionState struct {
	LastTab    string `json:"last_tab"`
	WindowX    int    `json:"window_x"`
	WindowY    int    `json:"window_y"`
	WindowW    int    `json:"window_w"`
	WindowH    int    `json:"window_h"`
	LastFolder string `json:"last_folder"`
	LastFilter string `json:"last_filter"`
	LastSort   string `json:"last_sort"`
}

func SaveSession(s SessionState) error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO last_session (id, state, updated_at) VALUES (1, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET state = excluded.state, updated_at = excluded.updated_at`,
		string(data), time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func LoadSession() (*SessionState, error) {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return &SessionState{}, nil
	}
	var data string
	err := db.QueryRow("SELECT state FROM last_session WHERE id = 1").Scan(&data)
	if err == sql.ErrNoRows {
		return &SessionState{}, nil
	}
	if err != nil {
		return nil, err
	}
	var s SessionState
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return &SessionState{}, nil
	}
	return &s, nil
}

type FileMeta struct {
	Path        string `json:"path"`
	ModTime     string `json:"mod_time"`
	Size        int64  `json:"size"`
	ThumbData   string `json:"thumb_data"`
	PreviewData string `json:"preview_data"`
	CachedAt    string `json:"cached_at"`
}

func GetFileMeta(path string) (*FileMeta, error) {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil, nil
	}
	var m FileMeta
	err := db.QueryRow(
		"SELECT path, mod_time, size, thumb_data, preview_data, cached_at FROM file_metadata WHERE path = ?", path,
	).Scan(&m.Path, &m.ModTime, &m.Size, &m.ThumbData, &m.PreviewData, &m.CachedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func SaveFileMeta(m FileMeta) error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	_, err := db.Exec(
		`INSERT INTO file_metadata (path, mod_time, size, thumb_data, preview_data, cached_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(path) DO UPDATE SET
		   mod_time = excluded.mod_time, size = excluded.size,
		   thumb_data = excluded.thumb_data, preview_data = excluded.preview_data,
		   cached_at = excluded.cached_at`,
		m.Path, m.ModTime, m.Size, m.ThumbData, m.PreviewData, m.CachedAt,
	)
	return err
}

func InvalidateFileMeta(path string) error {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		return nil
	}
	_, err := db.Exec("DELETE FROM file_metadata WHERE path = ?", path)
	return err
}

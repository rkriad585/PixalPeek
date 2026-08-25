package watcher

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Path string
	Op   string
	Time time.Time
}

type DirWatcher struct {
	watcher  *fsnotify.Watcher
	mu       sync.Mutex
	watched  map[string]bool
	callback func(Event)
	stopCh   chan struct{}
}

func New(callback func(Event)) (*DirWatcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	dw := &DirWatcher{
		watcher:  fsw,
		watched:  make(map[string]bool),
		callback: callback,
		stopCh:   make(chan struct{}),
	}
	go dw.loop()
	return dw, nil
}

func (dw *DirWatcher) Watch(dir string) error {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if dw.watched[abs] {
		return nil
	}
	if err := dw.watcher.Add(abs); err != nil {
		return err
	}
	dw.watched[abs] = true
	return nil
}

func (dw *DirWatcher) Unwatch(dir string) error {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if !dw.watched[abs] {
		return nil
	}
	if err := dw.watcher.Remove(abs); err != nil {
		return err
	}
	delete(dw.watched, abs)
	return nil
}

func (dw *DirWatcher) Stop() {
	close(dw.stopCh)
	dw.watcher.Close()
}

func (dw *DirWatcher) loop() {
	for {
		select {
		case ev, ok := <-dw.watcher.Events:
			if !ok {
				return
			}
			op := "modify"
			switch {
			case ev.Op&fsnotify.Create != 0:
				op = "create"
			case ev.Op&fsnotify.Remove != 0:
				op = "remove"
			case ev.Op&fsnotify.Rename != 0:
				op = "rename"
			case ev.Op&fsnotify.Write != 0:
				op = "modify"
			}
			if dw.callback != nil {
				dw.callback(Event{
					Path: ev.Name,
					Op:   op,
					Time: time.Now(),
				})
			}
		case err, ok := <-dw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[watcher] error: %v", err)
		case <-dw.stopCh:
			return
		}
	}
}

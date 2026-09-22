package watcher

import (
	"context"
	"lb/internal/backend"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type NodesWatcher struct {
	path     string
	pool     *backend.Pool
	debounce time.Duration
	log      *slog.Logger
}

func NewNodesWatcher(path string, pool *backend.Pool, log *slog.Logger) *NodesWatcher {
	return &NodesWatcher{
		path:     path,
		pool:     pool,
		debounce: 300 * time.Millisecond,
		log:      log,
	}
}

func (w *NodesWatcher) Start(ctx context.Context) error {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer fw.Close()

	abs, err := filepath.Abs(w.path)
	if err != nil {
		return err
	}
	if err := fw.Add(filepath.Dir(abs)); err != nil {
		return err
	}

	timer := time.NewTimer(w.debounce)
	timer.Stop()
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("nodes watcher stopped")
			return nil
		case ev, ok := <-fw.Events:
			if !ok {
				return nil
			}
			if ev.Name != abs || ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}
			timer.Reset(w.debounce)
		case err, ok := <-fw.Errors:
			if !ok {
				return nil
			}
			w.log.Error("nodes watcher error", "error", err)
		case <-timer.C:
			w.reload()
		}
	}
}

func (w *NodesWatcher) reload() {
	backends, err := backend.LoadNodesFromFile(w.path)
	if err != nil {
		w.log.Error("failed to reload nodes, keeping current list", slog.String("path", w.path), slog.String("error", err.Error()))
		return
	}
	added, removed := w.pool.Update(backends)
	w.log.Info("nodes reloaded", slog.Int("total", len(backends)), slog.Int("added", added), slog.Int("removed", removed))
}

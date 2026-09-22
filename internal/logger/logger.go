package logger

import (
	"lb/internal/config"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func New(cfg *config.Config) *slog.Logger {
	var handler slog.Handler
	if cfg.App.Env == "dev" {
		handler = tint.NewTextHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
			AddSource:  true,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	return slog.New(handler)
}

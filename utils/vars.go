package utils

import (
	"log/slog"
	"os"
	"time"
)

var (
	StartTime = time.Now()

	SLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))
)

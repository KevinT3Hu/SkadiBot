package utils

import (
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

var (
	StartTime = time.Now()

	SLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))

	RecvMsgCounter atomic.Uint64
	SendMsgCounter atomic.Uint64
)

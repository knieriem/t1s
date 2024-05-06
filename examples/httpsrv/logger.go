//go:build !logdiscard

package main

import (
	"log/slog"
	"os"
)

func newTextLogger(l slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: l,
	}))
}

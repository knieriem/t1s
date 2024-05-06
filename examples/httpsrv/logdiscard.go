//go:build logdiscard

package main

import (
	"log/slog"

	"github.com/knieriem/t1s/examples/internal/slogutil"
)

const logLevel = slog.LevelError

func newTextLogger(l slog.Level) *slog.Logger {
	return slogutil.Discard
}

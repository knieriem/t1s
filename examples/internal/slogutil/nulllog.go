package slogutil

import (
	"context"
	"log/slog"
)

var Discard = slog.New(discardHandler{})

type discardHandler struct{}

func (h discardHandler) Enabled(context.Context, slog.Level) bool {
	return false
}
func (h discardHandler) Handle(context.Context, slog.Record) error {
	return nil
}
func (h discardHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}
func (h discardHandler) WithGroup(string) slog.Handler {
	return h
}

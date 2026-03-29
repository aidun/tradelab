package logging

import (
	"io"
	"log/slog"
	"os"
)

func NewJSONLogger(component string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).With("component", component)
}

func NewDiscardLogger(component string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})).With("component", component)
}

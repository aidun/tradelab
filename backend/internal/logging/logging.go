package logging

import (
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

func NewJSONLogger(component string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).With("component", component)
}

func NewDiscardLogger(component string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})).With("component", component)
}

var (
	bearerPattern    = regexp.MustCompile(`(?i)bearer\s+[a-z0-9\-\._~\+/=]+`)
	passwordPattern  = regexp.MustCompile(`(?i)(password=)([^&\s]+)`)
	urlSecretPattern = regexp.MustCompile(`:\/\/([^:\s]+):([^@\s]+)@`)
)

func RedactString(value string) string {
	redacted := bearerPattern.ReplaceAllString(value, "Bearer [REDACTED]")
	redacted = passwordPattern.ReplaceAllString(redacted, "${1}[REDACTED]")
	redacted = urlSecretPattern.ReplaceAllString(redacted, "://$1:[REDACTED]@")
	redacted = strings.ReplaceAll(redacted, "Authorization", "Authorization[REDACTED]")
	return redacted
}

func RedactError(err error) string {
	if err == nil {
		return ""
	}

	return RedactString(err.Error())
}

func TruncateID(value string) string {
	if len(value) <= 8 {
		return value
	}

	return value[:8]
}

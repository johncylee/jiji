package jiji

import (
	"log/slog"
)

var Logger *slog.Logger

func init() {
	Logger = slog.With("package", "jiji")
}

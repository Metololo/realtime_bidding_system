package infrastructure

import (
	"log/slog"
	"os"
)

func NewSellerServiceLogger(level slog.Level) *slog.Logger {
	defaultAttrs := []slog.Attr{
		slog.String("service", "seller-service"),
	}

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: level}).WithAttrs(defaultAttrs)
	return slog.New(logHandler)
}

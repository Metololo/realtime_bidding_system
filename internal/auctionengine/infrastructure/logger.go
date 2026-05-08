package infrastructure

import (
	"log/slog"
	"os"
)

func NewAuctionEngineLogger(level slog.Level) *slog.Logger {
	defaultAttrs := []slog.Attr{
		slog.String("service", "auction-engine"),
	}

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: level}).WithAttrs(defaultAttrs)
	return slog.New(logHandler)

}

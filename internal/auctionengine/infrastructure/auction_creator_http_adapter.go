package infrastructure

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/google/uuid"
)

type AuctionCreatorHTTP struct {
	auctionService application.AuctionCreator
}

func NewAuctionCreatorHTTP(auctionService application.AuctionCreator) *AuctionCreatorHTTP {
	return &AuctionCreatorHTTP{
		auctionService: auctionService,
	}
}

type CreateAuctionRequest struct {
	ItemID       uuid.UUID `json:"itemID"`
	ReservePrice int64     `json:"reservePrice"`
}

func (a *AuctionCreatorHTTP) createAuction(w http.ResponseWriter, r *http.Request) {

	logger := slog.Default().With(
		slog.String("handler", "auction_creator_http"),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	logger.Info("request received")

	if r.Method != http.MethodPost {
		logger.Warn("request rejected", slog.String("reason", "method_not_allowed"))

		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("request rejected", slog.String("reason", "invalid_json"), slog.String("err", err.Error()))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctxLogger := logger.With(
		"operation", "create_auction",
		"item_id", req.ItemID,
		"reserve_price", req.ReservePrice,
	)

	auctionResult, err := a.auctionService.CreateAuction(application.CreateAuctionCommand{
		ItemID:       req.ItemID,
		ReservePrice: req.ReservePrice,
	})

	if err != nil {
		status, errMessage := mapErrorToHTTP(err)
		logCreateAuctionError(ctxLogger, status, err)

		http.Error(w, errMessage, status)
		return
	}

	ctxLogger.Info("auction created", slog.String("auction_id", auctionResult.ID.String()))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(auctionResult); err != nil {
		ctxLogger.Error("encode create auction response failed",
			slog.String("auction_id", auctionResult.ID.String()),
			slog.String("err", err.Error()),
		)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (a *AuctionCreatorHTTP) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.createAuction)
	return mux
}

func logCreateAuctionError(logger *slog.Logger, status int, err error) {
	if status < 500 {
		logger.Warn("create auction failed",
			slog.Int("status", status),
			slog.String("err", err.Error()),
		)
	} else {
		logger.Error("create auction failed",
			slog.Int("status", status),
			slog.String("err", err.Error()),
		)
	}
}

func mapErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidReservePrice),
		errors.Is(err, domain.ErrNilItemID),
		errors.Is(err, inmemory.ErrAuctionAlreadyExists),
		errors.Is(err, inmemory.ErrAuctionNotActive):
		return http.StatusBadRequest, err.Error()

	default:
		return http.StatusInternalServerError, "internal error"
	}
}

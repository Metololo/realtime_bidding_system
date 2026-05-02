package infrastructure

import (
	"encoding/json"
	"errors"
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
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	auctionResult, err := a.auctionService.CreateAuction(application.CreateAuctionCommand{
		ItemID:       req.ItemID,
		ReservePrice: req.ReservePrice,
	})

	if err != nil {
		status, errMessage := mapErrorToHTTP(err)
		http.Error(w, errMessage, status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(auctionResult); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (a *AuctionCreatorHTTP) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.createAuction)
	return mux
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

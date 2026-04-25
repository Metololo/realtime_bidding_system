package infrastructure

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/google/uuid"
)

type AuctionCreatorHTTP struct {
	auctionCreator application.AuctionCreator
}

func NewAuctionCreatorHTTP(auctionCreator application.AuctionCreator) *AuctionCreatorHTTP {
	return &AuctionCreatorHTTP{
		auctionCreator: auctionCreator,
	}
}

type CreateAuctionRequest struct {
	ItemID       uuid.UUID `json:"itemID"`
	ReservePrice int64     `json:"reservePrice"`
}

func (a *AuctionCreatorHTTP) createAuction(w http.ResponseWriter, r *http.Request) {
	fmt.Print("testttt")
	auctionResult, err := a.auctionCreator.CreateAuction(application.CreateAuctionCommand{
		ItemID:       uuid.New(),
		ReservePrice: 100,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(auctionResult); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (a *AuctionCreatorHTTP) StartHTTPServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.createAuction)
	return http.ListenAndServe(":8080", mux)
}

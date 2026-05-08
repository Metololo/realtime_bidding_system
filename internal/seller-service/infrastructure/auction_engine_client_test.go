package infrastructure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/seller-service/service"
	"github.com/google/uuid"
)

func TestAuctionEngineClientCreateAuction(t *testing.T) {
	itemID := uuid.New()
	auctionID := uuid.New()
	startTime := time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Minute)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var request createAuctionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if request.ItemID != itemID {
			t.Fatalf("expected item ID %s, got %s", itemID, request.ItemID)
		}

		if request.ReservePrice != 750 {
			t.Fatalf("expected reserve price 750, got %d", request.ReservePrice)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(AuctionCreatedResult{
			ID:           auctionID,
			ItemID:       itemID,
			ReservePrice: 750,
			StartTime:    startTime,
			EndTime:      endTime,
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewAuctionEngineClient(server.URL)
	sellRequest := service.SellRequest{
		DevilFruit: service.DevilFruit{UUID: itemID.String(), Name: "Gum Gum Fruit"},
		Price:      750,
	}

	result, err := client.CreateAuction(context.Background(), sellRequest)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID != auctionID {
		t.Fatalf("expected auction ID %s, got %s", auctionID, result.ID)
	}

	if result.ItemID != itemID {
		t.Fatalf("expected item ID %s, got %s", itemID, result.ItemID)
	}

	if result.ReservePrice != 750 {
		t.Fatalf("expected reserve price 750, got %d", result.ReservePrice)
	}
}

func TestAuctionEngineClientCreateAuctionReturnsErrorWhenItemIDIsInvalid(t *testing.T) {
	client := NewAuctionEngineClient("http://auction-engine")
	sellRequest := service.SellRequest{
		DevilFruit: service.DevilFruit{UUID: "not-a-uuid", Name: "Gum Gum Fruit"},
		Price:      750,
	}

	_, err := client.CreateAuction(context.Background(), sellRequest)

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestAuctionEngineClientCreateAuctionReturnsErrorWhenAuctionEngineRejectsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid reserve price", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewAuctionEngineClient(server.URL)
	sellRequest := service.SellRequest{
		DevilFruit: service.DevilFruit{UUID: uuid.New().String(), Name: "Gum Gum Fruit"},
		Price:      0,
	}

	_, err := client.CreateAuction(context.Background(), sellRequest)

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestAuctionEngineClientSendSellRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(AuctionCreatedResult{
			ID:           uuid.New(),
			ItemID:       uuid.New(),
			ReservePrice: 100,
			StartTime:    time.Now(),
			EndTime:      time.Now().Add(time.Minute),
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewAuctionEngineClient(server.URL)
	sellRequest := service.SellRequest{
		DevilFruit: service.DevilFruit{UUID: uuid.New().String(), Name: "Gum Gum Fruit"},
		Price:      100,
	}

	err := client.SendSellRequest(context.Background(), sellRequest)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

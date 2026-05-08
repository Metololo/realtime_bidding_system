package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/seller-service/service"
	"github.com/google/uuid"
)

type AuctionEngineClient struct {
	baseURL    string
	httpClient *http.Client
}

type AuctionCreatedResult struct {
	ID           uuid.UUID `json:"ID"`
	ItemID       uuid.UUID `json:"ItemID"`
	ReservePrice int64     `json:"ReservePrice"`
	StartTime    time.Time `json:"StartTime"`
	EndTime      time.Time `json:"EndTime"`
}

type createAuctionRequest struct {
	ItemID       uuid.UUID `json:"itemID"`
	ReservePrice int64     `json:"reservePrice"`
}

func NewAuctionEngineClient(baseURL string) *AuctionEngineClient {
	return NewAuctionEngineClientWithHTTPClient(baseURL, http.DefaultClient)
}

func NewAuctionEngineClientWithHTTPClient(baseURL string, httpClient *http.Client) *AuctionEngineClient {
	return &AuctionEngineClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (client *AuctionEngineClient) SendSellRequest(ctx context.Context, sellRequest service.SellRequest) error {
	_, err := client.CreateAuction(ctx, sellRequest)
	return err
}

func (client *AuctionEngineClient) CreateAuction(ctx context.Context, sellRequest service.SellRequest) (AuctionCreatedResult, error) {
	itemID, err := uuid.Parse(sellRequest.DevilFruit.UUID)
	if err != nil {
		return AuctionCreatedResult{}, fmt.Errorf("invalid devil fruit uuid %q: %w", sellRequest.DevilFruit.UUID, err)
	}

	requestBody := createAuctionRequest{
		ItemID:       itemID,
		ReservePrice: int64(sellRequest.Price),
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return AuctionCreatedResult{}, fmt.Errorf("failed to encode create auction request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL, bytes.NewReader(body))
	if err != nil {
		return AuctionCreatedResult{}, fmt.Errorf("failed to build create auction request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return AuctionCreatedResult{}, fmt.Errorf("failed to create auction: %w", err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			slog.DebugContext(ctx, "failed to close response body",
				slog.Any("body", requestBody))
		}
	}()

	if response.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(response.Body)
		return AuctionCreatedResult{}, fmt.Errorf("auction engine create auction failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var result AuctionCreatedResult
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return AuctionCreatedResult{}, fmt.Errorf("failed to decode create auction response: %w", err)
	}

	return result, nil
}

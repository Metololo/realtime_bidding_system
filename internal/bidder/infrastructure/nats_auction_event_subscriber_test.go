package infrastructure

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/bidder/application"
	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/Metololo/realtime_bidding_system/internal/bidder/strategies"
	"github.com/google/uuid"
)

func TestNATSAuctionEventSubscriber_HandleAuctionCreatedPayload_PlacesBid(t *testing.T) {
	bidderID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	auctionID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	auctionEngine := &recordingAuctionEngine{}
	strategy, err := strategies.NewBalancedBidStrategy(bidderID)
	if err != nil {
		t.Fatalf("failed to create strategy: %v", err)
	}
	var out bytes.Buffer
	subscriber := NewNATSAuctionEventSubscriber(
		"nats://example:4222",
		"",
		bidderID,
		application.NewBidPlacer(strategy, auctionEngine),
		&out,
	)

	err = subscriber.HandleAuctionCreatedPayload(mustJSON(t, auctionCreatedEventDTO{
		ID:           uuid.New(),
		At:           time.Now(),
		AuctionID:    auctionID,
		ItemID:       uuid.New(),
		ReservePrice: 100,
		StartedAt:    time.Now(),
		EndAt:        time.Now().Add(time.Second),
	}))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if auctionEngine.received.AuctionID != auctionID {
		t.Fatalf("expected bid on auction %s, got %s", auctionID, auctionEngine.received.AuctionID)
	}
	if auctionEngine.received.BidderID != bidderID {
		t.Fatalf("expected bidder %s, got %s", bidderID, auctionEngine.received.BidderID)
	}
	if auctionEngine.received.Amount < 110 || auctionEngine.received.Amount > 165 {
		t.Fatalf("expected balanced strategy amount in [110, 165], got %d", auctionEngine.received.Amount)
	}
	if !strings.Contains(out.String(), "placed bid") {
		t.Fatalf("expected placed bid output, got %q", out.String())
	}
}

func TestNATSAuctionEventSubscriber_HandleAuctionCreatedPayload_ReturnsBidError(t *testing.T) {
	bidderID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	expectedErr := errors.New("auction unavailable")
	auctionEngine := &recordingAuctionEngine{err: expectedErr}
	strategy, err := strategies.NewConservativeBidStrategy(bidderID)
	if err != nil {
		t.Fatalf("failed to create strategy: %v", err)
	}
	subscriber := NewNATSAuctionEventSubscriber(
		"nats://example:4222",
		"",
		bidderID,
		application.NewBidPlacer(strategy, auctionEngine),
		ioDiscard{},
	)

	err = subscriber.HandleAuctionCreatedPayload(mustJSON(t, auctionCreatedEventDTO{
		ID:           uuid.New(),
		At:           time.Now(),
		AuctionID:    uuid.New(),
		ItemID:       uuid.New(),
		ReservePrice: 100,
		StartedAt:    time.Now(),
		EndAt:        time.Now().Add(time.Second),
	}))

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped error %v, got %v", expectedErr, err)
	}
}

func TestNATSAuctionEventSubscriber_HandleAuctionClosedPayload_PrintsWinLoss(t *testing.T) {
	bidderID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	otherBidderID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	auctionID := uuid.New()

	tests := []struct {
		name     string
		winner   *winnerInfoDTO
		expected string
	}{
		{
			name: "won",
			winner: &winnerInfoDTO{
				BidderID: bidderID,
				Amount:   150,
			},
			expected: "WON",
		},
		{
			name: "lost",
			winner: &winnerInfoDTO{
				BidderID: otherBidderID,
				Amount:   175,
			},
			expected: "did not win",
		},
		{
			name:     "no bids",
			winner:   nil,
			expected: "no bids",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			subscriber := NewNATSAuctionEventSubscriber("nats://example:4222", "", bidderID, nil, &out)

			err := subscriber.HandleAuctionClosedPayload(mustJSON(t, auctionClosedEventDTO{
				ID:        uuid.New(),
				At:        time.Now(),
				AuctionID: auctionID,
				ItemID:    uuid.New(),
				Outcome:   "sold",
				ClosedAt:  time.Now(),
				Winner:    tt.winner,
			}))

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !strings.Contains(out.String(), tt.expected) {
				t.Fatalf("expected output to contain %q, got %q", tt.expected, out.String())
			}
		})
	}
}

type recordingAuctionEngine struct {
	received domain.BidProposal
	err      error
}

func (e *recordingAuctionEngine) SubmitBid(bidProposal domain.BidProposal) (domain.BidProposal, error) {
	e.received = bidProposal
	if e.err != nil {
		return domain.BidProposal{}, e.err
	}
	return bidProposal, nil
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return payload
}

package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	"github.com/google/uuid"
)

func TestAuctionCreatorHttp_CreateAuction(t *testing.T) {
	server := newTestAuctionServer(t)
	defer server.Close()

	itemID := uuid.New()

	body := `{
		"itemID": "` + itemID.String() + `",
		"reservePrice": 100
	}`

	resp, err := http.Post(
		server.URL,
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer closeBody(t, resp.Body)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var result application.AuctionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.ItemID != itemID {
		t.Errorf("expected ItemID %v, got %v", itemID, result.ItemID)
	}

	if result.ReservePrice != 100 {
		t.Errorf("expected ReservePrice 100, got %d", result.ReservePrice)
	}

	if result.ID == uuid.Nil {
		t.Error("expected non-empty auction ID")
	}

	if result.StartTime.IsZero() {
		t.Error("expected StartTime to be set")
	}

	if result.EndTime.IsZero() {
		t.Error("expected EndTime to be set")
	}

	if result.ID == uuid.Nil {
		t.Error("expected id to not be nil")
	}
}

func TestCreateAuction_RejectsNonPOSTMethods(t *testing.T) {
	server := newTestAuctionServer(t)
	defer server.Close()

	methods := []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, server.URL, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer closeBody(t, resp.Body)

			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Fatalf("expected 405, got %d", resp.StatusCode)
			}

			allow := resp.Header.Get("Allow")
			if allow != http.MethodPost {
				t.Errorf("expected Allow header to be POST, got %s", allow)
			}
		})
	}
}

func TestCreateAuction_InvalidReservePrice_ReturnsBadRequest(t *testing.T) {
	server := newTestAuctionServer(t)
	defer server.Close()

	itemID := uuid.New()

	body := `{
		"itemID": "` + itemID.String() + `",
		"reservePrice": 0
	}`

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer closeBody(t, resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateAuction_NilItemID_ReturnsBadRequest(t *testing.T) {
	server := newTestAuctionServer(t)
	defer server.Close()

	body := `{
		"itemID": "00000000-0000-0000-0000-000000000000",
		"reservePrice": 100
	}`

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer closeBody(t, resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func newTestAuctionServer(t *testing.T) *httptest.Server {
	t.Helper()

	activeAuctionManager := inmemory.NewActiveAuctionManager()
	scheduler := &testutils.FakeManualScheduler{}

	auctionService := application.NewAuctionService(
		activeAuctionManager,
		scheduler,
		infrastructure.NewSystemClock(),
		&testutils.FakeEventPublisher{},
	)

	handler := infrastructure.NewAuctionCreatorHTTP(auctionService)

	return httptest.NewServer(handler.Handler())
}

func closeBody(t *testing.T, body io.Closer) {
	t.Helper()
	if err := body.Close(); err != nil {
		t.Errorf("failed to close body: %v", err)
	}
}

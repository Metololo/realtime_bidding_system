package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	auctionpb "github.com/Metololo/realtime_bidding_system/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestBidPlacerGRPC_PlaceBid_Success(t *testing.T) {
	client, cleanup, auctionService, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	auctionID := createAuctionForBid(t, auctionService)
	bidderID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	resp, err := client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: auctionID.String(),
		BidderId:  bidderID.String(),
		Amount:    150,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.AuctionID != auctionID.String() {
		t.Fatalf("expected auction ID %s, got %s", auctionID.String(), resp.AuctionID)
	}

	if resp.BidderId != bidderID.String() {
		t.Fatalf("expected bidder ID %s, got %s", bidderID.String(), resp.BidderId)
	}

	if resp.Amount != 150 {
		t.Fatalf("expected amount 150, got %d", resp.Amount)
	}
}

func TestBidPlacerGRPC_PlaceBid_InvalidAuctionID_ReturnsInvalidArgument(t *testing.T) {
	client, cleanup, _, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	_, err := client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: "not-a-uuid",
		BidderId:  uuid.New().String(),
		Amount:    150,
	})

	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected status %s, got %s", codes.InvalidArgument, status.Code(err))
	}
}

func TestBidPlacerGRPC_PlaceBid_InvalidBidderID_ReturnsInvalidArgument(t *testing.T) {
	client, cleanup, auctionService, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	auctionID := createAuctionForBid(t, auctionService)

	_, err := client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: auctionID.String(),
		BidderId:  "not-a-uuid",
		Amount:    150,
	})

	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected status %s, got %s", codes.InvalidArgument, status.Code(err))
	}
}

func TestBidPlacerGRPC_PlaceBid_NilBidderID_ReturnsInvalidArgument(t *testing.T) {
	client, cleanup, auctionService, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	auctionID := createAuctionForBid(t, auctionService)

	_, err := client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: auctionID.String(),
		BidderId:  uuid.Nil.String(),
		Amount:    150,
	})

	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected status %s, got %s", codes.InvalidArgument, status.Code(err))
	}
}

func TestBidPlacerGRPC_PlaceBid_InvalidAmount_ReturnsInvalidArgument(t *testing.T) {
	client, cleanup, auctionService, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	auctionID := createAuctionForBid(t, auctionService)

	_, err := client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: auctionID.String(),
		BidderId:  uuid.New().String(),
		Amount:    0,
	})

	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected status %s, got %s", codes.InvalidArgument, status.Code(err))
	}
}

func TestBidPlacerGRPC_PlaceBid_UnknownAuction_ReturnsNotFound(t *testing.T) {
	client, cleanup, _, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	_, err := client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: uuid.New().String(),
		BidderId:  uuid.New().String(),
		Amount:    150,
	})

	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected status %s, got %s", codes.NotFound, status.Code(err))
	}
}

func TestBidPlacerGRPC_PlaceBid_ExpiredAuction_ReturnsFailedPrecondition(t *testing.T) {
	client, cleanup, auctionService, clock := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	auctionID := createAuctionForBid(t, auctionService)
	clock.Advance(101 * time.Millisecond)

	_, err := client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: auctionID.String(),
		BidderId:  uuid.New().String(),
		Amount:    150,
	})

	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected status %s, got %s", codes.FailedPrecondition, status.Code(err))
	}
}

func TestBidPlacerGRPC_PlaceBid_BidBelowReserve_ReturnsOutOfRange(t *testing.T) {
	client, cleanup, auctionService, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	auctionID := createAuctionForBid(t, auctionService)

	_, err := client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: auctionID.String(),
		BidderId:  uuid.New().String(),
		Amount:    50,
	})

	if status.Code(err) != codes.OutOfRange {
		t.Fatalf("expected status %s, got %s", codes.OutOfRange, status.Code(err))
	}
}

func TestBidPlacerGRPC_PlaceBid_BidNotHigherThanLeadingBid_ReturnsOutOfRange(t *testing.T) {
	client, cleanup, auctionService, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	auctionID := createAuctionForBid(t, auctionService)
	_, err := auctionService.PlaceBid(application.BidCommand{
		AuctionID: auctionID,
		BidderID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Amount:    150,
	})
	if err != nil {
		t.Fatalf("failed to place leading bid: %v", err)
	}

	_, err = client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: auctionID.String(),
		BidderId:  uuid.MustParse("33333333-3333-3333-3333-333333333333").String(),
		Amount:    150,
	})

	if status.Code(err) != codes.OutOfRange {
		t.Fatalf("expected status %s, got %s", codes.OutOfRange, status.Code(err))
	}
}

func TestBidPlacerGRPC_PlaceBid_BidderAlreadyPlacedBid_ReturnsAlreadyExists(t *testing.T) {
	client, cleanup, auctionService, _ := newTestBidPlacerGRPCClient(t)
	defer cleanup()

	auctionID := createAuctionForBid(t, auctionService)
	bidderID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	_, err := auctionService.PlaceBid(application.BidCommand{
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    150,
	})
	if err != nil {
		t.Fatalf("failed to place first bid: %v", err)
	}

	_, err = client.PlaceBid(context.Background(), &auctionpb.BidRequest{
		AuctionID: auctionID.String(),
		BidderId:  bidderID.String(),
		Amount:    200,
	})

	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected status %s, got %s", codes.AlreadyExists, status.Code(err))
	}
}

func newTestBidPlacerGRPCClient(t *testing.T) (auctionpb.AuctionEngineClient, func(), *application.AuctionService, *testutils.FakeClock) {
	t.Helper()

	clock := testutils.NewFakeClock(time.Now())
	auctionService := application.NewAuctionService(
		inmemory.NewActiveAuctionManager(),
		&testutils.FakeManualScheduler{},
		clock,
		&testutils.FakeEventPublisher{},
	)

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	auctionpb.RegisterAuctionEngineServer(server, infrastructure.NewBidPlacerGRCP(auctionService))

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("gRPC server stopped: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		server.Stop()
		if closeErr := listener.Close(); closeErr != nil {
			t.Errorf("failed to close listener: %v", closeErr)
		}
		t.Fatalf("failed to create gRPC client connection: %v", err)
	}

	cleanup := func() {
		if err := conn.Close(); err != nil {
			t.Errorf("failed to close gRPC connection: %v", err)
		}
		server.Stop()
		if err := listener.Close(); err != nil {
			t.Errorf("failed to close listener: %v", err)
		}
	}

	return auctionpb.NewAuctionEngineClient(conn), cleanup, auctionService, clock
}

func createAuctionForBid(t *testing.T, service *application.AuctionService) uuid.UUID {
	t.Helper()

	result, err := service.CreateAuction(application.CreateAuctionCommand{
		ItemID:       uuid.New(),
		ReservePrice: 100,
	})
	if err != nil {
		t.Fatalf("failed to create auction: %v", err)
	}
	return result.ID
}

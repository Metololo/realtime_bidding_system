package infrastructure

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	auctionpb "github.com/Metololo/realtime_bidding_system/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestAuctionEngineGRPCClient_SubmitBid_Success(t *testing.T) {
	auctionID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	bidderID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	server := &fakeAuctionEngineServer{
		response: &auctionpb.BidAccepted{
			AuctionID: auctionID.String(),
			BidderId:  bidderID.String(),
			Amount:    150,
		},
	}
	client, cleanup := newTestAuctionEngineGRPCClient(t, server)
	defer cleanup()

	result, err := client.SubmitBid(domain.BidProposal{
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    150,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AuctionID != auctionID {
		t.Fatalf("expected auction ID %s, got %s", auctionID, result.AuctionID)
	}
	if result.BidderID != bidderID {
		t.Fatalf("expected bidder ID %s, got %s", bidderID, result.BidderID)
	}
	if result.Amount != 150 {
		t.Fatalf("expected amount 150, got %d", result.Amount)
	}
	if server.lastRequest == nil {
		t.Fatal("expected gRPC server to receive request")
	}
	if server.lastRequest.AuctionID != auctionID.String() || server.lastRequest.BidderId != bidderID.String() || server.lastRequest.Amount != 150 {
		t.Fatalf("unexpected request: %#v", server.lastRequest)
	}
}

func TestAuctionEngineGRPCClient_SubmitBid_ValidatesProposalBeforeCallingServer(t *testing.T) {
	server := &fakeAuctionEngineServer{}
	client, cleanup := newTestAuctionEngineGRPCClient(t, server)
	defer cleanup()

	_, err := client.SubmitBid(domain.BidProposal{
		AuctionID: uuid.Nil,
		BidderID:  uuid.New(),
		Amount:    150,
	})

	if !errors.Is(err, domain.ErrNilAuctionID) {
		t.Fatalf("expected ErrNilAuctionID, got %v", err)
	}
	if server.lastRequest != nil {
		t.Fatalf("expected no gRPC request for invalid proposal, got %#v", server.lastRequest)
	}
}

func TestAuctionEngineGRPCClient_SubmitBid_MapsBidRejectedDetails(t *testing.T) {
	tests := []struct {
		name          string
		rejectionCode auctionpb.BidRejectionCode
		statusCode    codes.Code
		expectedErr   error
	}{
		{
			name:          "auction unavailable",
			rejectionCode: auctionpb.BidRejectionCode_AUCTION_UNAVAILABLE,
			statusCode:    codes.NotFound,
			expectedErr:   domain.ErrAuctionUnavailable,
		},
		{
			name:          "auction ended",
			rejectionCode: auctionpb.BidRejectionCode_AUCTION_ENDED,
			statusCode:    codes.FailedPrecondition,
			expectedErr:   domain.ErrAuctionEnded,
		},
		{
			name:          "bid too low",
			rejectionCode: auctionpb.BidRejectionCode_BID_TOO_LOW,
			statusCode:    codes.OutOfRange,
			expectedErr:   domain.ErrBidTooLow,
		},
		{
			name:          "bidder already placed bid",
			rejectionCode: auctionpb.BidRejectionCode_BIDDER_ALREADY_PLACED_BID,
			statusCode:    codes.AlreadyExists,
			expectedErr:   domain.ErrBidderAlreadyPlacedBid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := status.New(tt.statusCode, tt.name)
			st, err := st.WithDetails(&auctionpb.BidRejected{
				Code:    tt.rejectionCode,
				Message: tt.name,
			})
			if err != nil {
				t.Fatalf("failed to build status details: %v", err)
			}

			client, cleanup := newTestAuctionEngineGRPCClient(t, &fakeAuctionEngineServer{err: st.Err()})
			defer cleanup()

			_, err = client.SubmitBid(validBidProposal())

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestAuctionEngineGRPCClient_SubmitBid_PreservesStatusErrorWithoutBidRejectedDetails(t *testing.T) {
	client, cleanup := newTestAuctionEngineGRPCClient(t, &fakeAuctionEngineServer{
		err: status.Error(codes.Unavailable, "auction engine unavailable"),
	})
	defer cleanup()

	_, err := client.SubmitBid(validBidProposal())

	if status.Code(err) != codes.Unavailable {
		t.Fatalf("expected gRPC status %s, got %s", codes.Unavailable, status.Code(err))
	}
}

func validBidProposal() domain.BidProposal {
	return domain.BidProposal{
		AuctionID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		BidderID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Amount:    150,
	}
}

type fakeAuctionEngineServer struct {
	auctionpb.UnimplementedAuctionEngineServer
	response    *auctionpb.BidAccepted
	err         error
	lastRequest *auctionpb.BidRequest
}

func (s *fakeAuctionEngineServer) PlaceBid(_ context.Context, req *auctionpb.BidRequest) (*auctionpb.BidAccepted, error) {
	s.lastRequest = req
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func newTestAuctionEngineGRPCClient(t *testing.T, auctionServer auctionpb.AuctionEngineServer) (*AuctionEngineGRPCClient, func()) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	auctionpb.RegisterAuctionEngineServer(server, auctionServer)

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("gRPC server stopped: %v", err)
		}
	}()

	client, err := NewAuctionEngineGRPCClient(
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
		t.Fatalf("failed to create gRPC client: %v", err)
	}

	cleanup := func() {
		if err := client.Close(); err != nil {
			t.Errorf("failed to close client: %v", err)
		}
		server.Stop()
		if err := listener.Close(); err != nil {
			t.Errorf("failed to close listener: %v", err)
		}
	}

	return client, cleanup
}

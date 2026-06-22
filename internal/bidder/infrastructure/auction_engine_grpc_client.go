package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	auctionpb "github.com/Metololo/realtime_bidding_system/proto"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const defaultPlaceBidTimeout = 3 * time.Second

var ErrNilBidAccepted = errors.New("auction engine returned a nil bid accepted response")

type AuctionEngineGRPCClient struct {
	conn    *grpc.ClientConn
	client  auctionpb.AuctionEngineClient
	timeout time.Duration
}

func NewAuctionEngineGRPCClient(target string, opts ...grpc.DialOption) (*AuctionEngineGRPCClient, error) {
	if target == "" {
		return nil, errors.New("auction engine gRPC target is required")
	}

	if len(opts) == 0 {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("create auction engine gRPC client: %w", err)
	}

	return NewAuctionEngineGRPCClientFromConn(conn), nil
}

func NewAuctionEngineGRPCClientFromConn(conn *grpc.ClientConn) *AuctionEngineGRPCClient {
	return &AuctionEngineGRPCClient{
		conn:    conn,
		client:  auctionpb.NewAuctionEngineClient(conn),
		timeout: defaultPlaceBidTimeout,
	}
}

func (c *AuctionEngineGRPCClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *AuctionEngineGRPCClient) SubmitBid(bidProposal domain.BidProposal) (domain.BidProposal, error) {
	if _, err := domain.NewBidProposal(bidProposal.AuctionID, bidProposal.BidderID, bidProposal.Amount); err != nil {
		return domain.BidProposal{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	resp, err := c.client.PlaceBid(ctx, &auctionpb.BidRequest{
		AuctionID: bidProposal.AuctionID.String(),
		BidderId:  bidProposal.BidderID.String(),
		Amount:    bidProposal.Amount,
	})
	if err != nil {
		return domain.BidProposal{}, mapGRPCPlaceBidError(err)
	}

	return mapBidAccepted(resp)
}

func (c *AuctionEngineGRPCClient) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		c.timeout = timeout
	}
}

func mapBidAccepted(resp *auctionpb.BidAccepted) (domain.BidProposal, error) {
	if resp == nil {
		return domain.BidProposal{}, ErrNilBidAccepted
	}

	auctionID, err := parseUUID(resp.AuctionID, domain.ErrNilAuctionID)
	if err != nil {
		return domain.BidProposal{}, err
	}

	bidderID, err := parseUUID(resp.BidderId, domain.ErrNilBidderID)
	if err != nil {
		return domain.BidProposal{}, err
	}

	proposal, err := domain.NewBidProposal(auctionID, bidderID, resp.Amount)
	if err != nil {
		return domain.BidProposal{}, err
	}

	return *proposal, nil
}

func parseUUID(value string, nilErr error) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, err
	}
	if parsed == uuid.Nil {
		return uuid.Nil, nilErr
	}
	return parsed, nil
}

func mapGRPCPlaceBidError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	for _, detail := range st.Details() {
		switch d := detail.(type) {
		case *auctionpb.BidRejected:
			return mapBidRejected(d, st.Err())
		case *errdetails.BadRequest:
			return mapBadRequest(d, st.Err())
		}
	}

	return st.Err()
}

func mapBidRejected(rejection *auctionpb.BidRejected, fallback error) error {
	switch rejection.Code {
	case auctionpb.BidRejectionCode_AUCTION_UNAVAILABLE:
		return fmt.Errorf("%w: %s", domain.ErrAuctionUnavailable, rejection.Message)
	case auctionpb.BidRejectionCode_AUCTION_ENDED:
		return fmt.Errorf("%w: %s", domain.ErrAuctionEnded, rejection.Message)
	case auctionpb.BidRejectionCode_BID_TOO_LOW:
		return fmt.Errorf("%w: %s", domain.ErrBidTooLow, rejection.Message)
	case auctionpb.BidRejectionCode_BIDDER_ALREADY_PLACED_BID:
		return fmt.Errorf("%w: %s", domain.ErrBidderAlreadyPlacedBid, rejection.Message)
	default:
		return fallback
	}
}

func mapBadRequest(badRequest *errdetails.BadRequest, fallback error) error {
	for _, violation := range badRequest.FieldViolations {
		switch violation.Field {
		case "auctionID":
			return fmt.Errorf("%w: %s", domain.ErrNilAuctionID, violation.Description)
		case "bidderId":
			return fmt.Errorf("%w: %s", domain.ErrNilBidderID, violation.Description)
		case "amount":
			return fmt.Errorf("%w: %s", domain.ErrInvalidBidAmount, violation.Description)
		}
	}
	return fallback
}

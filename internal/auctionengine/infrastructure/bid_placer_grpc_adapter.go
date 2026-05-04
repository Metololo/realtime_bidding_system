package infrastructure

import (
	"context"
	"errors"
	"log"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	auctionpb "github.com/Metololo/realtime_bidding_system/proto"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNilBidRequest    = errors.New("bid request is nil")
	ErrInvalidAuctionID = errors.New("auctionID is not a valid UUID")
	ErrInvalidBidderID  = errors.New("bidderID is not a valid UUID")
)

type BidPlacerGRCP struct {
	auctionService application.BidPlacer
	auctionpb.UnimplementedAuctionEngineServer
}

func NewBidPlacerGRCP(auctionService application.BidPlacer) *BidPlacerGRCP {
	return &BidPlacerGRCP{
		auctionService: auctionService,
	}
}

func (b *BidPlacerGRCP) PlaceBid(
	ctx context.Context,
	req *auctionpb.BidRequest,
) (*auctionpb.BidAccepted, error) {
	log.Println("trying to place bid")
	cmd, err := mapToCommand(req)
	if err != nil {
		return nil, mapError(err)
	}

	result, err := b.auctionService.PlaceBid(cmd)
	if err != nil {
		return nil, mapError(err)
	}

	return mapResult(result), nil
}

func mapToCommand(req *auctionpb.BidRequest) (application.BidCommand, error) {
	if req == nil {
		return application.BidCommand{}, ErrNilBidRequest
	}

	auctionID, err := uuid.Parse(req.AuctionID)
	if err != nil {
		return application.BidCommand{}, ErrInvalidAuctionID
	}

	bidderID, err := uuid.Parse(req.BidderId)
	if err != nil {
		return application.BidCommand{}, ErrInvalidBidderID
	}

	return application.BidCommand{
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    req.Amount,
	}, nil
}

func mapResult(result *application.BidResult) *auctionpb.BidAccepted {
	if result == nil {
		return &auctionpb.BidAccepted{}
	}

	return &auctionpb.BidAccepted{
		AuctionID: result.AuctionID.String(),
		BidderId:  result.BidderID.String(),
		Amount:    result.Amount,
	}
}

func mapError(err error) error {
	if requestError := mapRequestError(err); requestError != nil {
		return requestError
	}

	if bidRejectionError := mapBidRejectionError(err); bidRejectionError != nil {
		return bidRejectionError
	}

	return status.Error(codes.Internal, "internal server error")
}

func mapRequestError(err error) error {
	switch {
	case errors.Is(err, ErrNilBidRequest):
		return invalidArgumentError("bid request is required", "request", "request must not be nil")

	case errors.Is(err, ErrInvalidAuctionID):
		return invalidArgumentError("invalid auctionID", "auctionID", "must be a valid UUID")

	case errors.Is(err, ErrInvalidBidderID):
		return invalidArgumentError("invalid bidderId", "bidderId", "must be a valid UUID")

	case errors.Is(err, domain.ErrNilBidderID):
		return invalidArgumentError("invalid bidderId", "bidderId", "must not be nil")

	case errors.Is(err, domain.ErrInvalidBidAmount):
		return invalidArgumentError("invalid bid amount", "amount", "must be greater than zero")

	default:
		return nil
	}
}

func mapBidRejectionError(err error) error {
	switch {
	case errors.Is(err, inmemory.ErrAuctionNotActive):
		return bidRejectedError(
			codes.NotFound,
			"auction unavailable",
			auctionpb.BidRejectionCode_AUCTION_UNAVAILABLE,
			"auction is not active",
		)

	case errors.Is(err, inmemory.ErrAuctionClosing),
		errors.Is(err, domain.ErrAuctionIsClosed),
		errors.Is(err, domain.ErrAuctionIsExpired):
		return bidRejectedError(
			codes.FailedPrecondition,
			"auction ended",
			auctionpb.BidRejectionCode_AUCTION_ENDED,
			"auction is closed",
		)

	case errors.Is(err, domain.ErrBidderAlreadyPlacedBid):
		return bidRejectedError(
			codes.AlreadyExists,
			"bidder already placed a bid",
			auctionpb.BidRejectionCode_BIDDER_ALREADY_PLACED_BID,
			"bidder has already placed a bid on this auction",
		)

	case errors.Is(err, domain.ErrAmountLowerThanReservePrice):
		return bidRejectedError(
			codes.OutOfRange,
			"bid below reserve",
			auctionpb.BidRejectionCode_BID_TOO_LOW,
			"bid amount is lower than the auction's reserve price",
		)

	case errors.Is(err, domain.ErrAmountNotHigherThanHighestBid):
		return bidRejectedError(
			codes.OutOfRange,
			"bid too low",
			auctionpb.BidRejectionCode_BID_TOO_LOW,
			"bid must be higher than the current highest bid",
		)

	default:
		return nil
	}
}

func invalidArgumentError(message, field, description string) error {
	st := status.New(codes.InvalidArgument, message)
	st, err := st.WithDetails(&errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{
				Field:       field,
				Description: description,
			},
		},
	})
	if err != nil {
		return status.Error(codes.InvalidArgument, message)
	}
	return st.Err()
}

func bidRejectedError(code codes.Code, message string, rejectionCode auctionpb.BidRejectionCode, rejectionMessage string) error {
	st := status.New(code, message)
	st, err := st.WithDetails(&auctionpb.BidRejected{
		Code:    rejectionCode,
		Message: rejectionMessage,
	})
	if err != nil {
		return status.Error(code, message)
	}
	return st.Err()
}

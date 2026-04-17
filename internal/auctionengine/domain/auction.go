package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuctionStatus string

const (
	StatusOpen      AuctionStatus = "OPEN"
	StatusClosed    AuctionStatus = "CLOSED"
	AuctionDuration               = 100 * time.Millisecond
)

var ErrInvalidReservePrice = errors.New("reserve price should be > 0")
var ErrInvalidItemID = errors.New("itemID is nil")
var ErrAuctionIsClosed = errors.New("auction is already closed")
var ErrAmountLowerThanReservePrice = errors.New("bid amount is lower than reserve price")
var ErrAuctionIsExpired = errors.New("auction is expired")
var ErrAmountLowerThanHighestBid = errors.New("bid amount is lower than the highest auction bid amount")

type Auction struct {
	id           uuid.UUID
	itemID       uuid.UUID
	reservePrice int64
	startAt      time.Time
	endAt        time.Time
	status       AuctionStatus
	leadingBid   *Bid
}

func NewAuction(itemID uuid.UUID, reservePrice int64) (*Auction, error) {
	if reservePrice <= 0 {
		return nil, ErrInvalidReservePrice
	}

	if itemID == uuid.Nil {
		return nil, ErrInvalidItemID
	}

	startAt := time.Now()
	endAt := startAt.Add(AuctionDuration)

	return &Auction{
		id:           uuid.New(),
		itemID:       itemID,
		reservePrice: reservePrice,
		startAt:      startAt,
		endAt:        endAt,
		status:       StatusOpen,
		leadingBid:   nil,
	}, nil
}

func (a *Auction) Close() error {
	if a.status == StatusClosed {
		return ErrAuctionIsClosed
	}

	a.status = StatusClosed
	return nil
}

func (auction *Auction) PlaceBid(bidderID uuid.UUID, amount int64) (*Bid, error) {

	if auction.isClosed() {
		return nil, ErrAuctionIsClosed
	}

	if auction.isExpired() {
		return nil, ErrAuctionIsExpired
	}

	bid, err := NewBid(bidderID, amount)

	if err != nil {
		return nil, err
	}

	if bid.amount < auction.reservePrice {
		return nil, ErrAmountLowerThanReservePrice
	}

	if auction.leadingBid != nil && bid.amount <= auction.leadingBid.amount {
		return nil, ErrAmountLowerThanHighestBid
	}

	auction.leadingBid = bid
	return bid, nil
}

func (a *Auction) isExpired() bool {
	return time.Now().After(a.endAt)
}

func (a *Auction) isClosed() bool {
	return a.status == StatusClosed
}

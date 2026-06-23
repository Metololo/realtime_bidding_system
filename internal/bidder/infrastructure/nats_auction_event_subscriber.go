package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/bidder/application"
	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	AuctionCreatedSubject = "auction.created"
	AuctionClosedSubject  = "auction.closed"
)

type NATSAuctionEventSubscriber struct {
	serverURL  string
	queueGroup string
	bidderID   uuid.UUID
	bidPlacer  *application.BidPlacer
	out        io.Writer
	conn       *nats.Conn
}

func NewNATSAuctionEventSubscriber(serverURL, queueGroup string, bidderID uuid.UUID, bidPlacer *application.BidPlacer, out io.Writer) *NATSAuctionEventSubscriber {
	if out == nil {
		out = io.Discard
	}
	return &NATSAuctionEventSubscriber{
		serverURL:  serverURL,
		queueGroup: queueGroup,
		bidderID:   bidderID,
		bidPlacer:  bidPlacer,
		out:        out,
	}
}

func (s *NATSAuctionEventSubscriber) Start(ctx context.Context) error {
	if s.serverURL == "" {
		return errors.New("nats server URL is required")
	}
	if s.bidderID == uuid.Nil {
		return domain.ErrNilBidderID
	}
	if s.bidPlacer == nil {
		return errors.New("bid placer is required")
	}

	conn, err := nats.Connect(
		s.serverURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(60),
		nats.ReconnectWait(time.Second),
	)
	if err != nil {
		return fmt.Errorf("connect to nats: %w", err)
	}
	s.conn = conn
	defer conn.Close()

	if err := s.subscribe(AuctionCreatedSubject, s.handleAuctionCreatedMessage); err != nil {
		return err
	}
	if err := s.subscribe(AuctionClosedSubject, s.handleAuctionClosedMessage); err != nil {
		return err
	}

	slog.Info("bidder subscribed to auction events",
		slog.String("nats_url", s.serverURL),
		slog.String("queue_group", s.queueGroup),
		slog.String("bidder_id", s.bidderID.String()),
	)

	<-ctx.Done()
	return nil
}

func (s *NATSAuctionEventSubscriber) subscribe(subject string, handler nats.MsgHandler) error {
	var (
		sub *nats.Subscription
		err error
	)
	if s.queueGroup == "" {
		sub, err = s.conn.Subscribe(subject, handler)
	} else {
		sub, err = s.conn.QueueSubscribe(subject, s.queueGroup, handler)
	}
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", subject, err)
	}
	if err := sub.SetPendingLimits(-1, -1); err != nil {
		return fmt.Errorf("set pending limits for %s: %w", subject, err)
	}
	return nil
}

func (s *NATSAuctionEventSubscriber) handleAuctionCreatedMessage(msg *nats.Msg) {
	if err := s.HandleAuctionCreatedPayload(msg.Data); err != nil {
		slog.Error("failed to handle auction.created event", slog.String("err", err.Error()))
	}
}

func (s *NATSAuctionEventSubscriber) handleAuctionClosedMessage(msg *nats.Msg) {
	if err := s.HandleAuctionClosedPayload(msg.Data); err != nil {
		slog.Error("failed to handle auction.closed event", slog.String("err", err.Error()))
	}
}

func (s *NATSAuctionEventSubscriber) HandleAuctionCreatedPayload(payload []byte) error {
	event, err := decodeAuctionCreatedEvent(payload)
	if err != nil {
		return err
	}

	auction, err := domain.NewAuction(event.AuctionID, event.ItemID, event.ReservePrice, event.StartedAt, event.EndAt)
	if err != nil {
		return fmt.Errorf("map auction.created event: %w", err)
	}

	bid, err := s.bidPlacer.PlaceBid(auction)
	if err != nil {
		return fmt.Errorf("place bid for auction %s: %w", event.AuctionID, err)
	}

	_, _ = fmt.Fprintf(s.out, "bidder %s placed bid %d on auction %s\n", bid.BidderID, bid.Amount, bid.AuctionID)
	return nil
}

func (s *NATSAuctionEventSubscriber) HandleAuctionClosedPayload(payload []byte) error {
	event, err := decodeAuctionClosedEvent(payload)
	if err != nil {
		return err
	}

	if event.Winner == nil {
		_, _ = fmt.Fprintf(s.out, "auction %s closed: no bids, bidder %s did not win\n", event.AuctionID, s.bidderID)
		return nil
	}

	if event.Winner.BidderID == s.bidderID {
		_, _ = fmt.Fprintf(s.out, "auction %s closed: bidder %s WON with amount %d\n", event.AuctionID, s.bidderID, event.Winner.Amount)
		return nil
	}

	_, _ = fmt.Fprintf(s.out, "auction %s closed: bidder %s did not win; winner %s amount %d\n", event.AuctionID, s.bidderID, event.Winner.BidderID, event.Winner.Amount)
	return nil
}

type auctionCreatedEventDTO struct {
	ID           uuid.UUID `json:"ID"`
	At           time.Time `json:"At"`
	AuctionID    uuid.UUID `json:"AuctionID"`
	ItemID       uuid.UUID `json:"ItemID"`
	ReservePrice int64     `json:"ReservePrice"`
	StartedAt    time.Time `json:"StartedAt"`
	EndAt        time.Time `json:"EndAt"`
}

type auctionClosedEventDTO struct {
	ID        uuid.UUID      `json:"ID"`
	At        time.Time      `json:"At"`
	AuctionID uuid.UUID      `json:"AuctionID"`
	ItemID    uuid.UUID      `json:"ItemID"`
	Outcome   string         `json:"Outcome"`
	ClosedAt  time.Time      `json:"ClosedAt"`
	Winner    *winnerInfoDTO `json:"Winner"`
}

type winnerInfoDTO struct {
	BidderID uuid.UUID `json:"BidderID"`
	Amount   int64     `json:"Amount"`
}

func decodeAuctionCreatedEvent(payload []byte) (auctionCreatedEventDTO, error) {
	var event auctionCreatedEventDTO
	if err := json.Unmarshal(payload, &event); err != nil {
		return auctionCreatedEventDTO{}, fmt.Errorf("decode auction.created event: %w", err)
	}
	return event, nil
}

func decodeAuctionClosedEvent(payload []byte) (auctionClosedEventDTO, error) {
	var event auctionClosedEventDTO
	if err := json.Unmarshal(payload, &event); err != nil {
		return auctionClosedEventDTO{}, fmt.Errorf("decode auction.closed event: %w", err)
	}
	return event, nil
}

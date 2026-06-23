package infrastructure

import (
	"encoding/json"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/nats-io/nats.go"
)

type NatsEventPublisher struct {
	ServerURL string
	nats      *nats.Conn
}

func NewNatsEventPublisher(url string) *NatsEventPublisher {
	return &NatsEventPublisher{
		ServerURL: url,
	}
}

func (n *NatsEventPublisher) StartClient() error {
	conn, err := nats.Connect(
		n.ServerURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(60),
		nats.ReconnectWait(time.Second),
	)
	if err != nil {
		return err
	}
	n.nats = conn
	return nil
}

func (n *NatsEventPublisher) Close() {
	if n == nil || n.nats == nil {
		return
	}
	n.nats.Close()
}

func (f *NatsEventPublisher) PublishAuctionClosedEvent(event domain.AuctionClosedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return f.nats.Publish("auction.closed", payload)
}

func (f *NatsEventPublisher) PublishAuctionCreatedEvent(event domain.AuctionCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return f.nats.Publish("auction.created", payload)
}

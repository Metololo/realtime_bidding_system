package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type fakeDevilFruitRepository struct {
	devilFruits []DevilFruit
}

func (repository *fakeDevilFruitRepository) List() []DevilFruit {
	devilFruits := make([]DevilFruit, len(repository.devilFruits))
	copy(devilFruits, repository.devilFruits)
	return devilFruits
}

type fakeSellRequestSender struct {
	mutex        sync.Mutex
	sellRequests []SellRequest
	err          error
}

func (sender *fakeSellRequestSender) SendSellRequest(ctx context.Context, sellRequest SellRequest) error {
	if sender.err != nil {
		return sender.err
	}

	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	sender.sellRequests = append(sender.sellRequests, sellRequest)
	return nil
}

func (sender *fakeSellRequestSender) Count() int {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()
	return len(sender.sellRequests)
}

func disableSlogOutput(t *testing.T) {
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })
}

func TestRunSendsImmediatelyAndThenOnInterval(t *testing.T) {
	disableSlogOutput(t)

	devilFruitRepository := &fakeDevilFruitRepository{devilFruits: []DevilFruit{
		{UUID: "devil-fruit-1", Name: "Gum Gum Fruit"},
	}}
	sellService := NewSellService(devilFruitRepository)
	sender := &fakeSellRequestSender{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		sellService.Run(ctx, sender, time.Millisecond)
		close(done)
	}()

	deadline := time.After(100 * time.Millisecond)
	for sender.Count() < 2 {
		select {
		case <-deadline:
			t.Fatalf("expected at least 2 sell requests, got %d", sender.Count())
		default:
			time.Sleep(time.Millisecond)
		}
	}

	cancel()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected service to stop after context cancellation")
	}
}

func TestCreateRandomSellRequestSelectsDevilFruitAndPrice(t *testing.T) {
	devilFruitRepository := &fakeDevilFruitRepository{devilFruits: []DevilFruit{
		{UUID: "devil-fruit-1", Name: "Gum Gum Fruit"},
		{UUID: "devil-fruit-2", Name: "Chop Chop Fruit"},
	}}
	sellService := NewSellServiceWithRandom(devilFruitRepository, rand.New(rand.NewSource(1)))

	sellRequest, err := sellService.CreateRandomSellRequest()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sellRequest.DevilFruit.UUID == "" {
		t.Fatal("expected sell request devil fruit UUID to be set")
	}

	if sellRequest.DevilFruit.Name == "" {
		t.Fatal("expected sell request devil fruit name to be set")
	}

	if sellRequest.Price < MinimumRandomPrice {
		t.Fatalf("expected price to be at least %d, got %d", MinimumRandomPrice, sellRequest.Price)
	}

	if sellRequest.Price > MaximumRandomPrice {
		t.Fatalf("expected price to be at most %d, got %d", MaximumRandomPrice, sellRequest.Price)
	}
}

func TestCreateRandomSellRequestReturnsErrorWithoutDevilFruits(t *testing.T) {
	devilFruitRepository := &fakeDevilFruitRepository{}
	sellService := NewSellServiceWithRandom(devilFruitRepository, rand.New(rand.NewSource(1)))

	_, err := sellService.CreateRandomSellRequest()

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestSendRandomSellRequestSendsOneRequest(t *testing.T) {
	disableSlogOutput(t)

	devilFruitRepository := &fakeDevilFruitRepository{devilFruits: []DevilFruit{
		{UUID: "devil-fruit-1", Name: "Gum Gum Fruit"},
		{UUID: "devil-fruit-2", Name: "Chop Chop Fruit"},
	}}
	sellService := NewSellService(devilFruitRepository)
	sender := &fakeSellRequestSender{}

	err := sellService.SendRandomSellRequest(context.Background(), sender)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(sender.sellRequests) != 1 {
		t.Fatalf("expected sender to receive 1 sell request, got %d", len(sender.sellRequests))
	}
}

func TestSendRandomSellRequestReturnsSenderError(t *testing.T) {
	disableSlogOutput(t)

	devilFruitRepository := &fakeDevilFruitRepository{devilFruits: []DevilFruit{
		{UUID: "devil-fruit-1", Name: "Gum Gum Fruit"},
	}}
	sellService := NewSellService(devilFruitRepository)
	sender := &fakeSellRequestSender{err: fmt.Errorf("auction engine unavailable")}

	err := sellService.SendRandomSellRequest(context.Background(), sender)

	if err == nil {
		t.Fatal("expected an error")
	}
}

package service

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

const (
	MinimumRandomPrice = 100
	MaximumRandomPrice = 10000
)

type DevilFruit struct {
	UUID string
	Name string
}

type DevilFruitRepository interface {
	List() []DevilFruit
}

type SellRequest struct {
	DevilFruit DevilFruit
	Price      int
}

type SellRequestSender interface {
	SendSellRequest(ctx context.Context, sellRequest SellRequest) error
}

type SellService struct {
	devilFruitRepository DevilFruitRepository
	random               *rand.Rand
	randomMutex          sync.Mutex
}

func NewSellService(devilFruitRepository DevilFruitRepository) *SellService {
	return NewSellServiceWithRandom(devilFruitRepository, rand.New(rand.NewSource(time.Now().UnixNano())))
}

func NewSellServiceWithRandom(devilFruitRepository DevilFruitRepository, random *rand.Rand) *SellService {
	return &SellService{
		devilFruitRepository: devilFruitRepository,
		random:               random,
	}
}

func (service *SellService) Run(ctx context.Context, sender SellRequestSender, interval time.Duration) {
	service.send(ctx, sender)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go service.send(ctx, sender)
		}
	}
}

func (service *SellService) send(ctx context.Context, sender SellRequestSender) {
	if err := service.SendRandomSellRequest(ctx, sender); err != nil {
		slog.ErrorContext(ctx, "sell request failed", slog.String("err", err.Error()))
	}
}

func (service *SellService) SendRandomSellRequest(ctx context.Context, sender SellRequestSender) error {
	sellRequest, err := service.CreateRandomSellRequest()
	if err != nil {
		return err
	}

	if err := sender.SendSellRequest(ctx, sellRequest); err != nil {
		return err
	}

	slog.DebugContext(ctx, "sell request sent",
		slog.String("item_id", sellRequest.DevilFruit.UUID),
		slog.String("item_name", sellRequest.DevilFruit.Name),
		slog.Int("price", sellRequest.Price),
	)
	return nil
}

func (service *SellService) CreateRandomSellRequest() (SellRequest, error) {
	devilFruits := service.devilFruitRepository.List()
	if len(devilFruits) == 0 {
		return SellRequest{}, fmt.Errorf("cannot create sell request without devil fruits")
	}

	service.randomMutex.Lock()
	devilFruit := devilFruits[service.random.Intn(len(devilFruits))]
	price := service.random.Intn(MaximumRandomPrice-MinimumRandomPrice+1) + MinimumRandomPrice
	service.randomMutex.Unlock()

	return SellRequest{
		DevilFruit: devilFruit,
		Price:      price,
	}, nil
}

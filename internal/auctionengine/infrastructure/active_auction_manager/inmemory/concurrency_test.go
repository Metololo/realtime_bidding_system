package inmemory

import (
	"errors"
	"sync"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

func TestActiveAuctionManagerPlaceBidConcurrentSameAuction(t *testing.T) {
	t.Parallel()

	manager := NewActiveAuctionManager()
	auction := newTestAuction(t)

	if err := manager.Save(auction); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	const bidCount = 5000

	var wg sync.WaitGroup
	wg.Add(bidCount)

	start := make(chan struct{})

	var mu sync.Mutex
	var acceptedCount int
	var maxAccepted int64

	for i := 0; i < bidCount; i++ {
		i := i

		go func() {
			defer wg.Done()
			<-start

			amount := int64(i + 1)

			_, err := manager.PlaceBid(auction.ID(), uuid.New(), amount)
			if err != nil {
				return
			}

			mu.Lock()
			acceptedCount++
			if amount > maxAccepted {
				maxAccepted = amount
			}
			mu.Unlock()
		}()
	}

	close(start)
	wg.Wait()

	if acceptedCount == 0 {
		t.Fatal("expected at least one accepted bid")
	}

	winner, err := manager.CloseAuction(auction.ID())
	if err != nil {
		t.Fatalf("CloseAuction() error = %v", err)
	}
	if winner == nil {
		t.Fatal("winner is nil")
	}

	if got, want := winner.Amount(), maxAccepted; got != want {
		t.Fatalf("winner amount = %d, want %d", got, want)
	}
}

func TestActiveAuctionManager_PlaceBid_And_CloseAuction_Concurrent(t *testing.T) {
	t.Parallel()

	manager := NewActiveAuctionManager()
	auction := newTestAuction(t)

	if err := manager.Save(auction); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	_, err := manager.PlaceBid(auction.ID(), uuid.New(), 100)
	if err != nil {
		t.Fatalf("seed PlaceBid() error = %v", err)
	}

	const bidCount = 500

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < bidCount; i++ {
		i := i
		wg.Add(1)

		go func() {
			defer wg.Done()
			<-start

			_, _ = manager.PlaceBid(auction.ID(), uuid.New(), int64(101+i))
		}()
	}

	var (
		closeWinner *domain.Bid
		closeErr    error
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start

		closeWinner, closeErr = manager.CloseAuction(auction.ID())
	}()

	close(start)
	wg.Wait()

	if closeErr != nil {
		t.Fatalf("CloseAuction() error = %v", closeErr)
	}
	if closeWinner == nil {
		t.Fatal("CloseAuction() winner is nil")
	}

	_, err = manager.PlaceBid(auction.ID(), uuid.New(), 999999)
	if !errors.Is(err, ErrAuctionNotActive) {
		t.Fatalf("PlaceBid() after close error = %v, want %v", err, ErrAuctionNotActive)
	}
}

func TestActiveAuctionManager_PlaceBid_Concurrent_MultipleAuctions(t *testing.T) {
	t.Parallel()

	manager := NewActiveAuctionManager()

	const (
		auctionCount       = 50
		bidsPerAuction     = 100
		totalBidGoroutines = auctionCount * bidsPerAuction
	)

	type auctionState struct {
		id           uuid.UUID
		maxAccepted  int64
		acceptedOnce bool
		mu           sync.Mutex
	}

	auctions := make([]*auctionState, 0, auctionCount)

	for i := 0; i < auctionCount; i++ {
		auction := newTestAuction(t)

		if err := manager.Save(auction); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		auctions = append(auctions, &auctionState{id: auction.ID()})
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(totalBidGoroutines)

	for ai := 0; ai < auctionCount; ai++ {
		ai := ai

		for bi := 0; bi < bidsPerAuction; bi++ {
			bi := bi

			go func() {
				defer wg.Done()
				<-start

				amount := int64(bi + 1)

				_, err := manager.PlaceBid(auctions[ai].id, uuid.New(), amount)
				if err != nil {
					return
				}

				st := auctions[ai]
				st.mu.Lock()
				st.acceptedOnce = true
				if amount > st.maxAccepted {
					st.maxAccepted = amount
				}
				st.mu.Unlock()
			}()
		}
	}

	close(start)
	wg.Wait()

	for _, st := range auctions {
		if !st.acceptedOnce {
			t.Fatalf("auction %s had no accepted bids", st.id)
		}

		winner, err := manager.CloseAuction(st.id)
		if err != nil {
			t.Fatalf("CloseAuction(%s) error = %v", st.id, err)
		}
		if winner == nil {
			t.Fatalf("CloseAuction(%s) winner is nil", st.id)
		}

		st.mu.Lock()
		want := st.maxAccepted
		st.mu.Unlock()

		if got := winner.Amount(); got != want {
			t.Fatalf("auction %s winner amount = %d, want %d", st.id, got, want)
		}
	}
}

func TestActiveAuctionManager_MultipleAuctions_BidsAndCloseRace(t *testing.T) {
	t.Parallel()

	manager := NewActiveAuctionManager()

	const (
		auctionCount   = 20
		bidsPerAuction = 100
	)

	type auctionState struct {
		id uuid.UUID
	}

	auctions := make([]auctionState, 0, auctionCount)

	for i := 0; i < auctionCount; i++ {
		auction := newTestAuction(t)

		if err := manager.Save(auction); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		if _, err := manager.PlaceBid(auction.ID(), uuid.New(), 101); err != nil {
			t.Fatalf("seed bid error = %v", err)
		}

		auctions = append(auctions, auctionState{id: auction.ID()})
	}

	start := make(chan struct{})
	var wg sync.WaitGroup

	for _, a := range auctions {
		a := a

		for bi := 0; bi < bidsPerAuction; bi++ {
			bi := bi
			wg.Add(1)

			go func() {
				defer wg.Done()
				<-start

				_, _ = manager.PlaceBid(a.id, uuid.New(), int64(2+bi))
			}()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			_, _ = manager.CloseAuction(a.id)
		}()
	}

	close(start)
	wg.Wait()

	for _, a := range auctions {
		_, err := manager.PlaceBid(a.id, uuid.New(), 999999)
		if !errors.Is(err, ErrAuctionNotActive) {
			t.Fatalf("PlaceBid after close for auction %s: got %v, want %v", a.id, err, ErrAuctionNotActive)
		}
	}
}

func TestActiveAuctionManager_CloseAuction_Concurrent_SameAuction(t *testing.T) {
	t.Parallel()

	manager := NewActiveAuctionManager()
	auction := newTestAuction(t)

	if err := manager.Save(auction); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	_, err := manager.PlaceBid(auction.ID(), uuid.New(), 101)
	if err != nil {
		t.Fatalf("seed bid error = %v", err)
	}

	const closeCalls = 50

	var wg sync.WaitGroup
	wg.Add(closeCalls)

	start := make(chan struct{})
	results := make(chan error, closeCalls)

	for i := 0; i < closeCalls; i++ {
		go func() {
			defer wg.Done()
			<-start

			_, err := manager.CloseAuction(auction.ID())
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var successCount int

	for err := range results {
		if err == nil {
			successCount++
			continue
		}

		if !errors.Is(err, ErrAuctionAlreadyClosing) &&
			!errors.Is(err, ErrAuctionNotActive) {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successCount != 1 {
		t.Fatalf("expected 1 successful close, got %d", successCount)
	}
}

func TestActiveAuctionManager_Save_Concurrent_SameID(t *testing.T) {
	t.Parallel()

	manager := NewActiveAuctionManager()
	auction := newTestAuction(t)

	const attempts = 50

	var wg sync.WaitGroup
	wg.Add(attempts)

	results := make(chan error, attempts)

	for i := 0; i < attempts; i++ {
		go func() {
			defer wg.Done()
			results <- manager.Save(auction)
		}()
	}

	wg.Wait()
	close(results)

	var successCount int

	for err := range results {
		if err == nil {
			successCount++
			continue
		}

		if !errors.Is(err, ErrAuctionAlreadyExists) {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successCount != 1 {
		t.Fatalf("expected 1 success, got %d", successCount)
	}
}

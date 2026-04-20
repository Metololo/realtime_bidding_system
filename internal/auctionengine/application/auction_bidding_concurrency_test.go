package application

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/repository/inmemory"
	"github.com/google/uuid"
)

func TestPlaceBidConcurrently(t *testing.T) {
	const numBidders = 1000

	for range 100 {
		runPlaceBidConcurrentlyOnce(t, numBidders)
	}
}

func runPlaceBidConcurrentlyOnce(t *testing.T, numBidders int) {
	t.Helper()

	auctionRepository := inmemory.NewAuctionRepository()
	auctionService := NewAuctionService(auctionRepository)

	auctionResult, err := auctionService.CreateAuction(newTestCreateAuctionCommand())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var (
		wg                    sync.WaitGroup
		mu                    sync.Mutex
		highestAcceptedAmount int64
		successfulBids        int
	)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	amounts := rng.Perm(numBidders)

	for i := range numBidders {
		wg.Add(1)
		go func(amount int64) {
			defer wg.Done()
			result, err := auctionService.PlaceBid(BidCommand{
				AuctionID: auctionResult.ID,
				BidderID:  uuid.New(),
				Amount:    150 + amount,
			})
			if err != nil {
				return
			}
			mu.Lock()
			successfulBids++
			if result.Amount > highestAcceptedAmount {
				highestAcceptedAmount = result.Amount
			}
			mu.Unlock()
		}(int64(amounts[i]))
	}

	wg.Wait()

	if successfulBids == 0 {
		t.Fatal("expected at least one successful bid, got 0")
	}

	auction, err := auctionRepository.FindByID(auctionResult.ID)
	if err != nil {
		t.Fatalf("expected to find auction, got %v", err)
	}

	if auction.LeadingBid() == nil {
		t.Fatal("expected leading bid to be set, got nil")
	}

	if auction.LeadingBid().Amount() != highestAcceptedAmount {
		t.Fatalf("expected leading bid amount %d, got %d", highestAcceptedAmount, auction.LeadingBid().Amount())
	}
}

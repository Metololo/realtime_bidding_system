package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewBid(t *testing.T) {

	bidderID := uuid.MustParse("1e234536-e29b-41d4-a716-446655440000")
	amount := int64(100)

	bid, err := NewBid(bidderID, amount)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bid == nil {
		t.Fatal("bid is nil")
	}

	if bid.bidderID != bidderID {
		t.Fatalf("expected bidderID to be %v, got %v", bidderID, bid.bidderID)
	}

	if bid.amount != amount {
		t.Fatalf("expected amount to be %v, got %v", amount, bid.amount)
	}
}

func TestBidderIDShouldNotBeNil(t *testing.T) {

	bidderID := uuid.Nil
	amount := int64(100)

	_, err := NewBid(bidderID, amount)

	if err == nil {
		t.Fatal("error is nil")
	}

	if !errors.Is(err, ErrNilBidderID) {
		t.Fatalf("expected error to be ErrNilBidderID, got %v", err)
	}
}

func TestBidAmountCannotBeNegative(t *testing.T) {

	bidderID := uuid.MustParse("123e4566-e29b-41d4-a716-446655440000")
	amount := int64(-2)

	_, err := NewBid(bidderID, amount)

	if err == nil {
		t.Fatal("error is nil")
	}

	if !errors.Is(err, ErrInvalidBidAmount) {
		t.Fatalf("expected error to be ErrInvalidBidAmount, got %v", err)
	}

}

func TestBidAmountCannotBeZero(t *testing.T) {

	bidderID := uuid.MustParse("123e4566-e29b-41d4-a716-446655440000")
	amount := int64(0)

	_, err := NewBid(bidderID, amount)

	if err == nil {
		t.Fatal("error is nil")
	}

	if !errors.Is(err, ErrInvalidBidAmount) {
		t.Fatalf("expected error to be ErrInvalidBidAmount, got %v", err)
	}
}

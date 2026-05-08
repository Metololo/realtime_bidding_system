package main

import (
	"testing"
	"time"
)

func TestIntervalFromEnvAcceptsMilliseconds(t *testing.T) {
	t.Setenv("SELLER_INTERVAL", "1ms")

	interval, err := intervalFromEnv()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if interval != time.Millisecond {
		t.Fatalf("expected 1ms, got %s", interval)
	}
}

func TestIntervalFromEnvAcceptsMicroseconds(t *testing.T) {
	t.Setenv("SELLER_INTERVAL", "100us")

	interval, err := intervalFromEnv()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if interval != 100*time.Microsecond {
		t.Fatalf("expected 100us, got %s", interval)
	}
}

func TestIntervalFromEnvRejectsZero(t *testing.T) {
	t.Setenv("SELLER_INTERVAL", "0s")

	_, err := intervalFromEnv()

	if err == nil {
		t.Fatal("expected an error")
	}
}

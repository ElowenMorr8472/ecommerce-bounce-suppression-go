package main

import (
	"testing"
	"time"
)

func TestRetryDelayHonorsHeaderThenBacksOff(t *testing.T) {
	if got := retryDelay("3", 0); got != 3*time.Second {
		t.Fatalf("Retry-After delay = %s, want 3s", got)
	}
	if got := retryDelay("", 2); got != 4*time.Second {
		t.Fatalf("exponential delay = %s, want 4s", got)
	}
}

func TestIdempotencyKeyIsStableForAddressCasing(t *testing.T) {
	if idempotencyKey("Customer@Example.com") != idempotencyKey("customer@example.com") {
		t.Fatal("expected the same idempotency key for address casing")
	}
}

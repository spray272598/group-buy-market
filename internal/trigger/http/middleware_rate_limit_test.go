package http

import "testing"

func TestRateLimit(t *testing.T) {
	s := NewRateLimitStore(1, 1)
	if !s.allow("u1") {
		t.Fatal("first allow")
	}
	if s.allow("u1") {
		t.Fatal("second should deny at 1qps burst1")
	}
	if !s.allow("u2") {
		t.Fatal("other user allow")
	}
}

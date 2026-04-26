package request

import (
	"code/internal/shared"
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func makeParams(client *http.Client, retries uint) shared.CrawlParams {
	return shared.CrawlParams{
		CTX:        context.Background(),
		HTTPClient: client,
		Retries:    retries,
	}
}

func TestRequest_RetrySucceedsOnSecondAttempt(t *testing.T) {
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	params := makeParams(server.Client(), 1)
	res := Request(params, server.URL)

	if res.Err != nil {
		t.Fatalf("expected success, got error: %v", res.Err)
	}
	if res.Response.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.Response.StatusCode)
	}
	if count.Load() != 2 {
		t.Errorf("expected 2 requests, got %d", count.Load())
	}
}

func TestRequest_RetryExhausted(t *testing.T) {
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	const retries = 2
	params := makeParams(server.Client(), retries)
	res := Request(params, server.URL)

	if res.Response == nil || res.Response.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 response after exhausted retries")
	}
	if count.Load() != retries+1 {
		t.Errorf("expected %d requests, got %d", retries+1, count.Load())
	}
}

func TestRequest_NoRetryOn4xx(t *testing.T) {
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	params := makeParams(server.Client(), 3)
	res := Request(params, server.URL)

	if res.Response == nil || res.Response.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 response without retry")
	}
	if count.Load() != 1 {
		t.Errorf("expected exactly 1 request (no retry on 404), got %d", count.Load())
	}
}

func TestRequest_ContextCancelStopsRetry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	params := shared.CrawlParams{
		CTX:        ctx,
		HTTPClient: server.Client(),
		Retries:    10,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	res := Request(params, server.URL)
	elapsed := time.Since(start)

	if res.Err == nil {
		t.Error("expected error after context cancel")
	}
	if elapsed > 2*time.Second {
		t.Errorf("context cancel did not stop retries in time, elapsed: %v", elapsed)
	}
}

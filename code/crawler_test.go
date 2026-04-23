package code

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testOptions(client *http.Client, url string) Options {
	return Options{
		URL:        url,
		Depth:      1,
		HTTPClient: client,
	}
}

func TestAnalyze_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result, err := Analyze(context.Background(), testOptions(server.Client(), server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report Report
	if err := json.Unmarshal(result, &report); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if report.Pages[0].HttpStatus != http.StatusOK {
		t.Errorf("expected status 200, got %d", report.Pages[0].HttpStatus)
	}
	if report.Pages[0].Status != "ok" {
		t.Errorf("expected status ok, got %s", report.Pages[0].Status)
	}
}

func TestAnalyze_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	result, err := Analyze(context.Background(), testOptions(server.Client(), server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report Report
	if err := json.Unmarshal(result, &report); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if report.Pages[0].HttpStatus != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", report.Pages[0].HttpStatus)
	}
	if report.Pages[0].Status != "error" {
		t.Errorf("expected status error, got %s", report.Pages[0].Status)
	}
}

func TestAnalyze_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	result, err := Analyze(context.Background(), testOptions(server.Client(), server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report Report
	if err := json.Unmarshal(result, &report); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if report.Pages[0].HttpStatus != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", report.Pages[0].HttpStatus)
	}
	if report.Pages[0].Status != "error" {
		t.Errorf("expected status error, got %s", report.Pages[0].Status)
	}
}

func TestAnalyze_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()

	_, err := Analyze(context.Background(), testOptions(server.Client(), url))
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}

func TestAnalyze_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := &http.Client{Timeout: time.Millisecond}
	_, err := Analyze(context.Background(), testOptions(client, server.URL))
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

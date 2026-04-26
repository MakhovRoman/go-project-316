package fetchasset

import (
	"code/internal/shared"
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func makeParams(client *http.Client) shared.CrawlParams {
	return shared.CrawlParams{
		CTX:        context.Background(),
		HTTPClient: client,
		Retries:    0,
		AssetCache: make(shared.AssetCache),
	}
}

func TestFetchAsset_DedupAcrossPages(t *testing.T) {
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count.Add(1)
		w.Header().Set("Content-Length", "5")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	params := makeParams(server.Client())
	asset := shared.Asset{URL: server.URL + "/logo.png", Type: shared.AssetType("image")}

	first := FetchAsset(params, asset)
	second := FetchAsset(params, asset)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 transport call, got %d", got)
	}
	if first != second {
		t.Errorf("expected identical asset on second call, got %+v vs %+v", first, second)
	}
	if first.SizeBytes != 5 {
		t.Errorf("expected size 5, got %d", first.SizeBytes)
	}
	if first.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", first.StatusCode)
	}
}

func TestFetchAsset_MissingContentLength(t *testing.T) {
	body := []byte("some asset body content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body[:5])
		w.(http.Flusher).Flush()
		_, _ = w.Write(body[5:])
	}))
	defer server.Close()

	params := makeParams(server.Client())
	asset := shared.Asset{URL: server.URL + "/asset.js", Type: shared.AssetType("script")}

	res := FetchAsset(params, asset)

	if res.SizeBytes != int64(len(body)) {
		t.Errorf("expected size %d, got %d", len(body), res.SizeBytes)
	}
	if res.Error != "" {
		t.Errorf("expected no error, got %q", res.Error)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
}

func TestFetchAsset_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer server.Close()

	params := makeParams(server.Client())
	asset := shared.Asset{URL: server.URL + "/missing.css", Type: shared.AssetType("style")}

	res := FetchAsset(params, asset)

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}
	if res.Error == "" {
		t.Error("expected non-empty error for 404")
	}
}

func TestFetchAsset_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	url := server.URL + "/asset.png"
	server.Close()

	params := makeParams(server.Client())
	asset := shared.Asset{URL: url, Type: shared.AssetType("image")}

	res := FetchAsset(params, asset)

	if res.StatusCode != 0 {
		t.Errorf("expected status 0 on network error, got %d", res.StatusCode)
	}
	if res.SizeBytes != 0 {
		t.Errorf("expected size 0 on network error, got %d", res.SizeBytes)
	}
	if res.Error == "" {
		t.Error("expected non-empty error for network failure")
	}
}

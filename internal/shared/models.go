package shared

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

type CrawlParams struct {
	CTX        context.Context
	HTTPClient *http.Client
	Host       string
	URL        string
	Depth      uint
	Body       io.Reader
	Visited    *Visited
	Limiter    *RateLimiter
	Delay      time.Duration
	Retries    uint
	RPS        uint
	AssetCache *AssetCache
}

type Visited struct {
	mu    sync.Mutex
	items map[string]struct{}
}

func NewVisited() *Visited {
	return &Visited{items: make(map[string]struct{})}
}

func (v *Visited) MarkIfNew(url string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, ok := v.items[url]; ok {
		return false
	}
	v.items[url] = struct{}{}
	return true
}

type AssetType string

type Asset struct {
	URL        string    `json:"url"`
	Type       AssetType `json:"type"`
	StatusCode int       `json:"status_code"`
	SizeBytes  int64     `json:"size_bytes"`
	Error      string    `json:"error"`
}

type AssetCache struct {
	mu    sync.Mutex
	items map[string]Asset
}

func NewAssetCache() *AssetCache {
	return &AssetCache{items: make(map[string]Asset)}
}

func (c *AssetCache) Get(url string) (Asset, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	a, ok := c.items[url]
	return a, ok
}

func (c *AssetCache) Set(url string, a Asset) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[url] = a
}

type RateLimiter struct {
	ticker *time.Ticker
}

func NewRateLimiter(delay time.Duration) *RateLimiter {
	if delay <= 0 {
		return nil
	}
	return &RateLimiter{ticker: time.NewTicker(delay)}
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	if r == nil {
		return nil
	}
	select {
	case <-r.ticker.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *RateLimiter) Stop() {
	if r != nil && r.ticker != nil {
		r.ticker.Stop()
	}
}

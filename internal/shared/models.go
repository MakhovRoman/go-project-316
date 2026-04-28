package shared

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

// CrawlParams — общий контекст обхода: HTTP-клиент, лимиты, кэш ассетов
// и состояние посещённых URL. Передаётся во все рабочие функции.
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
	UserAgent  string
}

// Visited — потокобезопасное множество уже посещённых URL.
type Visited struct {
	mu    sync.Mutex
	items map[string]struct{}
}

// NewVisited создаёт пустое множество посещённых URL.
func NewVisited() *Visited {
	return &Visited{items: make(map[string]struct{})}
}

// MarkIfNew атомарно добавляет url в множество и возвращает true, если он
// был добавлен впервые, либо false, если уже присутствовал.
func (v *Visited) MarkIfNew(url string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, ok := v.items[url]; ok {
		return false
	}
	v.items[url] = struct{}{}
	return true
}

// AssetType — тип статического ресурса страницы (image, script, style и т.п.).
type AssetType string

// Asset описывает один статический ресурс страницы: URL, тип, HTTP-статус и размер.
type Asset struct {
	URL        string    `json:"url"`
	Type       AssetType `json:"type"`
	StatusCode int       `json:"status_code"`
	SizeBytes  int64     `json:"size_bytes"`
	Error      string    `json:"error,omitempty"`
}

// AssetCache — потокобезопасный кэш ассетов по URL, чтобы не запрашивать
// один и тот же ресурс повторно при обходе нескольких страниц.
type AssetCache struct {
	mu    sync.Mutex
	items map[string]Asset
}

// NewAssetCache создаёт пустой кэш ассетов.
func NewAssetCache() *AssetCache {
	return &AssetCache{items: make(map[string]Asset)}
}

// Get возвращает закэшированный ассет по URL и признак его наличия.
func (c *AssetCache) Get(url string) (Asset, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	a, ok := c.items[url]
	return a, ok
}

// Set сохраняет ассет в кэш под ключом url.
func (c *AssetCache) Set(url string, a Asset) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[url] = a
}

// RateLimiter — простой ограничитель частоты запросов на базе time.Ticker.
// Nil-значение допустимо и означает «без ограничения».
type RateLimiter struct {
	ticker *time.Ticker
}

// NewRateLimiter создаёт ограничитель с заданным минимальным интервалом между
// разрешёнными вызовами. При delay <= 0 возвращает nil — ограничения нет.
func NewRateLimiter(delay time.Duration) *RateLimiter {
	if delay <= 0 {
		return nil
	}
	return &RateLimiter{ticker: time.NewTicker(delay)}
}

// Wait блокируется до следующего «тика» или отмены ctx. Для nil-получателя
// возвращает nil сразу (ограничения нет).
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

// Stop останавливает внутренний тикер. Безопасно вызывается на nil-получателе.
func (r *RateLimiter) Stop() {
	if r != nil && r.ticker != nil {
		r.ticker.Stop()
	}
}

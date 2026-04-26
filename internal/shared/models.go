package shared

import (
	"context"
	"io"
	"net/http"
	"time"
)

type CrawlParams struct {
	CTX        context.Context
	HTTPClient *http.Client
	Host       string
	URL        string
	Depth      uint
	Body       io.Reader
	Queue      *Queue
	Visited    Visited
	Delay      time.Duration
	Retries    uint
	RPS        uint
	AssetCache AssetCache
}

type QueueItem struct {
	URL   string
	Depth uint
}

type Queue []QueueItem

type Visited map[string]struct{}
type AssetType string
type Asset struct {
	URL        string    `json:"url"`
	Type       AssetType `json:"type"`
	StatusCode int       `json:"status_code"`
	SizeBytes  int64     `json:"size_bytes"`
	Error      string    `json:"error"`
}
type AssetCache map[string]Asset

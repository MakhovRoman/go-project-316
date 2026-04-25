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
}

type QueueItem struct {
	URL   string
	Depth uint
}

type Queue []QueueItem

type Visited map[string]struct{}

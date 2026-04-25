package code

import (
	"code/internal/linkchecker"
	"code/internal/parser"
	"net/http"
	"time"
)

type Options struct {
	URL         string
	Depth       uint
	Retries     uint
	RPS         uint
	Delay       time.Duration
	Timeout     time.Duration
	UserAgent   string
	Concurrency uint
	IndentJSON  uint
	HTTPClient  *http.Client
}

type Page struct {
	URL          string                   `json:"url"`
	Depth        uint                     `json:"depth"`
	HTTPStatus   int                      `json:"http_status"`
	Status       string                   `json:"status,omitempty"`
	Error        string                   `json:"error,omitempty"`
	BrokenLinks  []linkchecker.BrokenLink `json:"broken_links,omitempty"`
	DiscoveredAt time.Time                `json:"discovered_at"`
	SEO          parser.SEO               `json:"seo"`
}

type Report struct {
	RootURL     string    `json:"root_url"`
	Depth       uint      `json:"depth"`
	GeneratedAt time.Time `json:"generated_at"`
	Pages       []Page    `json:"pages"`
}

package code

import (
	"bytes"
	"code/internal/linkchecker"
	parser2 "code/internal/parser"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type Options struct {
	URL         string
	Depth       uint
	Retries     uint
	Delay       time.Duration
	Timeout     time.Duration
	UserAgent   string
	Concurrency uint
	IndentJSON  uint
	HTTPClient  *http.Client
}

type Pages struct {
	URL          string                   `json:"url"`
	Depth        uint                     `json:"depth"`
	HTTPStatus   int                      `json:"http_status"`
	Status       string                   `json:"status,omitempty"`
	Error        string                   `json:"error,omitempty"`
	BrokenLinks  []linkchecker.BrokenLink `json:"broken_links,omitempty"`
	DiscoveredAt time.Time                `json:"discovered_at"`
	SEO          parser2.SEO              `json:"seo"`
}

type Report struct {
	RootURL     string    `json:"root_url"`
	Depth       uint      `json:"depth"`
	GeneratedAt time.Time `json:"generated_at"`
	Pages       []Pages   `json:"pages"`
}

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	resp, err := opts.HTTPClient.Get(opts.URL)
	if err != nil {
		return nil, err
	}

	bodyBuffer, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("close body: %v", err)
		}
	}()

	var status = "ok"
	var statusErr string

	if resp.StatusCode != 200 {
		status = "error"
		statusErr = http.StatusText(resp.StatusCode)
	}

	brokenLinks, err := linkchecker.CheckLinks(bytes.NewReader(bodyBuffer), opts.HTTPClient, opts.URL)
	if err != nil {
		return nil, err
	}

	seo, err := parser2.ParseSEO(bytes.NewReader(bodyBuffer))
	if err != nil {
		return nil, err
	}

	report := Report{
		RootURL:     opts.URL,
		Depth:       opts.Depth,
		GeneratedAt: time.Now(),
		Pages: []Pages{
			{
				URL:          resp.Request.URL.String(),
				Depth:        opts.Depth,
				HTTPStatus:   resp.StatusCode,
				Status:       status,
				Error:        statusErr,
				DiscoveredAt: time.Now(),
				BrokenLinks:  brokenLinks,
				SEO:          seo,
			},
		},
	}

	result, err := json.Marshal(report)
	if err != nil {
		return nil, err
	}

	return result, nil
}

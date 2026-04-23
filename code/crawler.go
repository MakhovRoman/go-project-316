package code

import (
	"context"
	"encoding/json"
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

type ReportPages struct {
	URL        string `json:"url"`
	Depth      uint   `json:"depth"`
	HttpStatus int    `json:"http_status"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

type Report struct {
	RootURL     string        `json:"root_url"`
	Depth       uint          `json:"depth"`
	GeneratedAt time.Time     `json:"generated_at"`
	Pages       []ReportPages `json:"pages"`
}

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	resp, err := opts.HTTPClient.Get(opts.URL)
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

	report := Report{
		RootURL:     opts.URL,
		Depth:       opts.Depth,
		GeneratedAt: time.Now(),
		Pages: []ReportPages{
			{URL: resp.Request.URL.String(), Depth: opts.Depth, HttpStatus: resp.StatusCode, Status: status, Error: statusErr},
		},
	}

	result, err := json.Marshal(report)
	if err != nil {
		return nil, err
	}

	return result, nil
}

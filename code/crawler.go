package code

import (
	"code/internal/shared"
	"context"
	"encoding/json"
	"time"
)

const BaseDepth = 0

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	n := int(opts.Concurrency)
	if n <= 0 {
		n = 1
	}

	delay := makeDelay(opts.Delay, opts.RPS)
	limiter := shared.NewRateLimiter(delay)
	defer limiter.Stop()

	host, err := getHost(opts.URL)
	if err != nil {
		return nil, err
	}

	params := shared.CrawlParams{
		CTX:        ctx,
		HTTPClient: opts.HTTPClient,
		Host:       host,
		URL:        opts.URL,
		Depth:      opts.Depth,
		Visited:    shared.NewVisited(),
		Limiter:    limiter,
		Delay:      delay,
		Retries:    opts.Retries,
		RPS:        opts.RPS,
		AssetCache: shared.NewAssetCache(),
	}

	pages, err := bfs(params, opts.Depth, n)
	if err != nil {
		return nil, err
	}

	report := Report{
		RootURL:     opts.URL,
		Depth:       opts.Depth,
		GeneratedAt: time.Now(),
		Pages:       pages,
	}

	result, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}

	return result, nil
}

package code

import (
	"code/internal/shared"
	"context"
	"encoding/json"
	"time"
)

const BaseDepth = 0

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	var queue shared.Queue
	visited := make(shared.Visited)
	assetsCache := make(shared.AssetCache)

	host, err := getHost(opts.URL)
	if err != nil {
		return nil, err
	}

	crawlParams := shared.CrawlParams{
		CTX:        ctx,
		HTTPClient: opts.HTTPClient,
		Host:       host,
		URL:        opts.URL,
		Queue:      &queue,
		Visited:    visited,
		Delay:      makeDelay(opts.Delay, opts.RPS),
		Retries:    opts.Retries,
		RPS:        opts.RPS,
		AssetCache: assetsCache,
	}

	pages, err := bfs(crawlParams, opts.Depth, &queue, visited)
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

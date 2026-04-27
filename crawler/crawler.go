package crawler

import (
	"code/internal/helpers"
	"code/internal/shared"
	"context"
)

const BaseDepth = 0

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	n := int(opts.Concurrency) //#nosec G115 -- concurrency from CLI flag, fits in int
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
		URL:        helpers.NormalizeURL(opts.URL),
		Depth:      opts.Depth,
		Visited:    shared.NewVisited(),
		Limiter:    limiter,
		Delay:      delay,
		Retries:    opts.Retries,
		RPS:        opts.RPS,
		AssetCache: shared.NewAssetCache(),
		UserAgent:  opts.UserAgent,
	}

	pages, err := bfs(params, opts.Depth, n)
	if err != nil {
		return nil, err
	}

	report := Report{
		RootURL:     opts.URL,
		Depth:       opts.Depth,
		GeneratedAt: getTime(),
		Pages:       pages,
	}

	result, err := getResultFormat(opts.IndentJSON, report)
	if err != nil {
		return nil, err
	}

	return result, nil
}

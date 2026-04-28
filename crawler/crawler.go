// Package crawler обходит сайт в ширину начиная с заданного URL и формирует
// JSON-отчёт о страницах: их статусах, найденных ссылках, ассетах и SEO-метаданных.
package crawler

import (
	"code/internal/shared"
	"context"
)

// BaseDepth — глубина корневой страницы при обходе (страница, с которой начинается анализ).
const BaseDepth = 0

// Analyze обходит сайт начиная с opts.URL до глубины opts.Depth и возвращает
// сериализованный JSON-отчёт. Учитывает ограничения по конкурентности, задержке
// между запросами и числу повторов. Возвращает ошибку, если корневой URL некорректен
// или не удалось сформировать отчёт.
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
		URL:        opts.URL,
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

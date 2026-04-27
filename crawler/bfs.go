package crawler

import (
	"code/internal/helpers"
	"code/internal/shared"
	"context"
	"errors"
	"log"
	"sort"
	"sync"
)

type pool struct {
	params   shared.CrawlParams
	maxDepth uint
	sem      chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	pages    []Page
}

func newPool(params shared.CrawlParams, maxDepth uint, concurrency int) *pool {
	return &pool{
		params:   params,
		maxDepth: maxDepth,
		sem:      make(chan struct{}, concurrency),
	}
}

func (p *pool) run(url string, depth uint) {
	defer p.wg.Done()

	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	if p.params.CTX.Err() != nil {
		return
	}

	result, err := makePageReport(p.params, url, depth)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		log.Printf("crawl %s: %v", url, err)
		return
	}

	p.mu.Lock()
	p.pages = append(p.pages, result.page)
	p.mu.Unlock()

	p.spawn(result.internalURLs, depth+1)
}

func (p *pool) spawn(urls []string, childDepth uint) {
	if childDepth >= p.maxDepth {
		return
	}
	for _, u := range urls {
		if p.params.Visited.MarkIfNew(helpers.NormalizeURL(u)) {
			p.wg.Add(1)
			go p.run(helpers.NormalizeURL(u), childDepth)
		}
	}
}

func bfs(params shared.CrawlParams, maxDepth uint, concurrency int) ([]Page, error) {
	params.Visited.MarkIfNew(helpers.NormalizeURL(params.URL))

	rootResult, err := makePageReport(params, params.URL, BaseDepth)
	if err != nil {
		return nil, err
	}

	p := newPool(params, maxDepth, concurrency)
	p.pages = []Page{rootResult.page}
	p.spawn(rootResult.internalURLs, BaseDepth+1)
	p.wg.Wait()

	sort.Slice(p.pages, func(i, j int) bool {
		return p.pages[i].URL < p.pages[j].URL
	})

	return p.pages, nil
}

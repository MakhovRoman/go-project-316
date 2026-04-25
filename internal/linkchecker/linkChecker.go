package linkchecker

import (
	"code/internal/helpers"
	"code/internal/parser"
	"code/internal/shared"
	"log"
	"net/http"
	"net/url"
)

func CheckLinks(params shared.CrawlParams, path string, depth uint) ([]BrokenLink, error) {
	links, err := parser.ParseHTML(params.Body, path)

	if err != nil {
		return nil, err
	}

	var brokenLinks []BrokenLink

	for _, link := range links {
		safeURL, e := helpers.ValidateURL(link.URL)
		if e != nil {
			brokenLinks = append(brokenLinks, BrokenLink{URL: link.URL, Error: e.Error()})
			continue
		}
		req, e := http.NewRequestWithContext(params.CTX, http.MethodGet, safeURL, nil)
		if e != nil {
			brokenLinks = append(brokenLinks, BrokenLink{URL: link.URL, Error: e.Error()})
			continue
		}
		r, e := params.HTTPClient.Do(req)
		if e != nil {
			brokenLinks = append(brokenLinks, BrokenLink{URL: link.URL, Error: e.Error()})
			continue
		}

		if r.StatusCode != http.StatusOK {
			brokenLinks = append(brokenLinks, BrokenLink{URL: link.URL, StatusCode: r.StatusCode})
		} else {
			if err := addToQueue(link.URL, params.Host, depth, params.Queue, params.Visited); err != nil {
				return nil, err
			}
		}

		if err := r.Body.Close(); err != nil {
			log.Printf("Error closing body: %v", err)
		}
	}

	return brokenLinks, nil
}

func addToQueue(path, host string, depth uint, queue *shared.Queue, visited shared.Visited) error {
	u, err := url.Parse(path)
	if err != nil {
		return nil
	}

	if _, ok := visited[path]; !ok && u.Host == host {
		*queue = append(*queue, shared.QueueItem{
			URL:   path,
			Depth: depth + 1,
		})
		visited[path] = struct{}{}
	}

	return nil
}

package linkchecker

import (
	"code/internal/helpers"
	"code/internal/parser"
	"code/internal/request"
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

linksFor:
	for _, link := range links {
		safeURL, e := helpers.ValidateURL(link.URL)
		if e != nil {
			brokenLinks = append(brokenLinks, BrokenLink{URL: link.URL, Error: e.Error()})
			continue
		}

		var r *http.Response

		retries := int(params.Retries) // #nosec G115 -- retries from CLI flag, fits in int
		for i := 0; i <= retries; i++ {
			if err := shared.RetryDelay(params, i); err != nil {
				return nil, err
			}

			retry, err := request.DoRequestWithRetry(params, &r, i, safeURL)
			if err != nil {
				brokenLinks = append(brokenLinks, BrokenLink{URL: link.URL, Error: err.Error()})
				continue linksFor
			}
			if retry {
				continue
			}
			break
		}

		if r == nil {
			brokenLinks = append(brokenLinks, BrokenLink{URL: link.URL, Error: "no response"})
			continue linksFor
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

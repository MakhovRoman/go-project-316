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

type Result struct {
	Broken   []BrokenLink
	Internal []string
}

func CheckLinks(params shared.CrawlParams, path string) (Result, error) {
	links, err := parser.ParseHTML(params.Body, path)
	if err != nil {
		return Result{}, err
	}

	var internal []string
	broken := make([]BrokenLink, 0)
	seen := make(map[string]struct{})

linksFor:
	for _, link := range links {
		safeURL, e := helpers.ValidateURL(link.URL)
		if e != nil {
			continue
		}
		key := helpers.NormalizeURL(safeURL)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		var r *http.Response

		retries := int(params.Retries) // #nosec G115 -- retries from CLI flag, fits in int
		for i := 0; i <= retries; i++ {
			if err := shared.RetryDelay(params, i); err != nil {
				return Result{}, err
			}

			retry, err := request.DoRequestWithRetry(params, &r, i, safeURL)
			if err != nil {
				if isInternal(safeURL, params.Host) {
					broken = append(broken, BrokenLink{URL: link.URL, Error: err.Error()})
				}
				continue linksFor
			}
			if retry {
				continue
			}
			break
		}

		if r == nil {
			if isInternal(safeURL, params.Host) {
				broken = append(broken, BrokenLink{URL: link.URL, Error: "no response"})
			}
			continue linksFor
		}

		if isInternal(safeURL, params.Host) {
			if r.StatusCode >= http.StatusBadRequest {
				broken = append(broken, BrokenLink{URL: link.URL, StatusCode: r.StatusCode, Error: http.StatusText(int(r.StatusCode))})
			} else {
				internal = append(internal, safeURL)
			}
		}

		if err := r.Body.Close(); err != nil {
			log.Printf("Error closing body: %v", err)
		}
	}

	return Result{Broken: broken, Internal: internal}, nil
}

func isInternal(rawURL, host string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Host == host
}

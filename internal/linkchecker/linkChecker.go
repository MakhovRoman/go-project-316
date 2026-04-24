package linkchecker

import (
	"code/internal/parser"
	"io"
	"log"
	"net/http"
)

type BrokenLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

func CheckLinks(body io.Reader, client *http.Client, path string) ([]BrokenLink, error) {
	links, err := parser.ParseHTML(body, path)
	if err != nil {
		return nil, err
	}

	var brokenLinks []BrokenLink

	for _, link := range links {
		r, e := client.Get(link)
		if e != nil {
			brokenLinks = append(brokenLinks, BrokenLink{URL: link, Error: e.Error()})
			continue
		}

		if r.StatusCode != http.StatusOK {
			brokenLinks = append(brokenLinks, BrokenLink{URL: link, StatusCode: r.StatusCode})
		}

		if err := r.Body.Close(); err != nil {
			log.Printf("Error closing body: %v", err)
		}
	}

	return brokenLinks, nil
}

package parser

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type BrokenLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

func ParseHTML(body io.ReadCloser, client *http.Client, path string) ([]BrokenLink, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	var brokenLinks []BrokenLink

	for n := range doc.Descendants() {
		if n.Type != html.ElementNode || n.Data != "a" {
			continue
		}

		for _, a := range n.Attr {
			if a.Key != "href" || a.Val == "" {
				continue
			}

			ref, err := url.Parse(a.Val)
			if err != nil {
				continue
			}

			resolved := base.ResolveReference(ref)
			if resolved.Scheme != "http" && resolved.Scheme != "https" {
				continue
			}

			dataUrl := resolved.String()

			r, e := client.Get(dataUrl)
			if e != nil {
				brokenLinks = append(brokenLinks, BrokenLink{URL: dataUrl, Error: e.Error()})
				continue
			}

			if r.StatusCode != http.StatusOK {
				brokenLinks = append(brokenLinks, BrokenLink{URL: dataUrl, StatusCode: r.StatusCode})
			}

			if err := r.Body.Close(); err != nil {
				log.Printf("Error closing body: %v", err)
			}

		}

	}

	return brokenLinks, nil
}

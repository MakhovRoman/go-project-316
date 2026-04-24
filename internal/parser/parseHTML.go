package parser

import (
	"io"
	"net/url"

	"golang.org/x/net/html"
)

func ParseHTML(body io.Reader, path string) ([]string, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	var links []string

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

			if ref.Host == "" && ref.Path == "" && ref.RawQuery == "" {
				continue
			}

			resolved := base.ResolveReference(ref)
			if resolved.Scheme != "http" && resolved.Scheme != "https" {
				continue
			}

			dataURL := resolved.String()
			links = append(links, dataURL)
		}
	}

	return links, nil
}

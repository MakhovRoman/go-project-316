package parser

import (
	"code/internal/helpers"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

type Link struct {
	URL      string
	Internal bool
}

func ParseHTML(body io.Reader, path string) ([]Link, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	var links []Link

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
			if err := helpers.ValidateURL(resolved.String()); err != nil {
				continue
			}

			var link Link

			link.URL = resolved.String()
			links = append(links, link)
		}
	}

	return links, nil
}

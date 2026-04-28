package parser

import (
	"code/internal/helpers"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

// Link — найденная в HTML ссылка <a href>: абсолютный URL и признак внутренней ссылки.
type Link struct {
	URL      string
	Internal bool
}

// ParseHTML извлекает все валидные ссылки <a href> из HTML-документа.
// Относительные URL разрешаются относительно path; ссылки с пустым/некорректным
// или неподдерживаемым URL пропускаются.
func ParseHTML(body io.Reader, path string) ([]Link, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	var links []Link

	for n := range doc.Descendants() {
		if n.Type != html.ElementNode || n.Data != "a" {
			continue
		}

		for _, attr := range n.Attr {
			if attr.Key != "href" || attr.Val == "" {
				continue
			}

			ref, err := url.Parse(attr.Val)
			if err != nil {
				continue
			}

			if ref.Host == "" && ref.Path == "" && ref.RawQuery == "" {
				continue
			}

			resolved := baseURL.ResolveReference(ref)
			safeURL, err := helpers.ValidateURL(resolved.String())
			if err != nil {
				continue
			}

			var link Link

			link.URL = safeURL
			links = append(links, link)
		}
	}

	return links, nil
}

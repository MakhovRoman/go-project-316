package parser

import (
	"io"

	"github.com/PuerkitoBio/goquery"
)

// SEO — собранные SEO-метаданные страницы: наличие и содержимое <title>,
// <meta name="description"> и наличие <h1>.
type SEO struct {
	HasTitle       bool   `json:"has_title"`
	Title          string `json:"title"`
	HasDescription bool   `json:"has_description"`
	Description    string `json:"description"`
	HasH1          bool   `json:"has_h1"`
}

// ParseSEO извлекает SEO-метаданные из HTML-документа: title, description и факт наличия h1.
func ParseSEO(body io.Reader) (SEO, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return SEO{}, err
	}

	var seo SEO

	if title := doc.Find("title").First().Text(); title != "" {
		seo.HasTitle = true
		seo.Title = title
	}

	if des, ok := doc.Find("meta[name='description']").Attr("content"); ok && des != "" {
		seo.HasDescription = true
		seo.Description = des
	}

	if h1 := doc.Find("h1").First().Text(); h1 != "" {
		seo.HasH1 = true
	}

	return seo, nil
}

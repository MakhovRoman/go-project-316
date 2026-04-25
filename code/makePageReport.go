package code

import (
	"code/internal/linkchecker"
	"code/internal/parser"
	"code/internal/shared"
	"log"
	"net/http"
	"time"
)

func makePageReport(params shared.CrawlParams, path string, depth uint) (Page, error) {
	var page Page

	if err := SleepContext(params.CTX, params.Delay); err != nil {
		return page, err
	}

	resp, bodyBuffer, err := request(params.CTX, params.HTTPClient, path)
	if err != nil {
		return page, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("close body: %v", err)
		}
	}()

	var status = "ok"
	var statusErr string

	if resp.StatusCode != 200 {
		status = "error"
		statusErr = http.StatusText(resp.StatusCode)
	}

	bodyReader, err := makeReader(bodyBuffer, resp)
	if err != nil {
		return page, err
	}
	params.Body = bodyReader

	if err := SleepContext(params.CTX, params.Delay); err != nil {
		return page, err
	}

	brokenLinks, err := linkchecker.CheckLinks(params, path, depth)
	if err != nil {
		return page, err
	}

	seoReader, err := makeReader(bodyBuffer, resp)
	if err != nil {
		return page, err
	}
	seo, err := parser.ParseSEO(seoReader)
	if err != nil {
		return page, err
	}

	page = Page{
		URL:          resp.Request.URL.String(),
		Depth:        depth,
		HTTPStatus:   resp.StatusCode,
		Status:       status,
		Error:        statusErr,
		DiscoveredAt: time.Now(),
		BrokenLinks:  brokenLinks,
		SEO:          seo,
	}

	return page, nil
}

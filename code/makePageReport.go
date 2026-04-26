package code

import (
	"code/internal/fetchasset"
	"code/internal/linkchecker"
	"code/internal/parser"
	"code/internal/request"
	"code/internal/shared"
	"log"
	"net/http"
	"time"
)

func makePageReport(params shared.CrawlParams, path string, depth uint) (Page, error) {
	var page Page

	if err := shared.SleepContext(params.CTX, params.Delay); err != nil {
		return page, err
	}

	res := request.Request(params, path)
	if res.Err != nil {
		return page, res.Err
	}

	defer func() {
		if err := res.Response.Body.Close(); err != nil {
			log.Printf("close body: %v", err)
		}
	}()

	var status = "ok"
	var statusErr string

	if res.Response.StatusCode != 200 {
		status = "error"
		statusErr = http.StatusText(res.Response.StatusCode)
	}

	bodyReader, err := makeReader(res.Body, res.Response)
	if err != nil {
		return page, err
	}
	params.Body = bodyReader

	if err := shared.SleepContext(params.CTX, params.Delay); err != nil {
		return page, err
	}

	brokenLinks, err := linkchecker.CheckLinks(params, path, depth)
	if err != nil {
		return page, err
	}

	seoReader, err := makeReader(res.Body, res.Response)
	if err != nil {
		return page, err
	}
	seo, err := parser.ParseSEO(seoReader)
	if err != nil {
		return page, err
	}

	assetsReader, err := makeReader(res.Body, res.Response)
	if err != nil {
		return page, err
	}
	parsedAssets, err := parser.ParseAssets(assetsReader, path)
	if err != nil {
		return page, err
	}

	assets := make([]shared.Asset, 0, len(parsedAssets))
	for _, a := range parsedAssets {
		assets = append(assets, fetchasset.FetchAsset(params, a))
	}

	page = Page{
		URL:          res.Response.Request.URL.String(),
		Depth:        depth,
		HTTPStatus:   res.Response.StatusCode,
		Status:       status,
		Error:        statusErr,
		DiscoveredAt: time.Now(),
		BrokenLinks:  brokenLinks,
		SEO:          seo,
		Assets:       assets,
	}

	return page, nil
}

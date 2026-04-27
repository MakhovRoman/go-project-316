package code

import (
	"code/internal/fetchasset"
	"code/internal/linkchecker"
	"code/internal/parser"
	"code/internal/request"
	"code/internal/shared"
	"log"
	"net/http"
)

type pageResult struct {
	page         Page
	internalURLs []string
}

func makePageReport(params shared.CrawlParams, path string, depth uint) (pageResult, error) {
	res := request.Request(params, path)
	if res.Err != nil {
		return pageResult{}, res.Err
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
		return pageResult{}, err
	}
	params.Body = bodyReader

	linksResult, err := linkchecker.CheckLinks(params, path)
	if err != nil {
		return pageResult{}, err
	}

	seoReader, err := makeReader(res.Body, res.Response)
	if err != nil {
		return pageResult{}, err
	}
	seo, err := parser.ParseSEO(seoReader)
	if err != nil {
		return pageResult{}, err
	}

	assetsReader, err := makeReader(res.Body, res.Response)
	if err != nil {
		return pageResult{}, err
	}
	parsedAssets, err := parser.ParseAssets(assetsReader, path)
	if err != nil {
		return pageResult{}, err
	}

	assets := make([]shared.Asset, 0, len(parsedAssets))
	for _, a := range parsedAssets {
		assets = append(assets, fetchasset.FetchAsset(params, a))
	}

	page := Page{
		URL:          res.Response.Request.URL.String(),
		Depth:        depth,
		HTTPStatus:   res.Response.StatusCode,
		Status:       status,
		Error:        statusErr,
		DiscoveredAt: getTime(),
		BrokenLinks:  linksResult.Broken,
		SEO:          seo,
		Assets:       assets,
	}

	return pageResult{page: page, internalURLs: linksResult.Internal}, nil
}

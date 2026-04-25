package code

import (
	"bytes"
	"code/internal/helpers"
	"code/internal/linkchecker"
	"code/internal/parser"
	"code/internal/shared"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html/charset"
)

const LimitReader = 10 * 1024 * 1024

const BaseDepth = 0

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	var queue shared.Queue
	visited := make(shared.Visited)

	host, err := getHost(opts.URL)
	if err != nil {
		return nil, err
	}

	crawlParams := shared.CrawlParams{
		CTX:        ctx,
		HTTPClient: opts.HTTPClient,
		Host:       host,
		URL:        opts.URL,
		Queue:      &queue,
		Visited:    visited,
	}

	var pages []Page

	basePage, err := makePageReport(crawlParams, crawlParams.URL, BaseDepth)
	if err != nil {
		return nil, err
	}
	visited[opts.URL] = struct{}{}
	pages = append(pages, basePage)

	for i := BaseDepth; i < len(queue); i++ {

		if queue[i].Depth >= opts.Depth {
			break
		}
		page, err := makePageReport(crawlParams, queue[i].URL, queue[i].Depth)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				break
			}
			return nil, err
		}

		pages = append(pages, page)
	}

	report := Report{
		RootURL:     opts.URL,
		Depth:       opts.Depth,
		GeneratedAt: time.Now(),
		Pages:       pages,
	}

	result, err := json.Marshal(report)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getHost(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return u.Host, nil
}

func makeReader(buff []byte, resp *http.Response) (io.Reader, error) {
	if len(buff) == 0 {
		return bytes.NewReader(buff), nil
	}

	reader, err := charset.NewReader(bytes.NewReader(buff), resp.Header.Get("Content-Type"))
	if err != nil {
		return bytes.NewReader(buff), nil
	}

	return reader, nil
}

func request(ctx context.Context, client *http.Client, path string) (*http.Response, []byte, error) {
	safeURL, err := helpers.ValidateURL(path)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, safeURL, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	bodyBuffer, err := io.ReadAll(io.LimitReader(resp.Body, LimitReader))
	if err != nil {
		return nil, nil, err
	}

	return resp, bodyBuffer, nil
}

func makePageReport(params shared.CrawlParams, path string, depth uint) (Page, error) {
	var page Page

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

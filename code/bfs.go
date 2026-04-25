package code

import (
	"code/internal/shared"
	"context"
	"errors"
)

func bfs(params shared.CrawlParams, depth uint, queue *shared.Queue, visited shared.Visited) ([]Page, error) {
	var pages []Page

	basePage, err := makePageReport(params, params.URL, BaseDepth)
	if err != nil {
		return nil, err
	}
	visited[params.URL] = struct{}{}
	pages = append(pages, basePage)

	for i := BaseDepth; i < len(*queue); i++ {
		if (*queue)[i].Depth >= depth {
			break
		}
		page, err := makePageReport(params, (*queue)[i].URL, (*queue)[i].Depth)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				break
			}
			return nil, err
		}

		pages = append(pages, page)
	}

	return pages, nil
}

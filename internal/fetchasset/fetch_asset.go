package fetchasset

import (
	"code/internal/request"
	"code/internal/shared"
	"fmt"
	"io"
	"log"
	"net/http"
)

func FetchAsset(params shared.CrawlParams, asset shared.Asset) shared.Asset {
	if cached, ok := params.AssetCache.Get(asset.URL); ok {
		return cached
	}

	var resp *http.Response

	retries := int(params.Retries) // #nosec G115 -- retries from CLI flag, fits in int
	for i := 0; i <= retries; i++ {
		if err := shared.RetryDelay(params, i); err != nil {
			asset.Error = err.Error()
			params.AssetCache.Set(asset.URL, asset)
			return asset
		}

		retry, err := request.DoRequestWithRetry(params, &resp, i, asset.URL, http.MethodHead)
		if err != nil {
			asset.Error = err.Error()
			params.AssetCache.Set(asset.URL, asset)
			return asset
		}
		if retry {
			continue
		}

		break
	}

	if resp == nil {
		asset.Error = "no response"
		params.AssetCache.Set(asset.URL, asset)
		return asset
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing asset body: %v", err)
		}
	}()

	asset.StatusCode = resp.StatusCode

	if resp.ContentLength >= 0 {
		asset.SizeBytes = resp.ContentLength
	} else {
		n, err := io.Copy(io.Discard, io.LimitReader(resp.Body, request.LimitReader))
		if err != nil {
			asset.Error = err.Error()
			asset.SizeBytes = 0
		} else {
			asset.SizeBytes = n
		}
	}

	if resp.StatusCode >= 400 {
		asset.Error = fmt.Sprintf("http %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	params.AssetCache.Set(asset.URL, asset)
	return asset
}

package request

import (
	"code/internal/helpers"
	"code/internal/shared"
	"errors"
	"io"
	"net/http"
)

const LimitReader = 10 * 1024 * 1024

type ResultRequest struct {
	Response *http.Response
	Body     []byte
	Err      error
}

func (r *ResultRequest) setError(err error) {
	r.Err = err
}

func Request(params shared.CrawlParams, path string) ResultRequest {
	safeURL, err := helpers.ValidateURL(path)
	if err != nil {
		return ResultRequest{Err: err}
	}

	var resp *http.Response

	for i := 0; i <= int(params.Retries); i++ {
		if err := shared.RetryDelay(params, i); err != nil {
			return ResultRequest{Err: err}
		}

		//req, err := http.NewRequestWithContext(params.CTX, http.MethodGet, safeURL, nil)
		//if err != nil {
		//	return ResultRequest{Err: err}
		//}
		//
		//resp, err = params.HTTPClient.Do(req) // #nosec G704 -- URL validated and reconstructed via helpers.ValidateURL
		//if err != nil {
		//	if i == int(params.Retries) {
		//		return ResultRequest{Err: err}
		//	}
		//
		//	continue
		//} else if resp != nil && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
		//	if i == int(params.Retries) {
		//		break
		//	}
		//
		//	continue
		//}
		//
		//break

		retry, err := DoRequestWithRetry(params, &resp, i, safeURL)
		if err != nil {
			return ResultRequest{Err: err}
		}
		if retry {
			continue
		}

		break
	}

	if resp == nil {
		return ResultRequest{Err: errors.New("no response after retries")}
	}

	bodyBuffer, err := io.ReadAll(io.LimitReader(resp.Body, LimitReader))
	if err != nil {
		return ResultRequest{Err: err}
	}

	return ResultRequest{Response: resp, Body: bodyBuffer, Err: nil}
}

func DoRequestWithRetry(params shared.CrawlParams, resp **http.Response, attempt int, path string) (bool, error) {
	req, err := http.NewRequestWithContext(params.CTX, http.MethodGet, path, nil)
	if err != nil {
		return false, err
	}

	*resp, err = (params.HTTPClient).Do(req) // #nosec G704 -- URL validated and reconstructed via helpers.ValidateURL
	if err != nil {
		if attempt == int(params.Retries) {
			return false, err
		}

		return true, nil
	} else if *resp != nil && ((*resp).StatusCode == 429 || (*resp).StatusCode >= 500) {
		if attempt == int(params.Retries) {
			return false, nil
		}

		return true, nil
	}

	return false, nil
}

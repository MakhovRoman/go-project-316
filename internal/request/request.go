package request

import (
	"code/internal/helpers"
	"code/internal/shared"
	"errors"
	"io"
	"net/http"
)

// LimitReader — верхняя граница на размер вычитываемого тела ответа (10 MiB),
// чтобы один большой ответ не съел память.
const LimitReader = 10 * 1024 * 1024

// ResultRequest — результат GET-запроса страницы: оригинальный *http.Response,
// уже прочитанное тело и ошибка (если была).
type ResultRequest struct {
	Response *http.Response
	Body     []byte
	Err      error
}

// Request делает GET-запрос по path с учётом повторов (params.Retries) для 5xx/429
// и сетевых ошибок и возвращает ответ вместе с вычитанным телом, ограниченным LimitReader.
func Request(params shared.CrawlParams, path string) ResultRequest {
	safeURL, err := helpers.ValidateURL(path)
	if err != nil {
		return ResultRequest{Err: err}
	}

	var resp *http.Response

	retries := int(params.Retries) // #nosec G115 -- retries from CLI flag, fits in int
	for i := 0; i <= retries; i++ {
		if err := shared.RetryDelay(params, i); err != nil {
			return ResultRequest{Err: err}
		}

		retry, err := DoRequestWithRetry(params, &resp, i, safeURL, http.MethodGet)
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

// DoRequestWithRetry выполняет одну попытку HTTP-запроса заданным методом.
// Возвращает (true, nil), если стоит повторить (5xx/429 или сетевая ошибка и попытки ещё есть),
// (false, nil) — окончательный ответ получен, (false, err) — фатальная ошибка.
// Ответ записывается через указатель resp.
func DoRequestWithRetry(params shared.CrawlParams, resp **http.Response, attempt int, path string, method string) (bool, error) {
	req, err := http.NewRequestWithContext(params.CTX, method, path, nil)
	if err != nil {
		return false, err
	}

	retries := int(params.Retries) // #nosec G115 -- retries from CLI flag, fits in int

	if err := params.Limiter.Wait(params.CTX); err != nil {
		return false, err
	}

	if params.UserAgent != "" {
		req.Header.Set("User-Agent", params.UserAgent)
	}
	*resp, err = (params.HTTPClient).Do(req) // #nosec G704 -- URL validated and reconstructed via helpers.ValidateURL
	if err != nil {
		if attempt == retries {
			return false, err
		}

		return true, nil
	} else if *resp != nil && ((*resp).StatusCode == 429 || (*resp).StatusCode >= 500) {
		if attempt == retries {
			return false, nil
		}

		return true, nil
	}

	return false, nil
}

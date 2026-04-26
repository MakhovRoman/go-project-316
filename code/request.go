package code

//
//import (
//	"code/internal/helpers"
//	"code/internal/shared"
//	"io"
//	"net/http"
//	"time"
//)
//
//const BaseDelay = 100 * time.Millisecond
//
//func request(params shared.CrawlParams, path string) (*http.Response, []byte, error) {
//	safeURL, err := helpers.ValidateURL(path)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	var resp *http.Response
//
//	for i := 0; i <= int(params.Retries); i++ {
//		if i > 0 {
//			delay := params.Delay
//			if delay == 0 {
//				delay = BaseDelay
//			}
//
//			if err := SleepContext(params.CTX, delay); err != nil {
//				return nil, nil, err
//			}
//		}
//
//		req, err := http.NewRequestWithContext(params.CTX, http.MethodGet, safeURL, nil)
//		if err != nil {
//			return nil, nil, err
//		}
//
//		resp, err = params.HTTPClient.Do(req) // #nosec G704 -- URL validated and reconstructed via helpers.ValidateURL
//		if err != nil {
//			if i == int(params.Retries) {
//				return nil, nil, err
//			}
//
//			continue
//		} else if resp != nil && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
//			if i == int(params.Retries) {
//				break
//			}
//
//			continue
//		}
//
//		break
//	}
//
//	bodyBuffer, err := io.ReadAll(io.LimitReader(resp.Body, LimitReader))
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return resp, bodyBuffer, nil
//}

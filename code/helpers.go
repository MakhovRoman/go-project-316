package code

import (
	"bytes"
	"code/internal/helpers"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html/charset"
)

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

	resp, err := client.Do(req) // #nosec G704 -- URL validated and reconstructed via helpers.ValidateURL
	if err != nil {
		return nil, nil, err
	}

	bodyBuffer, err := io.ReadAll(io.LimitReader(resp.Body, LimitReader))
	if err != nil {
		return nil, nil, err
	}

	return resp, bodyBuffer, nil
}

func makeDelay(delay time.Duration, rps uint) time.Duration {
	if rps == 0 {
		return delay
	}

	return time.Second / time.Duration(int64(rps))
}

func SleepContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

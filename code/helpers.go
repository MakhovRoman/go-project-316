package code

import (
	"bytes"
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

func makeDelay(delay time.Duration, rps uint) time.Duration {
	if rps == 0 {
		return delay
	}

	return time.Second / time.Duration(int64(rps)) // #nosec G115 -- rps is a small positive value, overflow is impossible in practice
}

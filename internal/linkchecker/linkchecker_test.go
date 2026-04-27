package linkchecker

import (
	"code/internal/shared"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func makeParams(client *http.Client, host string, body string) shared.CrawlParams {
	return shared.CrawlParams{
		CTX:        context.Background(),
		HTTPClient: client,
		Host:       host,
		Body:       strings.NewReader(body),
		Visited:    shared.NewVisited(),
	}
}

func TestCheckLinks_CollectsBrokenLinks(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/not-found", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/server-error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	html := `
		<html><body>
			<a href="/ok">ok</a>
			<a href="/not-found">broken</a>
			<a href="/server-error">server error</a>
		</body></html>
	`
	params := makeParams(server.Client(), server.Listener.Addr().String(), html)

	res, err := CheckLinks(params, server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Broken) != 2 {
		t.Fatalf("expected 2 broken links, got %d: %+v", len(res.Broken), res.Broken)
	}

	codes := map[int]bool{}
	for _, b := range res.Broken {
		codes[b.StatusCode] = true
	}
	if !codes[http.StatusNotFound] || !codes[http.StatusInternalServerError] {
		t.Errorf("expected 404 and 500, got %+v", res.Broken)
	}
}

func TestCheckLinks_NetworkError(t *testing.T) {
	html := `
		<html><body>
			<a href="http://nonexistent.invalid.localhost">bad</a>
		</body></html>
	`
	params := makeParams(&http.Client{}, "example.com", html)

	res, err := CheckLinks(params, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Broken) != 0 {
		t.Fatalf("expected 0 broken links (external errors skipped), got %d: %+v", len(res.Broken), res.Broken)
	}
}

func TestCheckLinks_NoLinks(t *testing.T) {
	params := makeParams(&http.Client{}, "example.com", `<html><body><p>no links</p></body></html>`)

	res, err := CheckLinks(params, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Broken) != 0 {
		t.Errorf("expected no broken links, got %+v", res.Broken)
	}
}

func TestCheckLinks_IgnoresUnsupportedSchemes(t *testing.T) {
	html := `
		<html><body>
			<a href="mailto:foo@bar.com">email</a>
			<a href="#anchor">fragment</a>
		</body></html>
	`
	params := makeParams(&http.Client{}, "example.com", html)

	res, err := CheckLinks(params, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Broken) != 0 {
		t.Errorf("expected no broken links, got %+v", res.Broken)
	}
}

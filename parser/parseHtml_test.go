package parser

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupServer(t *testing.T, html string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	})
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/not-found", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/server-error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	return httptest.NewServer(mux)
}

func TestParseHTML_CollectsBrokenLinks(t *testing.T) {
	server := setupServer(t, `
		<html><body>
			<a href="/ok">ok</a>
			<a href="/not-found">broken</a>
			<a href="/server-error">server error</a>
		</body></html>
	`)
	defer server.Close()

	body := io.NopCloser(strings.NewReader(`
		<html><body>
			<a href="/ok">ok</a>
			<a href="/not-found">broken</a>
			<a href="/server-error">server error</a>
		</body></html>
	`))

	broken, err := ParseHTML(body, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(broken) != 2 {
		t.Fatalf("expected 2 broken links, got %d: %+v", len(broken), broken)
	}

	codes := map[int]bool{}
	for _, b := range broken {
		codes[b.StatusCode] = true
	}
	if !codes[http.StatusNotFound] || !codes[http.StatusInternalServerError] {
		t.Errorf("expected 404 and 500, got %+v", broken)
	}
}

func TestParseHTML_IgnoresUnsupportedSchemes(t *testing.T) {
	server := setupServer(t, "")
	defer server.Close()

	body := io.NopCloser(strings.NewReader(`
		<html><body>
			<a href="mailto:foo@bar.com">email</a>
			<a href="tel:+123">phone</a>
			<a href="javascript:void(0)">js</a>
			<a href="#anchor">fragment</a>
			<a href="">empty</a>
		</body></html>
	`))

	broken, err := ParseHTML(body, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(broken) != 0 {
		t.Errorf("expected no broken links, got %+v", broken)
	}
}

func TestParseHTML_NetworkError(t *testing.T) {
	body := io.NopCloser(strings.NewReader(`
		<html><body>
			<a href="http://nonexistent.invalid.localhost">bad</a>
		</body></html>
	`))

	broken, err := ParseHTML(body, &http.Client{}, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(broken) != 1 {
		t.Fatalf("expected 1 broken link, got %d", len(broken))
	}
	if broken[0].Error == "" {
		t.Error("expected non-empty error message")
	}
	if broken[0].StatusCode != 0 {
		t.Errorf("expected status code 0, got %d", broken[0].StatusCode)
	}
}

func TestParseHTML_ResolvesRelativeURLs(t *testing.T) {
	server := setupServer(t, "")
	defer server.Close()

	body := io.NopCloser(strings.NewReader(`
		<html><body>
			<a href="/not-found">relative root</a>
		</body></html>
	`))

	broken, err := ParseHTML(body, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(broken) != 1 {
		t.Fatalf("expected 1 broken link, got %d", len(broken))
	}
	if broken[0].URL != server.URL+"/not-found" {
		t.Errorf("expected absolute URL %s/not-found, got %s", server.URL, broken[0].URL)
	}
}

func TestParseHTML_NoLinks(t *testing.T) {
	body := io.NopCloser(strings.NewReader(`<html><body><p>no links here</p></body></html>`))

	broken, err := ParseHTML(body, &http.Client{}, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(broken) != 0 {
		t.Errorf("expected no broken links, got %+v", broken)
	}
}

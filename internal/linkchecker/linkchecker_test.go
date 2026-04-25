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
	var queue shared.Queue
	return shared.CrawlParams{
		CTX:        context.Background(),
		HTTPClient: client,
		Host:       host,
		Body:       strings.NewReader(body),
		Queue:      &queue,
		Visited:    make(shared.Visited),
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

	broken, err := CheckLinks(params, server.URL, 0)
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

func TestCheckLinks_NetworkError(t *testing.T) {
	html := `
		<html><body>
			<a href="http://nonexistent.invalid.localhost">bad</a>
		</body></html>
	`
	params := makeParams(&http.Client{}, "example.com", html)

	broken, err := CheckLinks(params, "http://example.com", 0)
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

func TestCheckLinks_NoLinks(t *testing.T) {
	params := makeParams(&http.Client{}, "example.com", `<html><body><p>no links</p></body></html>`)

	broken, err := CheckLinks(params, "http://example.com", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(broken) != 0 {
		t.Errorf("expected no broken links, got %+v", broken)
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

	broken, err := CheckLinks(params, "http://example.com", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(broken) != 0 {
		t.Errorf("expected no broken links, got %+v", broken)
	}
}

package parser

import (
	"strings"
	"testing"
)

func TestParseHTML_ExtractsAbsoluteLinks(t *testing.T) {
	body := strings.NewReader(`
		<html><body>
			<a href="https://example.com/page">absolute</a>
			<a href="http://other.com">other</a>
		</body></html>
	`)

	links, err := ParseHTML(body, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %v", len(links), links)
	}
}

func TestParseHTML_ResolvesRelativeLinks(t *testing.T) {
	body := strings.NewReader(`
		<html><body>
			<a href="/about">root-relative</a>
			<a href="page.html">relative</a>
		</body></html>
	`)

	links, err := ParseHTML(body, "https://example.com/blog/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %v", len(links), links)
	}

	expected := map[string]bool{
		"https://example.com/about":       true,
		"https://example.com/blog/page.html": true,
	}
	for _, l := range links {
		if !expected[l] {
			t.Errorf("unexpected link: %s", l)
		}
	}
}

func TestParseHTML_IgnoresUnsupportedSchemes(t *testing.T) {
	body := strings.NewReader(`
		<html><body>
			<a href="mailto:foo@bar.com">email</a>
			<a href="tel:+123">phone</a>
			<a href="javascript:void(0)">js</a>
			<a href="#anchor">fragment</a>
			<a href="">empty</a>
		</body></html>
	`)

	links, err := ParseHTML(body, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(links) != 0 {
		t.Errorf("expected no links, got %v", links)
	}
}

func TestParseHTML_NoLinks(t *testing.T) {
	body := strings.NewReader(`<html><body><p>no links here</p></body></html>`)

	links, err := ParseHTML(body, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(links) != 0 {
		t.Errorf("expected no links, got %v", links)
	}
}

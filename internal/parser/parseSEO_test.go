package parser

import (
	"strings"
	"testing"
)

func TestParseSEO_AllElementsPresent(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Example Test</title>
				<meta name="description" content="page description">
			</head>
			<body>
				<h1>Main Heading</h1>
			</body>
		</html>
	`

	seo, err := ParseSEO(strings.NewReader(html))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !seo.HasTitle {
		t.Error("expected HasTitle true")
	}
	if seo.Title != "Example Test" {
		t.Errorf("expected title 'Example Test', got %q", seo.Title)
	}
	if !seo.HasDescription {
		t.Error("expected HasDescription true")
	}
	if seo.Description != "page description" {
		t.Errorf("expected description 'page description', got %q", seo.Description)
	}
	if !seo.HasH1 {
		t.Error("expected HasH1 true")
	}
}

func TestParseSEO_NoElements(t *testing.T) {
	html := `<html><body><p>plain text</p></body></html>`

	seo, err := ParseSEO(strings.NewReader(html))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if seo.HasTitle {
		t.Error("expected HasTitle false")
	}
	if seo.Title != "" {
		t.Errorf("expected empty title, got %q", seo.Title)
	}
	if seo.HasDescription {
		t.Error("expected HasDescription false")
	}
	if seo.Description != "" {
		t.Errorf("expected empty description, got %q", seo.Description)
	}
	if seo.HasH1 {
		t.Error("expected HasH1 false")
	}
}

func TestParseSEO_DecodesHTMLEntities(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Tom &amp; Jerry</title>
				<meta name="description" content="Bread &amp; butter">
			</head>
			<body>
				<h1>A &lt; B</h1>
			</body>
		</html>
	`

	seo, err := ParseSEO(strings.NewReader(html))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if seo.Title != "Tom & Jerry" {
		t.Errorf("expected decoded title 'Tom & Jerry', got %q", seo.Title)
	}
	if seo.Description != "Bread & butter" {
		t.Errorf("expected decoded description 'Bread & butter', got %q", seo.Description)
	}
}

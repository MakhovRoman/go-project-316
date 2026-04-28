package crawler

import (
	"code/internal/linkchecker"
	"code/internal/parser"
	"code/internal/shared"
	"net/http"
	"time"
)

// Options описывает входные параметры обхода: стартовый URL, глубину,
// настройки сети (повторы, RPS, задержка, таймаут) и формат вывода.
type Options struct {
	URL         string
	Depth       uint
	Retries     uint
	RPS         uint
	Delay       time.Duration
	Timeout     time.Duration
	UserAgent   string
	Concurrency uint
	IndentJSON  bool
	HTTPClient  *http.Client
}

// Page — отчёт по одной странице сайта: HTTP-статус, найденные битые ссылки,
// ассеты и SEO-метаданные. Сериализуется в JSON.
type Page struct {
	URL          string                   `json:"url"`
	Depth        uint                     `json:"depth"`
	HTTPStatus   int                      `json:"http_status"`
	Status       string                   `json:"status"`
	Error        string                   `json:"error,omitempty"`
	SEO          parser.SEO               `json:"seo"`
	BrokenLinks  []linkchecker.BrokenLink `json:"broken_links"`
	Assets       []shared.Asset           `json:"assets"`
	DiscoveredAt string                   `json:"discovered_at"`
}

// Report — корневой объект JSON-отчёта: стартовый URL, заданная глубина обхода,
// время генерации и список обработанных страниц.
type Report struct {
	RootURL     string `json:"root_url"`
	Depth       uint   `json:"depth"`
	GeneratedAt string `json:"generated_at"`
	Pages       []Page `json:"pages"`
}

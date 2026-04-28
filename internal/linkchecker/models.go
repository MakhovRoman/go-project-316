package linkchecker

// BrokenLink описывает одну битую ссылку: абсолютный URL и причину —
// HTTP-статус (4xx/5xx) либо текст сетевой ошибки.
type BrokenLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

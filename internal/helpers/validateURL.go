package helpers

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateURL парсит rawURL и возвращает его каноническое строковое представление.
// Поддерживаются только схемы http и https; в остальных случаях возвращается ошибка.
func ValidateURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	return u.String(), nil
}

// NormalizeURL приводит URL к каноническому виду для дедупликации:
// убирает фрагмент (#...) и завершающий слэш в пути. Query-параметры сохраняются.
// При ошибке парсинга возвращает исходную строку как есть.
func NormalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.Fragment = ""
	u.Path = strings.TrimSuffix(u.Path, "/")
	return u.String()
}

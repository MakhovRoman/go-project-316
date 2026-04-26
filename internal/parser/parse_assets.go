package parser

import (
	"code/internal/helpers"
	"code/internal/shared"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const (
	AssetTypeImage  shared.AssetType = "image"
	AssetTypeScript shared.AssetType = "script"
	AssetTypeStyle  shared.AssetType = "style"
)

var assetTypeByTag = map[string]shared.AssetType{
	"img":    AssetTypeImage,
	"link":   AssetTypeStyle,
	"script": AssetTypeScript,
}

func ParseAssets(body io.Reader, path string) ([]shared.Asset, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	var assets []shared.Asset

	for n := range doc.Descendants() {
		if _, ok := assetTypeByTag[n.Data]; !ok {
			continue
		}

		asset := newAsset(n, baseURL)

		if asset.Type == "" || asset.URL == "" {
			continue
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

func newAsset(node *html.Node, baseURL *url.URL) shared.Asset {
	t := assetTypeByTag[node.Data]
	if node.Data == "link" {
		return handleHrefAsset(node, baseURL, t)
	}
	return handleSrcAsset(node, baseURL, t)
}

func getSafeURL(path string, baseURL *url.URL) (string, error) {
	ref, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	resolved := baseURL.ResolveReference(ref)
	safeURL, err := helpers.ValidateURL(resolved.String())
	if err != nil {
		return "", err
	}

	return safeURL, nil
}

func handleSrcAsset(node *html.Node, baseURL *url.URL, t shared.AssetType) shared.Asset {
	var asset shared.Asset

	for _, attr := range node.Attr {
		if attr.Key == "src" && attr.Val != "" {
			safeURL, err := getSafeURL(attr.Val, baseURL)
			if err != nil {
				continue
			}

			asset.Type = t
			asset.URL = safeURL
			break
		}
	}

	return asset
}

func handleHrefAsset(node *html.Node, baseURL *url.URL, t shared.AssetType) shared.Asset {
	var asset shared.Asset
	var assetType shared.AssetType
	var href string

	for _, attr := range node.Attr {
		if attr.Key == "rel" && strings.Contains(attr.Val, "stylesheet") {
			assetType = t
		}
		if attr.Key == "href" && attr.Val != "" {
			safeURL, err := getSafeURL(attr.Val, baseURL)
			if err != nil {
				continue
			}

			href = safeURL
		}

		if assetType != "" && href != "" {
			break
		}
	}

	if assetType != "" && href != "" {
		asset.URL = href
		asset.Type = t
	}

	return asset
}

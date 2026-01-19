package services

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ResolveZipLinks(baseURL string, rawHTML string) (string, error) {
	if strings.TrimSpace(rawHTML) == "" {
		return "", errors.New("html is empty")
	}
	if baseURL == "" {
		return "", errors.New("base url is empty")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base url: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for i, attr := range node.Attr {
				if strings.EqualFold(attr.Key, "href") && strings.Contains(strings.ToLower(attr.Val), ".zip") {
					parsed, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}
					if parsed.IsAbs() {
						continue
					}
					node.Attr[i].Val = base.ResolveReference(parsed).String()
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	var builder strings.Builder
	if err := html.Render(&builder, doc); err != nil {
		return "", fmt.Errorf("render html: %w", err)
	}

	return builder.String(), nil
}

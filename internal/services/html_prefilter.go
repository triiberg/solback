package services

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func ExtractZipTables(rawHTML string) ([]string, error) {
	if strings.TrimSpace(rawHTML) == "" {
		return []string{}, nil
	}

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var tables []*html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "table" {
			if tableHasZipLink(node) {
				tables = append(tables, node)
				return
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	results := make([]string, 0, len(tables))
	for _, table := range tables {
		var builder strings.Builder
		if err := html.Render(&builder, table); err != nil {
			return nil, fmt.Errorf("render table: %w", err)
		}
		results = append(results, builder.String())
	}

	return results, nil
}

func tableHasZipLink(table *html.Node) bool {
	if table == nil {
		return false
	}

	var found bool
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found {
			return
		}
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if strings.EqualFold(attr.Key, "href") && strings.Contains(strings.ToLower(attr.Val), ".zip") {
					found = true
					return
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
			if found {
				return
			}
		}
	}
	walk(table)

	return found
}
